package instancerepo

import (
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog"
)

type Instance struct {
	ID        string    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime;not null"`

	CreatedBy string `gorm:"not null"`
	TGChatID  int64  `gorm:"not null"`

	UpdatedAt time.Time      `gorm:"not null;autoUpdateTime"`
	Status    InstanceStatus `gorm:"not null;type:string;index"`

	SessionAPIToken string `gorm:"not null"`
	IP              net.IP `gorm:"type:string;uniqueIndex"`
	CloudInstanceID string
}

func (r *Instance) MarshalZerologObject(e *zerolog.Event) {
	e.Str("id", r.ID).
		Str("cloud_instance_id", r.CloudInstanceID).
		Stringer("ip", r.IP).
		Stringer("status", r.Status)
}

func (r *Instance) CloudName() string {
	return fmt.Sprintf("neko-%s", r.ID)
}

// InstanceStatus
// ENUM(Creating, Started, Running, Deleting, Deleted)
//
//go:generate go tool go-enum --sql --marshal -f models.go
type InstanceStatus int //nolint: recvcheck // codegen :(
