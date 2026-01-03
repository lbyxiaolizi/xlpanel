package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type ProductGroup struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:255;not null"`
	Slug        string    `gorm:"size:255;uniqueIndex;not null"`
	Description string    `gorm:"type:text"`
	SortOrder   int       `gorm:"not null;default:0"`
	Active      bool      `gorm:"not null;default:true"`
	Products    []Product `gorm:"foreignKey:ProductGroupID"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

type Product struct {
	ID             uint64        `gorm:"primaryKey"`
	ProductGroupID uint64        `gorm:"not null;index"`
	Name           string        `gorm:"size:255;not null"`
	Slug           string        `gorm:"size:255;uniqueIndex;not null"`
	Description    string        `gorm:"type:text"`
	ModuleName     string        `gorm:"size:128;not null;index"`
	Active         bool          `gorm:"not null;default:true"`
	ConfigGroups   []ConfigGroup `gorm:"many2many:product_config_groups"`
	CreatedAt      time.Time     `gorm:"not null"`
	UpdatedAt      time.Time     `gorm:"not null"`
}

type ConfigGroup struct {
	ID          uint64         `gorm:"primaryKey"`
	Name        string         `gorm:"size:255;not null"`
	Description string         `gorm:"type:text"`
	Options     []ConfigOption `gorm:"foreignKey:ConfigGroupID"`
	Products    []Product      `gorm:"many2many:product_config_groups"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
}

type ProductConfigGroup struct {
	ProductID     uint64    `gorm:"primaryKey"`
	ConfigGroupID uint64    `gorm:"primaryKey"`
	CreatedAt     time.Time `gorm:"not null"`
}

type ConfigOption struct {
	ID            uint64            `gorm:"primaryKey"`
	ConfigGroupID uint64            `gorm:"not null;index"`
	Name          string            `gorm:"size:255;not null"`
	InputType     string            `gorm:"size:64;not null"`
	Required      bool              `gorm:"not null;default:false"`
	SortOrder     int               `gorm:"not null;default:0"`
	SubOptions    []ConfigSubOption `gorm:"foreignKey:ConfigOptionID"`
	CreatedAt     time.Time         `gorm:"not null"`
	UpdatedAt     time.Time         `gorm:"not null"`
}

type ConfigSubOption struct {
	ID             uint64    `gorm:"primaryKey"`
	ConfigOptionID uint64    `gorm:"not null;index"`
	Name           string    `gorm:"size:255;not null"`
	SortOrder      int       `gorm:"not null;default:0"`
	Pricing        Pricing   `gorm:"embedded"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

type Pricing struct {
	SetupFee    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Monthly     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Quarterly   decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Yearly      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Triennially decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
}
