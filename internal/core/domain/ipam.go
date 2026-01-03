package domain

import "time"

type IPStatus string

const (
	IPStatusAvailable IPStatus = "available"
	IPStatusAllocated IPStatus = "allocated"
	IPStatusReserved  IPStatus = "reserved"
)

type Subnet struct {
	ID        uint64      `gorm:"primaryKey"`
	CIDR      string      `gorm:"size:64;not null;uniqueIndex"`
	Gateway   string      `gorm:"size:64"`
	Netmask   string      `gorm:"size:64"`
	IPs       []IPAddress `gorm:"foreignKey:SubnetID"`
	CreatedAt time.Time   `gorm:"not null"`
	UpdatedAt time.Time   `gorm:"not null"`
}

type IPAddress struct {
	ID        uint64    `gorm:"primaryKey"`
	SubnetID  uint64    `gorm:"not null;index;uniqueIndex:idx_subnet_ip"`
	IP        string    `gorm:"size:64;not null;uniqueIndex:idx_subnet_ip"`
	Gateway   string    `gorm:"size:64"`
	Netmask   string    `gorm:"size:64"`
	Status    IPStatus  `gorm:"size:32;not null;default:'available'"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}
