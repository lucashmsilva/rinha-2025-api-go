package entities

import "fmt"

type Metrics struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentSummary struct {
	Default  Metrics `json:"default"`
	Fallback Metrics `json:"fallback"`
}

func (p PaymentSummary) String() string {
	return fmt.Sprintf("Default.TotalRequests: %v | Default.TotalAmount: %v | Fallback.TotalRequests: %v | Fallback.TotalAmount: %v", p.Default.TotalRequests, p.Default.TotalAmount, p.Fallback.TotalRequests, p.Fallback.TotalAmount)
}
