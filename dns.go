package sslhostsproxier

import (
	"log"
	"strings"

	"github.com/miekg/dns"
)

// https://github.com/virtualzone/go-hole/tree/main

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		name := strings.ToLower(q.Name)
		res, errCode := processDnsQuery(name, q.Qtype)
		m.Answer = append(m.Answer, res...)
		m.Rcode = errCode
	}
}

func processDnsQuery(name string, qtype uint16) ([]dns.RR, int) {
	if qtype != dns.TypeA {
		return []dns.RR{}, dns.RcodeNameError
	}

	rr, err := dns.NewRR(name + " A 127.0.0.1")
	if err != nil {
		log.Println(err)
		return []dns.RR{}, dns.RcodeNameError
	}

	arr := make([]dns.RR, 1)
	arr[0] = rr

	return arr, dns.RcodeSuccess
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
	w.Close()
}

type DnsServer struct {
	dns.Server
}

func CreateDnsServer() *DnsServer {
	dns.HandleFunc(".", handleDnsRequest)

	return &DnsServer{
		dns.Server{
			Addr: "127.0.0.1:53",
			Net:  "udp",
		},
	}
}
