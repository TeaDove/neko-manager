package nekosupplier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Supplier struct {
	statsPath string

	client *http.Client
}

func New(client *http.Client) *Supplier {
	return &Supplier{client: client, statsPath: "/api/stats"}
}

type Stats struct {
	HasHost         bool      `json:"has_host"`
	HostId          string    `json:"host_id,omitempty"`
	ServerStartedAt time.Time `json:"server_started_at"`
	TotalUsers      int       `json:"total_users"`
	LastUserLeftAt  time.Time `json:"last_user_left_at,omitempty"`
	TotalAdmins     int       `json:"total_admins"`
	LastAdminLeftAt time.Time `json:"last_admin_left_at,omitempty"`
}

func (r *Supplier) GetStats(ctx context.Context, ip string, sessionAPIToken string) (Stats, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s%s", ip, r.statsPath), nil)
	if err != nil {
		return Stats{}, errors.WithStack(err)
	}

	req.Header.Add("Authorization", "Bearer "+sessionAPIToken)

	resp, err := r.client.Do(req)
	if err != nil {
		return Stats{}, errors.Wrap(err, "client do")
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return Stats{}, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stats Stats

	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		return Stats{}, errors.Wrap(err, "decode response")
	}

	if stats.ServerStartedAt.IsZero() {
		return Stats{}, errors.New("bad server started")
	}

	return stats, nil
}
