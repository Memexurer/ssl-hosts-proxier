package sslhostsproxier

import (
	"context"
	"net"
	"net/http"
)

type Logger func(log string)

type App struct {
	dnsRunning   bool
	httpsRunning bool

	ShuttingDown bool

	domain string
	target string

	dns    *DnsServer
	server *http.Server
}

func CreateApp(domain string, target string) App {
	dnsServer := CreateDnsServer()
	httpsServer := CreateSSLReverseProxy(domain, target)

	return App{
		false,
		false,
		false,
		domain,
		target,
		dnsServer,
		httpsServer,
	}
}

func (app *App) IsReady() bool {
	return app.dnsRunning && app.httpsRunning
}

func (app *App) Start(log Logger) {
	app.dns.NotifyStartedFunc = func() {
		log("DNS server started")
		app.dnsRunning = true
	}
	err := CreateLocalNrptResolution(app.domain)
	if err != nil {
		panic(err)
	}
	log("NRPT resolution created")

	certPath := GetServerCert(app.domain)
	TrustCertificate(certPath)
	log("HTTPS reverse proxy certificate for " + app.domain + " added to system keystore")

	go func() {
		app.dns.ListenAndServe()
	}()

	go func() {
		ln, err := net.Listen("tcp", app.server.Addr)
		if err != nil {
			panic(err)
		}

		app.httpsRunning = true
		log("HTTPS proxy started")

		_ = app.server.ServeTLS(ln, "", "")
	}()

}

func (app *App) Stop(log Logger) {
	app.ShuttingDown = true

	err := DeleteLocalNrptResolution(app.domain)
	if err != nil {
		panic(err)
	}
	log("Deleted local NRPT resolution")

	err = UntrustUs(app.domain)
	if err != nil {
		panic(err)
	}
	log("Untrusted our certfificate")

	err = app.server.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
	log("HTTPS proxy shutdown")

	app.dns.Shutdown()
	log("DNS server shutdown")
}
