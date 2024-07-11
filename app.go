package sslhostsproxier

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type Logger func(log string)

type App struct {
	dnsRunning   bool
	httpRunning  bool
	httpsRunning bool

	ShuttingDown bool

	domainTargetProxyMap map[string]string

	dns         *DnsServer
	httpsServer *http.Server
	httpServer  *http.Server
}

func CreateApp(domainTargetProxyMap map[string]string) App {
	dnsServer := CreateDnsServer()
	httpsServer := CreateSSLReverseProxy(domainTargetProxyMap)
	httpServer := CreateHttpToSSLProxy(domainTargetProxyMap)

	return App{
		false,
		false,
		false,
		false,
		domainTargetProxyMap,
		dnsServer,
		httpsServer,
		httpServer,
	}
}

func (app *App) IsReady() bool {
	return app.dnsRunning && app.httpsRunning && app.httpRunning
}

func (app *App) Start(log Logger) {
	app.dns.NotifyStartedFunc = func() {
		log("DNS server started")
		app.dnsRunning = true
	}

	CreateTempCertsBundle()

	log("Please restart every cmd to make sure the REQUESTS_CA_BUNDLE env was updated")

	for domain, _ := range app.domainTargetProxyMap {
		err := CreateLocalNrptResolution(domain)
		if err != nil {
			panic(err)
		}
		log(fmt.Sprintf("[%s] Added NRPT entry", domain))

		certPath := GetServerCert(domain)
		TrustCertificate(certPath)

		err = AppendCustomCertToBundle(certPath)
		if err != nil {
			panic(err)
		}

		log(fmt.Sprintf("[%s] Added certificate to keystore", domain))
	}

	go func() {
		app.dns.ListenAndServe()
	}()

	go func() {
		ln, err := net.Listen("tcp", app.httpsServer.Addr)
		if err != nil {
			panic(err)
		}

		app.httpsRunning = true
		log("HTTPS proxy started")

		err = app.httpsServer.ServeTLS(ln, "", "")
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	go func() {
		l, err := net.Listen("tcp", app.httpServer.Addr)
		if err != nil {
			panic(err)
		}

		app.httpRunning = true
		log("HTTP->HTTPS proxy started")

		err = app.httpServer.Serve(l)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

}

func (app *App) Stop(log Logger) {
	app.ShuttingDown = true

	DeleteTempCertsBundle()
	log("Please restart every cmd to make sure the REQUESTS_CA_BUNDLE env was updated")

	for domain, _ := range app.domainTargetProxyMap {
		err := DeleteLocalNrptResolution(domain)
		if err != nil {
			panic(err)
		}
		log(fmt.Sprintf("[%s] Removed NRPT entry", domain))

		err = UntrustUs(domain)
		if err != nil {
			panic(err)
		}
		log(fmt.Sprintf("[%s] Removed certificate from keystore", domain))

	}

	err := app.httpsServer.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
	log("HTTPS proxy shutdown")

	err = app.httpServer.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
	log("HTTP->HTTPS proxy shutdown")

	app.dns.Shutdown()
	log("DNS server shutdown")
}
