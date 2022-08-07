package main

import (
	"context"
	"embed"
	_ "embed"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
)

type Done <-chan struct{}

//go:embed web/*
var web embed.FS

func main() {
	crt := flag.String("crt", "", "cert")
	key := flag.String("key", "", "key")
	flag.Parse()

	logger := log.Default()
	handler := http.NewServeMux()
	handler.Handle("/favicon.ico", recovers(logging(logger, fav())))
	handler.Handle("/", recovers(logging(logger, land())))

	var s http.Server = http.Server{
		Addr:    ":https",
		Handler: handler,
	}

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

func fav() http.Handler {
	fav, err := web.ReadFile("favicon.ico")
	log.Println(err)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(fav)
	})
}

func land() http.Handler {
	land, err := web.ReadFile("index.html")
	log.Println(err)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(land)
	})
}

func logging(logger *log.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("Connection from %v requesting %v", r.RemoteAddr, r.RequestURI)
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
