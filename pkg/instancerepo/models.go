package instancerepo

import (
	"fmt"
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
	IP              string `gorm:"uniqueIndex"`
	CloudInstanceID string
}

func (r *Instance) MarshalZerologObject(e *zerolog.Event) {
	e.Str("id", r.ID).
		Str("cloud_instance_id", r.CloudInstanceID).
		Str("ip", r.IP).
		Stringer("status", r.Status)
}

const NEKO = "neko"

func (r *Instance) CloudName() string {
	return fmt.Sprintf("%s-%s", NEKO, r.ID)
}

func (r *Instance) UserLoginURL() string {
	return "http://%s?pwd=neko"
}

func (r *Instance) AdminLoginURL() string {
	return "http://%s?pwd=admin"
}

// InstanceStatus
// ENUM(Creating, Started, Running, Deleting, Deleted)
//
//go:generate go tool go-enum --sql --marshal -f models.go
type InstanceStatus int //nolint: recvcheck // codegen :(
