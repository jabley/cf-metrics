package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/authentication"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/net"
)

type CFInfo struct {
	prefix   string
	API      string
	username string
	password string
}

type Zone struct {
	config        coreconfig.ReadWriter
	authenticator authentication.Repository
	endpointRepo  coreconfig.EndpointRepository
	orgRepo       organizations.OrganizationRepository
	spaceRepo     spaces.SpaceRepository
}

// See: http://stackoverflow.com/questions/7922270/obtain-users-home-directory
// we can't cross compile using cgo and use user.Current()
var userHomeDir = func() string {

	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}

	return os.Getenv("HOME")
}

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
	spawnWorkers(zones, metrics)

	for m := range metrics {
		logMetric(m)
	}
	// wg.Wait()
}

func spawnWorkers(cfInfos []CFInfo, metrics chan AppMetrics) {
	zones := make([]Zone, len(cfInfos))

	homeDir := userHomeDir()
	errorHandler := func(err error) {
	}

	for i, info := range cfInfos {
		// Ensure we can login to each zone
		zones[i] = Zone{}
		configPath := filepath.Join(homeDir, "."+info.prefix, "config.json")

		gateways := make(map[string]net.Gateway)
		zones[i].config = coreconfig.NewRepositoryFromFilepath(configPath, errorHandler)
		repoLocator := api.NewRepositoryLocator(zones[i].config, gateways, nil)
		zones[i].authenticator = repoLocator.GetAuthenticationRepository()
		zones[i].endpointRepo = repoLocator.GetEndpointRepository()
		zones[i].orgRepo = repoLocator.GetOrganizationRepository()
		zones[i].spaceRepo = repoLocator.GetSpaceRepository()
	}

	for _, zone := range zones {
		go workLoop(zone, metrics)
	}
}

func workLoop(zone Zone, metrics chan AppMetrics) {
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

		zones = append(zones, CFInfo{prefix: prefix,
			API:      api,
			username: username,
			password: password})
	}
	return zones
}

func getDefaultConfig(name, fallback string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return fallback
}
