package main

import (
	"embed"
	"encoding/json"
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
		Token      string    `json:"token"`
		Domain     string    `json:"domain"`
		HasQueries bool      `json:"-"`
		Queries    []query_t `json:"queries"`
	}

	addr := fmt.Sprintf(":%d", port)
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/favicon.ico" {
			rw.WriteHeader(404)
			return
		}

		contentType := r.Header.Get("Accept")

		token := strings.TrimPrefix(r.URL.Path, "/")
		queries := get_results(token)
		model := model_t{
			Token:      token,
			Domain:     strings.TrimSuffix(ctx.domain, "."),
			HasQueries: len(queries) > 0,
			Queries:    queries,
		}

		logrus.Infof("http request for '%s' from '%s'", token, r.RemoteAddr)

		if contentType == "application/json" || strings.HasPrefix(contentType, "application/json;") {
			rw.Header().Add("Content-Type", "application/json")
			if jdata, e := json.Marshal(model); e != nil {
				rw.WriteHeader(500)
				logrus.Panicf("could not marshal json: %v", e)
			} else {
				rw.Write(jdata)
			}
		} else {
			if tmpl, e := template.ParseFS(f, "template.html"); e != nil {
				rw.WriteHeader(500)
				logrus.Panicf("could not parse template for request '%v': %v", r.RequestURI, e)
				return
			} else {
				tmpl.Execute(rw, model)
			}
		}
	})
	server := &http.Server{
		Addr: addr,
	}
	logrus.Infof("http server listening on '%s'", server.Addr)
	if e := server.ListenAndServe(); e != nil {
		logrus.Panic(e)
	}
	wg.Done()
}
