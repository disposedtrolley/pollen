package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args[1:]) == 1 && os.Args[1] == "server" {
		serve()
	} else {
		log.Fatal(forecast())
	}
}
