package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Order struct {
	ID         uint64    `gorm:"primaryKey"`
	CustomerID uint64    `gorm:"not null;index"`
	ProductID  uint64    `gorm:"not null;index"`
	ServiceID  uint64    `gorm:"index"`
	Status     string    `gorm:"size:64;not null"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}

type Service struct {
	ID              uint64       `gorm:"primaryKey"`
	CustomerID      uint64       `gorm:"not null;index"`
	ProductID       uint64       `gorm:"not null;index"`
	Status          string       `gorm:"size:64;not null"`
	TimesUsed       int          `gorm:"not null;default:0"`
	IPAddressID     *uint64      `gorm:"index"`
	ConfigSelection JSONMap      `gorm:"type:jsonb;not null"`
	PluginConfig    PluginConfig `gorm:"type:jsonb;not null"`
	Product         Product      `gorm:"foreignKey:ProductID"`
	IPAddress       *IPAddress   `gorm:"foreignKey:IPAddressID"`
	CreatedAt       time.Time    `gorm:"not null"`
	UpdatedAt       time.Time    `gorm:"not null"`
}

type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	payload, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (m *JSONMap) Scan(value any) error {
	if m == nil {
		return errors.New("json map scan: nil receiver")
	}
	data, err := normalizeJSONBytes(value)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		*m = JSONMap{}
		return nil
	}
	return json.Unmarshal(data, m)
}

type PluginConfig struct {
	Region string            `json:"region,omitempty"`
	NodeID string            `json:"node_id,omitempty"`
	Values map[string]string `json:"values,omitempty"`
}

func (c PluginConfig) Value() (driver.Value, error) {
	payload, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *PluginConfig) Scan(value any) error {
	if c == nil {
		return errors.New("plugin config scan: nil receiver")
	}
	data, err := normalizeJSONBytes(value)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		*c = PluginConfig{}
		return nil
	}
	return json.Unmarshal(data, c)
}

func normalizeJSONBytes(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	switch data := value.(type) {
	case []byte:
		return data, nil
	case string:
		return []byte(data), nil
	default:
		return nil, errors.New("unsupported JSON value type")
	}
}
