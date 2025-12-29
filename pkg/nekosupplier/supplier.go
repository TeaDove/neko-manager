package nekosupplier

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Supplier struct {
	client *http.Client
}

func New(client *http.Client) *Supplier {
	return &Supplier{client: client}
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

func (r *Stats) LastUsageAt() time.Time {
	if r.TotalUsers != 0 || r.TotalAdmins != 0 {
		return time.Time{}
	}

	lastUsage := r.ServerStartedAt

	if r.LastUserLeftAt.After(lastUsage) {
		lastUsage = r.LastUserLeftAt
	}

	if r.LastAdminLeftAt.After(lastUsage) {
		lastUsage = r.LastAdminLeftAt
	}

	return lastUsage
}

func (r *Supplier) doRequest(ctx context.Context, ip string, sessionAPIToken string, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s%s", ip, path), nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req.Header.Add("Authorization", "Bearer "+sessionAPIToken)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "client do")
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (r *Supplier) GetStats(ctx context.Context, ip string, sessionAPIToken string) (Stats, error) {
	body, err := r.doRequest(ctx, ip, sessionAPIToken, "/api/stats")
	if err != nil {
		return Stats{}, errors.Wrap(err, "/api/stats")
	}

	var stats Stats

	err = json.Unmarshal(body, &stats)
	if err != nil {
		return Stats{}, errors.Wrap(err, "decode response")
	}

	if stats.ServerStartedAt.IsZero() {
		return Stats{}, errors.New("bad server started")
	}

	return stats, nil
}

func (r *Supplier) GetScreenshot(ctx context.Context, ip string, sessionAPIToken string) ([]byte, error) {
	body, err := r.doRequest(ctx, ip, sessionAPIToken, "/api/room/screen/shot.jpg")
	if err != nil {
		return nil, errors.Wrap(err, "/api/room/screen/shot.jpg")
	}

	if len(body) == 0 {
		return nil, errors.New("empty screenshot")
	}

	return body, nil
}
