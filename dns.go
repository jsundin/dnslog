package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

func (ctx *context_t) dns_main(wg *sync.WaitGroup, port int) {
	var e error

	dns.HandleFunc(ctx.domain, ctx.dns_resolver)

	if ctx.forward {
		var upstreamServer string
		if ctx.forwardDNS == "" {
			clientConf, e := dns.ClientConfigFromFile(resolvConf)
			if e != nil {
				logrus.Panic(e)
			}
			upstreamServer = clientConf.Servers[0]
		} else {
			upstreamServer = ctx.forwardDNS
		}
		if !strings.Contains(upstreamServer, ":") {
			upstreamServer += ":53"
		}
		logrus.Warnf("will forward unknown queries to '%s' -- DON'T USE IN PRODUCTION!", upstreamServer)
		dns.HandleFunc(".", func(rw dns.ResponseWriter, m *dns.Msg) {
			ctx.dns_forwarder(rw, m, upstreamServer)
		})
	}

	addr := fmt.Sprintf(":%d", port)
	server := &dns.Server{
		Addr: addr,
		Net:  "udp4",
	}
	logrus.Infof("dns server listening on '%s'", server.Addr)
	if e = server.ListenAndServe(); e != nil {
		logrus.Panic(e)
	}
	wg.Done()
}

func (ctx *context_t) dns_resolver(rw dns.ResponseWriter, m *dns.Msg) {
	logrus.Debugf("local dns query: %v", m)
	ctx.add_result(m)
	reply := &dns.Msg{}
	reply.SetRcode(m, dns.RcodeNameError)
	rw.WriteMsg(reply)

}

func (ctx *context_t) dns_forwarder(rw dns.ResponseWriter, m *dns.Msg, upstreamServer string) {
	qstr := "["
	for i := 0; i < len(m.Question); i++ {
		if i > 0 {
			qstr += ", "
		}
		qstr += fmt.Sprintf("%s %s", dns.TypeToString[m.Question[i].Qtype], m.Question[i].Name)
	}
	qstr += "]"

	logrus.Debugf("upstream dns query: %v", m.Question)
	if reply, e := dns.Exchange(m, upstreamServer); e == nil {
		if reply != nil {
			rw.WriteMsg(reply)
		} else {
			reply = &dns.Msg{}
			reply.SetRcode(m, dns.RcodeServerFailure)
			rw.WriteMsg(reply)
		}
	} else {
		logrus.Warnf("upstream query '%s' failed: %v", qstr, e)
		reply = &dns.Msg{}
		reply.SetRcode(m, dns.RcodeServerFailure)
		rw.WriteMsg(reply)
	}
}
