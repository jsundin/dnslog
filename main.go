package main

import (
	"flag"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

type query_t struct {
	When  string `json:"timestamp"`
	TxID  uint16 `json:"txid"`
	QType string `json:"type"`
	Name  string `json:"name"`
}

type context_t struct {
	domain     string
	forward    bool
	forwardDNS string
}

const resolvConf = "/etc/resolv.conf"

var result_lock *sync.Mutex = &sync.Mutex{}
var results map[string][]query_t = make(map[string][]query_t)

func (ctx *context_t) add_result(m *dns.Msg) {
	result_lock.Lock()
	defer result_lock.Unlock()

	for _, q := range m.Question {
		nameLower := strings.ToLower(q.Name)
		if !strings.HasSuffix(nameLower, ctx.domain) {
			logrus.Warnf("unexpected question for domain '%v' -- ignoring", q.Name)
			continue
		}
		token := strings.TrimSuffix(nameLower, ctx.domain)
		if token == "" || token == "." {
			logrus.Warnf("unexpected question for domain '%v' without token -- ignoring", q.Name)
			continue
		}
		token = strings.TrimSuffix(token, ".")
		if tokeni := strings.LastIndex(token, "."); tokeni > 0 {
			token = token[tokeni+1:]
		}
		logrus.Infof("dns question with token '%s': {id=%04x, type=%s, name=%s}", token, m.Id, dns.TypeToString[q.Qtype], q.Name)

		query := query_t{
			When:  time.Now().Format("2006-01-02 15:04:05"),
			TxID:  m.Id,
			QType: dns.TypeToString[q.Qtype],
			Name:  q.Name,
		}
		results[token] = append(results[token], query)
	}
}

func get_results(token string) []query_t {
	result_lock.Lock()
	defer result_lock.Unlock()
	return append([]query_t{}, results[token]...)
}

func main() {
	var dnsPort int
	var httpPort int
	var logLevel string
	ctx := &context_t{}

	flag.IntVar(&dnsPort, "dns", 1053, "dns port")
	flag.IntVar(&httpPort, "http", 8080, "http port")
	flag.StringVar(&ctx.domain, "domain", "dnslog.lab.", "domain (with a . at the end)")
	flag.BoolVar(&ctx.forward, "forward", false, "forward non-domain queries to upstream dns")
	flag.StringVar(&ctx.forwardDNS, "upstream", "", "which upstream dns to use (defaults to whatever is in resolv.conf, e.g 'ns.server.com:53', port optional)")
	flag.StringVar(&logLevel, "level", "info", "loglevel")
	flag.Parse()

	if lvl, e := logrus.ParseLevel(logLevel); e != nil {
		logrus.Panicf("bad loglevel: '%s'", logLevel)
	} else {
		logrus.SetLevel(lvl)
	}

	if ctx.forwardDNS != "" && !ctx.forward {
		logrus.Panicf("upstream dns can only be set if forward is enabled")
	}

	if !strings.HasSuffix(ctx.domain, ".") {
		logrus.Panicf("domain must end with a '.': '%s'", ctx.domain)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go ctx.dns_main(wg, dnsPort)
	go ctx.http_main(wg, httpPort)
	wg.Wait()
}

func ig(v ...interface{}) {}
