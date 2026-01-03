package providers

import (
	"context"

	"xlpanel/internal/plugins"
)

type Crypto struct {
	config map[string]string
}

func NewCrypto() *Crypto {
	return &Crypto{config: map[string]string{}}
}

func (p *Crypto) Name() string {
	return "Crypto"
}

func (p *Crypto) Provider() string {
	return "crypto"
}

func (p *Crypto) Initialize(config map[string]string) error {
	p.config = config
	return nil
}

func (p *Crypto) Charge(ctx context.Context, request plugins.PaymentRequest) (plugins.PaymentResponse, error) {
	return plugins.PaymentResponse{
		TransactionID: "crypto-demo",
		Status:        "pending",
		Message:       "crypto address generated",
	}, nil
}
