package providers

import (
	"context"

	"xlpanel/internal/plugins"
)

type UnionPay struct {
	config map[string]string
}

func NewUnionPay() *UnionPay {
	return &UnionPay{config: map[string]string{}}
}

func (p *UnionPay) Name() string {
	return "UnionPay"
}

func (p *UnionPay) Provider() string {
	return "unionpay"
}

func (p *UnionPay) Initialize(config map[string]string) error {
	p.config = config
	return nil
}

func (p *UnionPay) Charge(ctx context.Context, request plugins.PaymentRequest) (plugins.PaymentResponse, error) {
	return plugins.PaymentResponse{
		TransactionID: "unionpay-demo",
		Status:        "pending",
		Message:       "unionpay transaction created",
	}, nil
}
