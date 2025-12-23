package nekoproxy

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"

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
	dialer := net.Dialer{Timeout: 3 * time.Minute, KeepAlive: 3 * time.Minute}

	reverseProxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			rewriteRequestURL(req, r.target.Load())
		},
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&dialer).DialContext,
			ForceAttemptHTTP2:     false, // Because safari for some reason works bad with HTTP2 and web-sockets
			MaxIdleConns:          200,
			IdleConnTimeout:       10 * time.Minute,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	zerolog.Ctx(ctx).
		Info().
		Msg("making.reverse.proxy")

	return &reverseProxy
}
