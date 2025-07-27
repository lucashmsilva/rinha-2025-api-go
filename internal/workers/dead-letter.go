package workers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lucashmsilva/rinha-2025-api-go/internal/entities"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/repositories"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/service"
)

const (
	queueMaxSize          = 50000
	maxDefaultProcRetries = 5
	maxTolerableRespTime  = 1500 * time.Millisecond
	maxAgeInQueue         = 1500 * time.Millisecond

	FailureProcError   = "processor_error"
	FailureProcTimeout = "processor_timeout"
)

type DLQ struct {
	healthChecker *HealthChecker
	procService   *service.ProcessorService
	paymentRep    *repositories.PaymentRepository

	queue chan *entities.PaymentRetry
}

func NewDQL(healthChecker *HealthChecker, procService *service.ProcessorService, paymentRep *repositories.PaymentRepository) *DLQ {
	return &DLQ{
		healthChecker: healthChecker,
		procService:   procService,
		paymentRep:    paymentRep,
		queue:         make(chan *entities.PaymentRetry),
	}
}

func (dlq *DLQ) PushToQueue(pr *entities.PaymentRetry) {
	go func() {
		dlq.queue <- pr
	}()
}

func (dlq *DLQ) StartDQLWorker() {
	go func() {
		for paymentRetry := range dlq.queue {
			ignoreSleep := dlq.retry(paymentRetry)

			if !ignoreSleep {
				time.Sleep(HealthCheckInterval)
			}
		}
	}()
}

// TODO: Before calling the processor, check if fail reason is timeout, if it is, call processor to check if payment was delivered after the timeout (read the processor code)
func (dlq *DLQ) retry(pr *entities.PaymentRetry) bool {

	defaultHealth, fallbackHealth := dlq.healthChecker.GetProcessorsHealth()
	if defaultHealth == nil || fallbackHealth == nil {
		return false
	}

	ignoreSleep := true

	// default is healthy. Retry with default
	if !defaultHealth.Failing /* && defaultHealth.MinResponseTime <= int64(maxTolerableRespTime) */ {
		rep, err := dlq.paymentRep.Start()
		if err != nil {
			dlq.PushToQueue(pr)
			return true
		}

		_, err = rep.Create(pr.P, service.ProcessorDefault)
		if err != nil {
			dlq.PushToQueue(pr)
			return true
		}

		paymentJSONBytes, _ := json.Marshal(pr.P)
		_, resStatus, err := dlq.procService.MakeRequestDefault(http.MethodPost, "/payments", bytes.NewReader(paymentJSONBytes), 7000)
		// request failed. Push back to queue
		if err != nil || resStatus > 399 {
			pr.FailureCount++
			pr.LastProcessorUsed = service.ProcessorDefault
			pr.LastFailureReason = FailureProcError

			if resStatus >= 399 {
				pr.LastFailureReason = FailureProcTimeout
			}

			dlq.PushToQueue(pr)
			ignoreSleep = false
			rep.Cancel()
		}

		rep.Finish()
		return ignoreSleep
	}

	// default is unhealthy and payment is not old enough to be sent to fallback. Push back to queue
	if pr.P.RequestedAt.UnixMilli()-time.Now().UnixMilli() < maxAgeInQueue.Milliseconds() {
		dlq.PushToQueue(pr)
		ignoreSleep = false

		return ignoreSleep
	}

	// default is unhealthy and payment is not already old enough to be sent to fallback and fallback is healthy. Retry with fallback
	if !fallbackHealth.Failing /*&& fallbackHealth.MinResponseTime <= int64(maxTolerableRespTime)*/ {
		rep, err := dlq.paymentRep.Start()
		if err != nil {
			dlq.PushToQueue(pr)
			return true
		}

		_, err = rep.Create(pr.P, service.ProcessorFallback)
		if err != nil {
			dlq.PushToQueue(pr)
			return true
		}

		paymentJSONBytes, _ := json.Marshal(pr.P)
		_, resStatus, err := dlq.procService.MakeRequestFallback(http.MethodPost, "/payments", bytes.NewReader(paymentJSONBytes), 7000)
		if err != nil {
			pr.FailureCount++
			pr.LastProcessorUsed = service.ProcessorFallback
			pr.LastFailureReason = FailureProcError

			if resStatus >= 399 {
				pr.LastFailureReason = FailureProcTimeout
			}

			dlq.PushToQueue(pr)
			ignoreSleep = false
			rep.Cancel()
		}

		rep.Finish()
		return ignoreSleep
	}

	// both fallback and default are unhealthy. Pushing back to queue
	dlq.PushToQueue(pr)
	ignoreSleep = false

	return ignoreSleep
}
