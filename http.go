package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

//go:embed template.html
var f embed.FS

func (ctx *context_t) http_main(wg *sync.WaitGroup, port int) {
	type model_t struct {
		Token      string
		Domain     string
		HasQueries bool
		Queries    []query_t
	}

	addr := fmt.Sprintf(":%d", port)
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/favicon.ico" {
			rw.WriteHeader(404)
			return
		}

		token := strings.TrimPrefix(r.URL.Path, "/")
		queries := get_results(token)
		model := model_t{
			Token:      token,
			Domain:     strings.TrimSuffix(ctx.domain, "."),
			HasQueries: len(queries) > 0,
			Queries:    queries,
		}

		tmpl, e := template.ParseFS(f, "template.html")
		if e != nil {
			rw.WriteHeader(500)
			logrus.Panicf("could not parse template for request '%v': %v", r.RequestURI, e)
			return
		}
		tmpl.Execute(rw, model)
	})
	server := &http.Server{
		Addr: addr,
	}
	logrus.Infof("http server listening on '%s'", server.Addr)
	e := server.ListenAndServe()
	if e != nil {
		logrus.Panic(e)
	}
	wg.Done()
}
