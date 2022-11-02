package main

import (
	"context"
	"embed"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
)

//go:embed web/icon.ico web/index.min.html web/style.css
var web embed.FS

func init(){
	f, err := os.OpenFile("/var/log/carrot", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
}

func main() {
	crt := flag.String("crt", "", "cert")
	key := flag.String("key", "", "key")
	flag.Parse()

	handler := http.NewServeMux()
	landing := recovers(logs(land(web)))
	favicon := recovers(logs(icon(web)))
	handler.Handle("/favicon.ico",favicon )
	handler.Handle("/", landing )
	handler.Handle("/authenicate", recovers(auth()))

	s := http.Server{
		Addr:    ":https",
		Handler: handler,
	}

	shutdown := func() {
		if err := s.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Shutdown: %v", err)
		}
	}

	done := func(f func()) <-chan struct{} {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		d := make(chan struct{})
		go func() {
			defer close(d)
			<-signals
			f()
		}()
		return d
	}(shutdown)

	if err := s.ListenAndServeTLS(*crt, *key); err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe: %v", err)
		return
	}
	<-done
}
