package main

import (
	"log"
	"net/http"
	"os"
)

var logger = log.New(os.Stdout, "[http]: ", log.LstdFlags)

func LogRequest(r *http.Request) {
	logger.Printf("Request: %s %s", r.Method, r.URL.String())
}
