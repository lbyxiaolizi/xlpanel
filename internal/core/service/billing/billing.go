package billing

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrServiceNotRecurring = errors.New("service is not recurring")
	ErrServiceNotDue       = errors.New("service is not due for invoicing")
	ErrInvalidDueDate      = errors.New("standard due date must be between 1 and 31")
	ErrInvalidCost         = errors.New("cost must be non-negative")
	ErrInvalidTaxRate      = errors.New("tax rate must be non-negative")
)

type Service struct {
	ID             uint64
	Recurring      bool
	NextDueDate    time.Time
	StandardDueDay int
	Prorata        bool
	RecurringCost  decimal.Decimal
	TaxRate        decimal.Decimal
	Currency       string
}

type Invoice struct {
	ServiceID  uint64
	IssueDate  time.Time
	DueDate    time.Time
	Currency   string
	LineItems  []InvoiceLineItem
	Subtotal   decimal.Decimal
	TaxAmount  decimal.Decimal
	Total      decimal.Decimal
	Prorata    bool
	PeriodFrom time.Time
	PeriodTo   time.Time
}

type InvoiceLineItem struct {
	Description string
	Amount      decimal.Decimal
}

func CalculateProrata(startDate time.Time, standardDueDate int, cost decimal.Decimal) (decimal.Decimal, time.Time, error) {
	if standardDueDate < 1 || standardDueDate > 31 {
		return decimal.Zero, time.Time{}, ErrInvalidDueDate
	}
	if cost.IsNegative() {
		return decimal.Zero, time.Time{}, ErrInvalidCost
	}

	location := startDate.Location()
	year, month, day := startDate.Date()
	dueDate := normalizeDueDate(time.Date(year, month, 1, 0, 0, 0, 0, location), standardDueDate)
	if day > dueDate.Day() || (day == dueDate.Day() && startDate.After(dueDate)) {
		month = month + 1
		dueDate = normalizeDueDate(time.Date(year, month, 1, 0, 0, 0, 0, location), standardDueDate)
	}

	previousMonth := month - 1
	previousYear := year
	if month == time.January {
		previousMonth = time.December
		previousYear = year - 1
	}
	lastDueDate := normalizeDueDate(time.Date(previousYear, previousMonth, 1, 0, 0, 0, 0, location), standardDueDate)
	if dueDate.Equal(lastDueDate) || dueDate.Before(lastDueDate) {
		return decimal.Zero, time.Time{}, errors.New("invalid billing window")
	}

	periodSeconds := dueDate.Sub(lastDueDate).Seconds()
	remainingSeconds := dueDate.Sub(startDate).Seconds()
	if remainingSeconds < 0 {
		return decimal.Zero, time.Time{}, ErrServiceNotDue
	}
	if periodSeconds <= 0 {
		return decimal.Zero, time.Time{}, errors.New("invalid billing period")
	}

	ratio := decimal.NewFromFloat(remainingSeconds).Div(decimal.NewFromFloat(periodSeconds))
	amount := ratio.Mul(cost)
	return amount, dueDate, nil
}

func GenerateInvoice(service Service, now time.Time) (Invoice, error) {
	if !service.Recurring {
		return Invoice{}, ErrServiceNotRecurring
	}
	if service.NextDueDate.After(now) {
		return Invoice{}, ErrServiceNotDue
	}
	if service.RecurringCost.IsNegative() {
		return Invoice{}, ErrInvalidCost
	}
	if service.TaxRate.IsNegative() {
		return Invoice{}, ErrInvalidTaxRate
	}

	amount := service.RecurringCost
	periodFrom := service.NextDueDate
	periodTo := service.NextDueDate
	prorata := false

	if service.Prorata {
		var err error
		amount, periodTo, err = CalculateProrata(now, service.StandardDueDay, service.RecurringCost)
		if err != nil {
			return Invoice{}, err
		}
		prorata = true
		periodFrom = now
	}

	taxAmount := amount.Mul(service.TaxRate).Div(decimal.NewFromInt(100))
	total := amount.Add(taxAmount)

	invoice := Invoice{
		ServiceID: service.ID,
		IssueDate: now,
		DueDate:   service.NextDueDate,
		Currency:  service.Currency,
		LineItems: []InvoiceLineItem{
			{
				Description: "Recurring service charge",
				Amount:      amount,
			},
		},
		Subtotal:   amount,
		TaxAmount:  taxAmount,
		Total:      total,
		Prorata:    prorata,
		PeriodFrom: periodFrom,
		PeriodTo:   periodTo,
	}

	return invoice, nil
}

func normalizeDueDate(monthStart time.Time, standardDueDate int) time.Time {
	lastDay := time.Date(monthStart.Year(), monthStart.Month()+1, 0, 0, 0, 0, 0, monthStart.Location()).Day()
	dueDay := standardDueDate
	if dueDay > lastDay {
		dueDay = lastDay
	}
	return time.Date(monthStart.Year(), monthStart.Month(), dueDay, 0, 0, 0, 0, monthStart.Location())
}
