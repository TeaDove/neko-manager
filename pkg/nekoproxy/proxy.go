package nekoproxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/must_utils"
)

type Proxy struct {
	targets   map[string]url.URL
	targetsMU sync.RWMutex

	notFound url.URL
	idLen    int

	baseURL string
}

func New(idLen int, baseURL string) *Proxy {
	proxy := &Proxy{
		notFound: *must_utils.Must(url.Parse("https://google.com")),
		targets:  make(map[string]url.URL),
		idLen:    idLen,
		baseURL:  baseURL,
	}

	return proxy
}

func (r *Proxy) GetProxyURL(id string) *string {
	if r.baseURL == "" {
		return nil
	}

	proxyURL := fmt.Sprintf("%s/%s", r.baseURL, id)

	return &proxyURL
}

func (r *Proxy) AddTarget(id string, target *url.URL) {
	r.targetsMU.Lock()
	defer r.targetsMU.Unlock()

	r.targets[id] = *target
}

func (r *Proxy) DeleteTarget(path string) {
	r.targetsMU.Lock()
	defer r.targetsMU.Unlock()

	delete(r.targets, path)
}

func (r *Proxy) getTarget(path string) *url.URL {
	if len(path) < r.idLen+1 {
		return &r.notFound
	}

	id := path[1 : r.idLen+1]

	r.targetsMU.RLock()
	defer r.targetsMU.RUnlock()

	target, ok := r.targets[id]
	if !ok {
		return &r.notFound
	}

	return &target
}

func (r *Proxy) MakeSTDProxy(ctx context.Context) *httputil.ReverseProxy {
	dialer := net.Dialer{Timeout: 3 * time.Minute, KeepAlive: 3 * time.Minute}

	reverseProxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			rewriteRequestURL(req, r.getTarget(req.URL.Path))
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
		Msg("reverse.proxy.created")

	return &reverseProxy
}
