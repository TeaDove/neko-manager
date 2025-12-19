package instancerepo

import (
	"fmt"
	"net"
	"time"
)

type Instance struct {
	ID        string         `gorm:"primaryKey"`
	CreatedAt time.Time      `gorm:"not null;autoCreateTime;not null"`
	CreatedBy string         `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null;autoUpdateTime"`
	Status    InstanceStatus `gorm:"not null;type:string"`

	IP net.IP `gorm:"type:string;uniqueIndex"`
}

func (r *Instance) CloudName() string {
	return fmt.Sprintf("neko-%s", r.ID)
}

// InstanceStatus
// ENUM(Creating, Running, Deleted)
//
//go:generate go tool go-enum --sql --marshal -f models.go
type InstanceStatus int
