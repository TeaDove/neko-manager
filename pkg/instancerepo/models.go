package instancerepo

import (
	"net"
	"time"
)

type Instance struct {
	ID        string         `gorm:"primaryKey"`
	CreatedAt time.Time      `gorm:"not null;autoCreateTime;not null"`
	UpdatedAt time.Time      `gorm:"not null;autoUpdateTime"`
	Status    InstanceStatus `gorm:"not null;type:string"`

	IP net.IP `gorm:"uniqueIndex"`
}

// InstanceStatus
// ENUM(CREATING, RUNNING, DELETED)
//
//go:generate go tool go-enum --sql --marshal -f models.go
type InstanceStatus int
