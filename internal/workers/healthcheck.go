package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/lucashmsilva/rinha-2025-api-go/internal/infra/database"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/service"
)

const (
	HealthCheckInterval = 5500 * time.Millisecond
)

type HealthChecker struct {
	procService *service.ProcessorService
	db          *database.Db

	defaultIsFailing     atomic.Bool
	defaultMinRespTime   atomic.Int64
	defaultFallingCycles atomic.Int64

	fallbackIsFailing     atomic.Bool
	fallbackMinRespTime   atomic.Int64
	fallbackFallingCycles atomic.Int64
}

type Health struct {
	Failing         bool  `json:"failing"`
	MinResponseTime int64 `json:"minResponseTime"`
	FallingCycles   int64
}

func (h *Health) String() string {
	return fmt.Sprintf("Failing: %v| MinResponseTime: %v| FallingCycles: %v", h.Failing, h.MinResponseTime, h.FallingCycles)
}

func NewHealthChecker(db *database.Db, procService *service.ProcessorService) *HealthChecker {

	return &HealthChecker{
		procService: procService,
		db:          db,
	}
}

func (h *HealthChecker) StartHealthChecker() {
	ticker := time.NewTicker(HealthCheckInterval)
	go func() {
		for range ticker.C {
			defaultHealth, fallbackHealth := h.checkProcessors()
			h.updateHealthStore(service.ProcessorDefault, defaultHealth)
			h.updateHealthStore(service.ProcessorFallback, fallbackHealth)

			slog.Info("default processor health", "health", defaultHealth)
			slog.Info("fallback processor health", "health", fallbackHealth)
		}
	}()
}

func (h *HealthChecker) GetProcessorsHealth() (*Health, *Health) {
	return h.buildDefaultProcessorHealth(), h.buildFallbackProcessorHealth()
}

func (h *HealthChecker) checkProcessors() (*Health, *Health) {
	defaultProcessorHealth, err := h.callHealthCheckEndpoint(service.ProcessorDefault)
	if err != nil {
		return nil, nil
	}

	h.defaultIsFailing.Store(defaultProcessorHealth.Failing)
	h.defaultMinRespTime.Store(defaultProcessorHealth.MinResponseTime)

	if defaultProcessorHealth.Failing {
		h.defaultFallingCycles.Add(1)
	} else {
		h.defaultFallingCycles.Store(0)
	}

	fallbackProcessorHealth, err := h.callHealthCheckEndpoint(service.ProcessorFallback)
	if err != nil {
		return nil, nil
	}

	h.fallbackIsFailing.Store(fallbackProcessorHealth.Failing)
	h.fallbackMinRespTime.Store(fallbackProcessorHealth.MinResponseTime)

	if fallbackProcessorHealth.Failing {
		h.fallbackFallingCycles.Add(1)
	} else {
		h.fallbackFallingCycles.Store(0)
	}

	return h.buildDefaultProcessorHealth(), h.buildFallbackProcessorHealth()
}

func (h *HealthChecker) callHealthCheckEndpoint(processor string) (*Health, error) {
	var healthCheckRes Health

	resBody, resStatus, err := h.procService.MakeRequest(processor, http.MethodGet, "/payments/service-health", nil, 0)
	if err != nil || resStatus > 399 {
		slog.Error("Error health checking processor", "err", err, "resStatus", resStatus)
		return nil, err
	}

	json.NewDecoder(resBody).Decode(&healthCheckRes)

	return &healthCheckRes, nil
}

func (h *HealthChecker) buildDefaultProcessorHealth() *Health {
	return &Health{
		Failing:         h.defaultIsFailing.Load(),
		MinResponseTime: h.defaultMinRespTime.Load(),
		FallingCycles:   h.defaultFallingCycles.Load(),
	}
}

func (h *HealthChecker) buildFallbackProcessorHealth() *Health {
	return &Health{
		Failing:         h.fallbackIsFailing.Load(),
		MinResponseTime: h.fallbackMinRespTime.Load(),
		FallingCycles:   h.fallbackFallingCycles.Load(),
	}
}

func (h *HealthChecker) updateHealthStore(processor string, health *Health) {
	if health == nil {
		return
	}

	_, err := h.db.Conn.Exec(context.TODO(), `
		UPDATE processor_health
		SET is_falling = $1, min_response_time = $2, falling_cycles = $3
		WHERE processor = $4
	`,
		health.Failing, health.MinResponseTime, health.FallingCycles, processor,
	)
	if err != nil {
		slog.Error("Error updating health in store:", "err", err)
		return
	}
}

func (h *HealthChecker) getHealthFromStore() (*Health, *Health) {
	var defaultHealth, fallbackHealth Health

	res, err := h.db.Conn.Query(context.TODO(), "SELECT * FROM processor_health LIMIT 2")
	if err != nil {
		slog.Error("Error getting health from store:", "err", err)
		return nil, nil
	}

	var processor string
	var failing bool
	var minResponseTime int64
	var fallingCycles int64

	for res.Next() {
		err = res.Scan(&processor, &failing, &minResponseTime, &fallingCycles)
		if err != nil {
			slog.Error(fmt.Sprintf("Error scanning data: %v", err))
			return nil, nil
		}

		if processor == service.ProcessorDefault {
			defaultHealth = Health{failing, minResponseTime, fallingCycles}
		} else {
			fallbackHealth = Health{failing, minResponseTime, fallingCycles}
		}
	}

	return &defaultHealth, &fallbackHealth
}
