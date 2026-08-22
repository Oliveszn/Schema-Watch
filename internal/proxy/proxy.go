package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Oliveszn/Schema-Watch/internal/config"
	"github.com/Oliveszn/Schema-Watch/internal/schema"
	"github.com/Oliveszn/Schema-Watch/internal/store"
)

type OnDiff func(diff *schema.Diff)

type Proxy struct {
	target *url.URL
	rp     *httputil.ReverseProxy
	store  *store.Store
	onDiff OnDiff
	cfg    *config.Config
}

func New(targetURL string, st *store.Store, onDiff OnDiff, cfg *config.Config) (*Proxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	p := &Proxy{
		target: target,
		store:  st,
		onDiff: onDiff,
		cfg:    cfg,
	}

	rp := httputil.NewSingleHostReverseProxy(target)
	rp.ModifyResponse = p.captureResponse
	p.rp = rp

	return p, nil
}

func (p *Proxy) Handler() http.Handler {
	return p.rp
}

func (p *Proxy) captureResponse(resp *http.Response) error {
	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return nil
	}

	endpoint := endpointKey(resp.Request)
	if p.cfg.IsEndpointIgnored(endpoint) {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))

	sch, err := schema.Extract(body)
	if err != nil {
		return nil
	}
	sch = p.cfg.FilterSchema(sch)

	diff := p.store.CheckAndUpdate(endpoint, sch)
	if diff != nil && p.onDiff != nil {
		p.onDiff(diff)
	}

	return nil
}

func endpointKey(req *http.Request) string {
	return req.Method + " " + req.URL.Path
}
