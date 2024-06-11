package sslhostsproxier

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
)

type Proxy struct {
	proxy  *httputil.ReverseProxy
	remote *url.URL
}

var proxies map[string]Proxy = make(map[string]Proxy)

func CreateHttpToSSLProxy(hosts map[string]string) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r,
			"https://"+r.Host+r.URL.String(),
			http.StatusMovedPermanently)
	})

	server := &http.Server{
		Addr:    "127.0.0.1:80",
		Handler: mux,
	}

	return server
}

func CreateSSLReverseProxy(hosts map[string]string) *http.Server {
	mux := http.NewServeMux()

	for domain, dest := range hosts {
		remote, err := url.Parse(dest)
		if err != nil {
			panic(err)
		}

		proxies[domain] = Proxy{
			httputil.NewSingleHostReverseProxy(remote),
			remote,
		}
	}

	handler := func() func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			proxy, ok := proxies[r.Host]
			if ok {
				r.Host = proxy.remote.Host
				proxy.proxy.ServeHTTP(w, r)
			} else {
				w.Write([]byte("unknown host"))
				w.WriteHeader(400)
			}
		}
	}

	mux.HandleFunc("/", handler())

	mappedToSlice := reflect.ValueOf(hosts).MapKeys()
	domains := make([]string, len(mappedToSlice))
	for i, v := range mappedToSlice {
		domains[i] = v.Interface().(string)
	}

	server := &http.Server{
		Addr:      "127.0.0.1:443",
		Handler:   mux,
		TLSConfig: CreateTLSConfig(domains),
	}

	return server
}
