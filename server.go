package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func serveForecast(w http.ResponseWriter, r *http.Request) {
	if err := forecast(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	latestForecast, err := selectForecast(time.Now().UTC())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(latestForecast)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func serve() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", serveForecast)

	log.Printf("Listening on :%s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
