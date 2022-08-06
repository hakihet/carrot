package main

import (
	"context"
	_ "embed"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
)

type Done <-chan struct{}

func main() {
	crt := flag.String("crt", "", "cert")
	key := flag.String("key", "", "key")
	flag.Parse()

	var s http.Server = http.Server{
		Addr:    ":https",
		Handler: nil,
	}

	logger := log.Default()
	handler := http.NewServeMux()

	handler.Handle("/", recovers(logging(logger, http.FileServer(http.Dir("/var/carlotz/")))))

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	done := func(s chan os.Signal, f func()) Done {
		d := make(chan struct{})
		go func(s chan os.Signal, c chan<- struct{}) {
			defer close(c)
			<-s
			f()
		}(s, d)
		return d
	}(signals, func() {
		if err := s.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Shutdown: %v", err)
		}
	})

	if err := s.ListenAndServeTLS(*crt, *key); err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe: %v", err)
		return
	}
	<-done
}

func logging(logger *log.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Println("before")
		defer logger.Println("after")
		h.ServeHTTP(w, r)
	})
}

func recovers(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		h.ServeHTTP(w, r)
	})
}