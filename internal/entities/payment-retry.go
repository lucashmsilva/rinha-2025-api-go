package entities

type PaymentRetry struct {
	P                 *Payment
	FailureCount      int
	LastProcessorUsed string
	LastFailureReason string
}
