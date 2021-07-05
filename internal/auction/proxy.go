package auction

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/nukowsk/bukowskis/internal/types"
)

type Proxy struct {
	url     *url.URL
	handler http.Handler
}

func NewProxy(url *url.URL) *Proxy {
	handler := httputil.NewSingleHostReverseProxy(url)

	handler.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("vanillaProxyError: %s", err)
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	return &Proxy{
		url,
		handler,
	}
}

func (p Proxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	req.URL.Host = p.url.Host
	req.URL.Scheme = p.url.Scheme
	// req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Header.Set("X-Bukowskis-Version", "0.01")
	req.Host = p.url.Host
	p.handler.ServeHTTP(res, req)
}

type MockProxy struct{}

func (mp MockProxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	body := types.JsResponse{}
	err := json.NewEncoder(res).Encode(body)
	if err != nil {
		log.Printf("Encoding error %s", err)
	}
}
