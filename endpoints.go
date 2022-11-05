package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

// Create the JWT key used to create the signature
var key = []byte("my_secret_key")

var users = map[string]string{
	"me": "123",
}

type Credentials struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type Claims struct {
	Account string `json:"username"`
	jwt.StandardClaims
}

func auth() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var cred Credentials
		err := json.NewDecoder(r.Body).Decode(&cred)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		expectedPassword, ok := users[cred.Account]
		if !ok || expectedPassword != cred.Password {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expirationTime := time.Now().Add(5 * time.Minute)
		claims := &Claims{
			Account: cred.Account,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
		})
	})
}

func icon(fs embed.FS) http.Handler {
	fav, err := fs.ReadFile("web/icon.ico")
	if err != nil {
		log.Panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(fav)
	})
}

func land(fs embed.FS) http.Handler {
	land, err := fs.ReadFile("web/index.min.html")
	if err != nil {
		log.Panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(land)
	})
}
