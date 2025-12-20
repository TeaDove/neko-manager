package nekosupplier

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type Supplier struct {
	metricsPath string

	client *http.Client
}

func New(client *http.Client) *Supplier {
	return &Supplier{client: client, metricsPath: "/metrics"}
}

func (r *Supplier) Ping(ctx context.Context, ip string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s%s", ip, r.metricsPath), nil)
	if err != nil {
		return errors.WithStack(err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to ping nekosupplier")
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
