package main

import (
	"log"
	"net/http"
)

type Service struct{}

func (*Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(501)
	w.Write([]byte("Not Implemented"))
}

func main() {
	log.Fatal(http.ListenAndServe(":8080", &Service{}))
}
