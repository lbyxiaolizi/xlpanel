package billing

import (
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/openhost/openhost/internal/core/domain"
)

type CouponType string

const (
	CouponPercentage   CouponType = "percentage"
	CouponFixedAmount  CouponType = "fixed_amount"
	CouponOverridePrice CouponType = "override_price"
)

var (
	ErrCouponInactive     = errors.New("coupon is inactive")
	ErrCouponNotFound     = errors.New("coupon not found")
	ErrCouponInvalid      = errors.New("coupon is invalid")
	ErrCouponUsageExceeded = errors.New("coupon usage exceeded")
)

type Coupon struct {
	Code        string
	Type        CouponType
	Amount      decimal.Decimal
	MaxCycles   int
	StartsAt    *time.Time
	EndsAt      *time.Time
	Active      bool
	Description string
}

type CartItem struct {
	Description string
	UnitPrice   decimal.Decimal
	Quantity    int64
	Recurring   bool
	Discount    decimal.Decimal
	Total       decimal.Decimal
}

type CouponResolver func(code string) (Coupon, error)

func ApplyCoupon(cartItem CartItem, couponCode string, resolver CouponResolver, service *domain.Service, now time.Time) (CartItem, error) {
	code := strings.TrimSpace(couponCode)
	if code == "" {
		return cartItem, ErrCouponNotFound
	}
	if resolver == nil {
		return cartItem, ErrCouponNotFound
	}
	coupon, err := resolver(code)
	if err != nil {
		return cartItem, err
	}
	return applyCoupon(cartItem, coupon, service, now)
}

func applyCoupon(cartItem CartItem, coupon Coupon, service *domain.Service, now time.Time) (CartItem, error) {
	if !coupon.Active {
		return cartItem, ErrCouponInactive
	}
	if coupon.Amount.IsNegative() {
		return cartItem, ErrCouponInvalid
	}
	if coupon.StartsAt != nil && now.Before(*coupon.StartsAt) {
		return cartItem, ErrCouponInactive
	}
	if coupon.EndsAt != nil && now.After(*coupon.EndsAt) {
		return cartItem, ErrCouponInactive
	}
	if coupon.MaxCycles > 0 && cartItem.Recurring {
		if service == nil {
			return cartItem, ErrCouponUsageExceeded
		}
		if service.TimesUsed >= coupon.MaxCycles {
			return cartItem, ErrCouponUsageExceeded
		}
	}

	quantity := cartItem.Quantity
	if quantity <= 0 {
		quantity = 1
	}
	subtotal := cartItem.UnitPrice.Mul(decimal.NewFromInt(quantity))
	discount := decimal.Zero

	switch coupon.Type {
	case CouponPercentage:
		discount = subtotal.Mul(coupon.Amount).Div(decimal.NewFromInt(100))
	case CouponFixedAmount:
		discount = coupon.Amount
	case CouponOverridePrice:
		overrideTotal := coupon.Amount.Mul(decimal.NewFromInt(quantity))
		discount = subtotal.Sub(overrideTotal)
	default:
		return cartItem, ErrCouponInvalid
	}

	if discount.IsNegative() {
		discount = decimal.Zero
	}
	if discount.GreaterThan(subtotal) {
		discount = subtotal
	}

	total := subtotal.Sub(discount)
	cartItem.Discount = discount
	cartItem.Total = total

	return cartItem, nil
}
