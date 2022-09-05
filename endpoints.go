package main

import (
	"log"
	"net/http"
)

func icon() http.Handler {
	fav, err := web.ReadFile("web/icon.ico")
	if err != nil {
		log.Panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(fav)
	})
}

func land() http.Handler {
	land, err := web.ReadFile("web/index.min.html")
	if err != nil {
		log.Panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(land)
	})
}
