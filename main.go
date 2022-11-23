package main

import (
	"os"
)

func main() {
	if len(os.Args[1:]) == 1 && os.Args[1] == "server" {
		serve()
	} else {
		forecast()
	}
}
