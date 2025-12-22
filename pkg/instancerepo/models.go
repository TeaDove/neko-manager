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

	CreatedBy string `gorm:"not null"`
	TGChatID  int64  `gorm:"not null"`

	UpdatedAt time.Time      `gorm:"not null;autoUpdateTime"`
	Status    InstanceStatus `gorm:"not null;type:string;index"`

	SessionAPIToken string `gorm:"not null"`
	IP              string
	CloudFolderID   string `gorm:"not null"`
	ProxyURL        string
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

//go:embed instance.gohtml
var reprRaw string
var reprTemplate = template.Must(template.New("instance").Parse(reprRaw))

func (r *Instance) Repr(stats *nekosupplier.Stats) (string, error) {
	var buf bytes.Buffer

	err := reprTemplate.Execute(&buf, map[string]any{
		"Instance": r,
		"Stats":    stats,
		"Elapsed": func(v time.Time) string {
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
// ENUM(Creating, Started, Running, Deleting, Deleted)
//
//go:generate go tool go-enum --sql --marshal -f models.go
type InstanceStatus int //nolint: recvcheck // codegen :(

func (s InstanceStatus) EmojiStatus() string {
	var emoji string

	switch s {
	case InstanceStatusRunning:
		emoji = "✅"
	case InstanceStatusDeleting:
		emoji = "❌"
	default:
		emoji = "🪧"
	}

	return fmt.Sprintf("%s %s %s", emoji, s.String(), emoji)
}
