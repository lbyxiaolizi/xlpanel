package ipam

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openhost/openhost/internal/core/domain"
)

var (
	ErrNoAvailableIP = errors.New("no available ip addresses")
)

func AllocateIP(db *gorm.DB, subnetID uint64) (domain.IPAddress, error) {
	if db == nil {
		return domain.IPAddress{}, fmt.Errorf("db is required")
	}

	var allocated domain.IPAddress
	err := db.Transaction(func(tx *gorm.DB) error {
		var subnet domain.Subnet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&subnet, subnetID).Error; err != nil {
			return fmt.Errorf("lock subnet: %w", err)
		}

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("subnet_id = ? AND status = ?", subnetID, domain.IPStatusAvailable).
			Order("id").
			First(&allocated).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNoAvailableIP
			}
			return fmt.Errorf("find available ip: %w", err)
		}

		if err := tx.Model(&allocated).
			Update("status", domain.IPStatusAllocated).Error; err != nil {
			return fmt.Errorf("update ip status: %w", err)
		}

		return nil
	})

	if err != nil {
		return domain.IPAddress{}, err
	}

	return allocated, nil
}
