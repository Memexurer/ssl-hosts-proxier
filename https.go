package sslhostsproxier

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func CreateSSLReverseProxy(hostname string, dest string) *http.Server {
	mux := http.NewServeMux()

	remote, err := url.Parse(dest)
	if err != nil {
		panic(err)
	}

	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			r.Host = remote.Host
			p.ServeHTTP(w, r)
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	mux.HandleFunc("/", handler(proxy))

	server := &http.Server{
		Addr:      "127.0.0.1:443",
		Handler:   mux,
		TLSConfig: CreateTLSConfig(hostname),
	}

	return server
}
