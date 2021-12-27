package main

import (
	"fmt"
	"sync"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

func (ctx *context_t) dns_main(wg *sync.WaitGroup, port int) {
	addr := fmt.Sprintf(":%d", port)
	dns.HandleFunc(ctx.domain, func(rw dns.ResponseWriter, m *dns.Msg) {
		logrus.Debugf("dns query: %v", m)
		ctx.add_result(m)
		reply := &dns.Msg{}
		reply.SetRcode(m, dns.RcodeNameError)
		rw.WriteMsg(reply)
	})
	server := &dns.Server{
		Addr: addr,
		Net:  "udp4",
	}
	logrus.Infof("dns server listening on '%s'", server.Addr)
	e := server.ListenAndServe()
	if e != nil {
		logrus.Panic(e)
	}
	wg.Done()
}
