package instancerepo

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"neko-manager/pkg/nekosupplier"
	"time"

	"github.com/pkg/errors"

	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/time_utils"
)

type Instance struct {
	ID        string    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime;not null"`

	LastHealthOk *time.Time
	UpdatedAt    time.Time      `gorm:"not null;autoUpdateTime"`
	Status       InstanceStatus `gorm:"not null;type:string;index"`

	CreatedBy      string `gorm:"not null"`
	TGChatID       int64  `gorm:"not null"`
	TGThreadChatID *int

	SessionAPIToken string `gorm:"not null"`
	IP              *string
	CloudFolderID   string `gorm:"not null"`
	ProxyURL        *string
	CloudInstanceID *string
	ResourceSize    ResourcesSize `gorm:"not null;type:string"`
}

func (r *Instance) ToSupplierDTO() *nekosupplier.Instance {
	return &nekosupplier.Instance{
		ID:              r.ID,
		SessionAPIToken: r.SessionAPIToken,
		IP:              *r.IP,
	}
}

func (r *Instance) MarshalZerologObject(e *zerolog.Event) {
	e.Str("id", r.ID).
		Stringer("status", r.Status)

	if r.CloudInstanceID != nil {
		e.Str("cloud_instance_id", *r.CloudInstanceID)
	}

	if r.IP != nil {
		e.Str("ip", *r.IP)
	}
}

const NEKO = "neko"

func (r *Instance) CloudName() string {
	return fmt.Sprintf("%s-%s", NEKO, r.ID)
}

//go:embed instance.gohtml
var reprRaw string
var reprTemplate = template.Must(template.New("instance").Parse(reprRaw))

func (r *Instance) Repr(stats *nekosupplier.Stats) (string, error) {
	var buf bytes.Buffer

	err := reprTemplate.Execute(&buf, map[string]any{
		"Instance": r,
		"Stats":    stats,
		"Elapsed": func(v time.Time) string {
			if v.IsZero() {
				return "never"
			}

			return time_utils.RoundDuration(time.Since(v))
		},
	},
	)
	if err != nil {
		return "", errors.Wrap(err, "execute template")
	}

	return buf.String(), nil
}

// InstanceStatus
// ENUM(Creating, Started, Restarting, Running, Deleting, Deleted)
//
//go:generate go tool go-enum --sql --marshal --names -f models.go
type InstanceStatus int //nolint: recvcheck // codegen :(

func (s InstanceStatus) EmojiStatus() string {
	var emoji string

	switch s {
	case InstanceStatusRunning:
		emoji = "✅"
	case InstanceStatusDeleting, InstanceStatusDeleted:
		emoji = "❌"
	default:
		emoji = "⚠️"
	}

	return fmt.Sprintf("%s %s %s", emoji, s.String(), emoji)
}

// ResourcesSize
// ENUM(s, m, l).
type ResourcesSize int //nolint: recvcheck // codegen :(
