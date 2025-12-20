package nekosupplier

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/teadove/teasutils/utils/test_utils"
)

type Supplier struct {
	statsPath string

	client *http.Client
}

func New(client *http.Client) *Supplier {
	return &Supplier{client: client, statsPath: "/api/stats"}
}

func (r *Supplier) GetStats(ctx context.Context, ip string, sessionAPIToken string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s%s", ip, r.statsPath), nil)
	if err != nil {
		return errors.WithStack(err)
	}

	req.Header.Add("token", sessionAPIToken)

	resp, err := r.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "client do")
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	test_utils.Pprint(body)

	return nil
}
