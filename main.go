package main

import (
	"log"

	"github.com/benlocal/ai4/internal/runtime"
)

func main() {
	err := runtime.Start()
	if err != nil {
		log.Fatalf("Error starting runtime: %v", err)
	}

}
