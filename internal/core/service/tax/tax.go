package tax

import (
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
)

type Calculator struct {
	db *gorm.DB
}

func NewCalculator(db *gorm.DB) *Calculator {
	return &Calculator{db: db}
}

func (c *Calculator) CalculateForCustomer(customerID uint64, amount decimal.Decimal) (decimal.Decimal, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, nil
	}

	var user domain.User
	if err := c.db.Select("id", "country", "state").First(&user, customerID).Error; err != nil {
		return decimal.Zero, err
	}

	return c.calculateForRegion(user.Country, user.State, amount)
}

func (c *Calculator) calculateForRegion(country, state string, amount decimal.Decimal) (decimal.Decimal, error) {
	country = strings.TrimSpace(strings.ToUpper(country))
	state = strings.TrimSpace(state)
	if country == "" {
		return decimal.Zero, nil
	}

	var rules []domain.TaxRule
	if err := c.db.Where("active = ? AND country = ? AND (state = ? OR state = '')", true, country, state).
		Order("priority DESC, id ASC").
		Find(&rules).Error; err != nil {
		return decimal.Zero, err
	}
	if len(rules) == 0 {
		return decimal.Zero, nil
	}

	totalRate := decimal.Zero
	inclusive := false
	for _, rule := range rules {
		totalRate = totalRate.Add(rule.Rate)
		if rule.IsInclusive {
			inclusive = true
		}
	}

	if totalRate.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, nil
	}

	if inclusive {
		rateFactor := totalRate.Div(decimal.NewFromInt(100))
		return amount.Sub(amount.Div(decimal.NewFromInt(1).Add(rateFactor))), nil
	}
	return amount.Mul(totalRate).Div(decimal.NewFromInt(100)), nil
}
