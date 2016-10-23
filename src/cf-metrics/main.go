package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"
	// "github.com/cloudfoundry/cli/cf/configuration/coreconfig"
)

type CFInfo struct {
	API      string
	username string
	password string
}

// type Zone struct {
// 	config        coreconfig.ReadWriter
// 	authenticator authentication.Repository
// 	endpointRepo  coreconfig.EndpointRepository
// 	orgRepo       organizations.OrganizationRepository
// 	spaceRepo     spaces.SpaceRepository
// }

type AppMetrics struct {
}

func main() {
	// check that we have one or more CFs to poll
	//
	zones := parseZones()

	if len(zones) == 0 {
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)

	logMetric := func(metrics AppMetrics) {
		if err := encoder.Encode(metrics); err != nil {
			panic(err)
		}
	}

	metrics := make(chan AppMetrics)

	// wg := &sync.WaitGroup{}
	// wg.Add(1)
	go workLoop(zones, metrics)

	for m := range metrics {
		logMetric(m)
	}
	// wg.Wait()
}

func workLoop(zones []CFInfo, metrics chan AppMetrics) {
	t := time.NewTicker(time.Duration(10) * time.Second)

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

	zonePrefixes := strings.Split(getDefaultConfig("ZONE_PREFIXES", ""), ",")

	for _, prefix := range zonePrefixes {
		if len(prefix) == 0 {
			continue
		}
		username := getDefaultConfig(prefix+"_USERNAME", "")
		password := getDefaultConfig(prefix+"_PASSWORD", "")
		api := getDefaultConfig(prefix+"_API", "")

		zones = append(zones, CFInfo{API: api, username: username, password: password})
	}
	return zones
}

func getDefaultConfig(name, fallback string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return fallback
}
