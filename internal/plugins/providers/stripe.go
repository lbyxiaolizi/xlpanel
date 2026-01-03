package providers

import (
	"context"

	"xlpanel/internal/plugins"
)

type Stripe struct {
	config map[string]string
}

func NewStripe() *Stripe {
	return &Stripe{config: map[string]string{}}
}

func (p *Stripe) Name() string {
	return "Stripe"
}

func (p *Stripe) Provider() string {
	return "stripe"
}

func (p *Stripe) Initialize(config map[string]string) error {
	p.config = config
	return nil
}

func (p *Stripe) Charge(ctx context.Context, request plugins.PaymentRequest) (plugins.PaymentResponse, error) {
	return plugins.PaymentResponse{
		TransactionID: "stripe-demo",
		Status:        "pending",
		Message:       "stripe payment intent created",
	}, nil
}
