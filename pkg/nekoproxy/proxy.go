package nekoproxy

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/must_utils"
)

type Proxy struct {
	target atomic.Pointer[url.URL]
	URL    string
}

func New() *Proxy {
	proxy := &Proxy{}

	proxy.SetTarget(must_utils.Must(url.Parse("https://google.com")))

	return proxy
}

func (r *Proxy) SetTarget(target *url.URL) {
	r.target.Store(target)
}

func (r *Proxy) MakeSTDProxy(ctx context.Context) *httputil.ReverseProxy {
	reverseProxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			rewriteRequestURL(req, r.target.Load())
		},
	}

	zerolog.Ctx(ctx).
		Info().
		Msg("making.reverse.proxy")

	return &reverseProxy
}
