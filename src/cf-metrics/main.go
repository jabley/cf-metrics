package main

import (
	"os"
	"sync"
	"time"
)

type CFInfo struct {
	API      string
	username string
	password string
}

func main() {
	// check that we have one or more CFs to poll
	//
	zones := parseZones()

	if len(zones) == 0 {
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go workLoop(zones)
	wg.Wait()
}

func workLoop(zones []CFInfo) {
	t := time.NewTicker(time.Duration(5) * time.Second)

	for {
		select {
		case <-t.C:
			//	for each zone
			//		for each org
			// 			for each space
			//				for each app
			//					fetch metrics
		}
	}
}

func parseZones() []CFInfo {
	zones := make([]CFInfo, 0)

	return zones
}

func getDefaultConfig(name, fallback string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return fallback
}
