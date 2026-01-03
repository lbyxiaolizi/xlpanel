package providers

import (
	"context"

	"xlpanel/internal/plugins"
)

type AlipayFaceToFace struct {
	config map[string]string
}

func NewAlipayFaceToFace() *AlipayFaceToFace {
	return &AlipayFaceToFace{config: map[string]string{}}
}

func (p *AlipayFaceToFace) Name() string {
	return "Alipay Face-to-Face"
}

func (p *AlipayFaceToFace) Provider() string {
	return "alipay_f2f"
}

func (p *AlipayFaceToFace) Initialize(config map[string]string) error {
	p.config = config
	return nil
}

func (p *AlipayFaceToFace) Charge(ctx context.Context, request plugins.PaymentRequest) (plugins.PaymentResponse, error) {
	return plugins.PaymentResponse{
		TransactionID: "alipay-demo",
		Status:        "pending",
		Message:       "alipay f2f qr generated",
	}, nil
}
