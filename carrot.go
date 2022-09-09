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

func main() {
	crt := flag.String("crt", "", "cert")
	key := flag.String("key", "", "key")
	flag.Parse()

	f, err := os.OpenFile("/var/log/carrot", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(f, "carrot: ", log.LstdFlags)
	handler := http.NewServeMux()
	handler.Handle("/favicon.ico", recovers(logging(logger, icon())))
	handler.Handle("/", recovers(logging(logger, land())))

	var s http.Server = http.Server{
		Addr:    ":https",
		Handler: handler,
	}

	done := func() <-chan struct{} {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		d := make(chan struct{})
		go func() {
			defer close(d)
			<-signals
			if err := s.Shutdown(context.Background()); err != nil {
				log.Printf("HTTP Shutdown: %v", err)
			}
		}()
		return d
	}()

	if err := s.ListenAndServeTLS(*crt, *key); err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe: %v", err)
		return
	}
	<-done
}
