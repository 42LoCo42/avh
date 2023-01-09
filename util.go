package main

import (
	"log"
	"net/http"
)

func noAuth(w http.ResponseWriter, name string) {
	log.Printf("User %s failed login!", name)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}

func badReq(w http.ResponseWriter, msg string) {
	log.Print(msg)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(msg))
}

func onErr(w http.ResponseWriter, err error) {
	log.Print(err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}
