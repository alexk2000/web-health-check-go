package web

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"web-health-check/pkg/config"
)

var httpServer *http.Server

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(`{"status": "ok"}`))
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`{"version": "%v"}`, os.Getenv("VERSION"))))
}

func createHttpServer(conf config.Config) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/version", versionHandler)
	s := &http.Server{
		Addr:    fmt.Sprintf(":%v", conf.Port),
		Handler: mux,
	}

	return s
}

func runHttpServer(conf config.Config) {
	httpServer = createHttpServer(conf)
	go func() {
		log.Println(httpServer.ListenAndServe())
	}()
}

func reloader() {
	for conf := range config.Conf.SubscribeOnChange() {
		log.Println("Reloading web server")
		httpServer.Close()
		runHttpServer(conf)
	}
}

func StartAsync() {
	conf := config.Conf.Get()
	runHttpServer(conf)
	go reloader()
}
