package main

import (
	"log"
	"os"
	"time"
)

func main() {
	if len(os.Args[1:]) == 1 && os.Args[1] == "server" {
		serve()
	}

	if len(os.Args[1:]) == 1 && os.Args[1] == "tick" {
		log.Println("Periodically updating forecast every hour...")
		ticker := time.NewTicker(1 * time.Hour)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					if err := forecast(); err != nil {
						log.Printf("fetch forecast: %v", err)
					} else {
						log.Println("Successfully fetched forecast.")
					}
				}
			}
		}()

		<-quit
		return
	}

	if err := forecast(); err != nil {
		log.Fatal(err)
	}
}
