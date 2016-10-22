package main

import (
	"os"
	"sync"
)

func main() {
	// check that we have one or more CFs to poll
	//
	wg := &sync.WaitGroup{}
	wg.Add(1)

	wg.Wait()
}

func getDefaultConfig(name, fallback string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return fallback
}
