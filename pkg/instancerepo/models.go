package instancerepo

import (
	"fmt"
	"neko-manager/pkg/nekosupplier"
	"strings"
	"time"

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
	return fmt.Sprintf("http://%s?pwd=neko", r.IP)
}

func (r *Instance) AdminLoginURL() string {
	return fmt.Sprintf("http://%s?pwd=admin", r.IP)
}

func (r *Instance) cloudURL() string {
	return fmt.Sprintf(
		"https://console.yandex.cloud/folders/%s/compute/instance/%s/overview",
		r.CloudFolderID,
		r.CloudInstanceID,
	)
}

func (r *Instance) ssh() string {
	return fmt.Sprintf("ssh -oStrictHostKeyChecking=no -i ~/.ssh/id_rsa_yc -v neko@%s", r.IP)
}

func (r *Instance) Repr(stats *nekosupplier.Stats) string {
	var (
		builder strings.Builder
		now     = time.Now().UTC()
	)

	builder.WriteString(
		fmt.Sprintf(
			`🐈‍⬛ Neko instance &lt;<code>%s</code>&gt; created by @%s; <b>%s</b>
Alive for %s

`,
			r.ID,
			r.CreatedBy,
			r.Status.EmojiStatus(),
			time_utils.RoundDuration(now.Sub(r.CreatedAt)),
		),
	)

	if r.IP != "" {
		builder.WriteString(
			fmt.Sprintf(
				`User login: %s
Admin login: %s
IP: %s
SSH: <code>%s</code>

`,
				r.UserLoginURL(),
				r.AdminLoginURL(),
				r.IP,
				r.ssh(),
			),
		)
	}

	if r.CloudInstanceID != "" {
		builder.WriteString(fmt.Sprintf(`Cloud page: <a href="%s">yc</a>

`, r.cloudURL()))
	}

	if stats != nil {
		reprStats(stats, &builder, now)
	}

	return builder.String()
}

func reprStats(stats *nekosupplier.Stats, builder *strings.Builder, now time.Time) {
	builder.WriteString("Statistics: ")

	var usage bool
	if stats.HasHost {
		usage = true

		fmt.Fprintf(builder, "host = <code>%s</code> ", stats.HostId)
	}

	if stats.TotalUsers != 0 {
		usage = true

		fmt.Fprintf(builder, "total users = %d ", stats.TotalUsers)
	} else if !stats.LastUserLeftAt.IsZero() {
		usage = true

		fmt.Fprintf(builder, "last user left = %s ago ", time_utils.RoundDuration(now.Sub(stats.LastUserLeftAt)))
	}

	if stats.TotalAdmins != 0 {
		usage = true

		fmt.Fprintf(builder, "total admins = %d ", stats.TotalAdmins)
	} else if !stats.LastAdminLeftAt.IsZero() {
		usage = true

		fmt.Fprintf(builder,
			"last admin left = %s ago ",
			time_utils.RoundDuration(now.Sub(stats.LastAdminLeftAt)))
	}

	if !usage {
		fmt.Fprintf(builder,
			"no usage for = %s ", time_utils.RoundDuration(now.Sub(stats.ServerStartedAt)))
	}
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
