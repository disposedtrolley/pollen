package main

import (
	"io"
	"log"
	"net/http"
)

func serveForecast(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "This is my website!\n")
}

func serve() {
	http.HandleFunc("/", serveForecast)
	log.Fatal(http.ListenAndServe(":4321", nil))
}
