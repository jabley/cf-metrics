package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/cf/api/appinstances"
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/utils/config"
)

var _ = initI18nFunc()

func initI18nFunc() bool {
	config, err := config.LoadConfig()
	if err != nil {
		fmt.Println(terminal.FailureColor("FAILED"))
		fmt.Println("Error read/writing config: ", err.Error())
		os.Exit(1)
	}
	i18n.T = i18n.Init(config)
	return true
}

type CFInfo struct {
	ZoneName string
	Prefix   string
	API      string
	Username string
	Password string
}

func (i CFInfo) String() string {
	return fmt.Sprintf("{Prefix: %s, API: <%s>, Username: <%s>}", i.Prefix, i.API, i.Username)
}

type AppMetrics struct {
	Zone      string    `json:"zone"`
	Space     string    `json:"space"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Stats     appinstances.StatsAPIResponse
	// TODO add events as well
}

func main() {
	var (
		verbose bool
	)

	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	flag.Usage = func() {
		basename := filepath.Base(os.Args[0])
		fmt.Printf("Usage: %s\n", basename)
		fmt.Printf("\n"+
			"ZONE_PREFIXES – the environment variable that contains a CSV list of prefixes\n"+
			"for ENV vars that can be used for authenticating with a Cloud Foundry\n"+
			"\nFor example:\n\n> env ZONE_PREFIXES=PWS PWS_USERNAME=someuser@example.com \\\n"+
			"  PWS_API=https://api.run.pivotal.io \\\n  PWS_PASSWORD=some-complex-passphrase %s\n\n", basename)
		flag.PrintDefaults()
	}

	flag.Parse()
	zones := parseZones()

	// check that we have one or more CFs to poll
	if len(zones) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)

	logMetric := func(metrics AppMetrics) {
		if err := encoder.Encode(metrics); err != nil {
			panic(err)
		}
	}

	metrics := make(chan AppMetrics)

	writer := os.Stdout

	logger := trace.NewLogger(writer, verbose, os.Getenv("CF_TRACE"), "")

	spawnWorkers(zones, metrics, writer, logger)

	for m := range metrics {
		logMetric(m)
	}
}

func spawnWorkers(cfInfos []CFInfo,
	metrics chan AppMetrics,
	writer io.Writer,
	logger trace.Printer) {
	zones := NewZones(cfInfos, writer, logger)

	for _, zone := range zones {
		go readSpacesLoop(&zone)
		go readMetricsLoop(&zone, metrics)
	}
}

func readSpacesLoop(zone *Zone) {
	t := time.NewTicker(time.Duration(1) * time.Minute)

	// channel used to do the initial poll
	start := make(chan struct{})

	// This one weird trick to do the initial poll
	go func() {
		start <- struct{}{}
	}()

	for {
		select {
		case <-start:
			pollSpaces(zone)
		case <-t.C:
			pollSpaces(zone)
		}
	}
}

// pollSpaces takes a pointer to the Zone, since it needs to update the spaces map.
func pollSpaces(zone *Zone) {
	spaces := make(map[string]string)

	err := zone.spaceRepo.ListSpaces(func(space models.Space) bool {
		spaces[space.GUID] = space.Name
		return true
	})

	if err != nil {
		// We'll try again later.
		return
	}

	zone.muSpaces.Lock()
	defer zone.muSpaces.Unlock()
	zone.spaces = spaces
}

func readMetricsLoop(zone *Zone, metrics chan AppMetrics) {
	t := time.NewTicker(time.Duration(10) * time.Second)

	// channel used to do the initial poll
	start := make(chan struct{})

	// This one weird trick to do the initial poll
	go func() {
		start <- struct{}{}
	}()

	for {
		select {
		case <-start:
			pollMetrics(zone, metrics)
		case <-t.C:
			pollMetrics(zone, metrics)
		}
	}
}

func pollMetrics(zone *Zone, metrics chan AppMetrics) {
	now := time.Now()
	err := zone.appRepo.ListApps(func(app models.Application) bool {
		if app.State == models.ApplicationStateStarted {
			go fetchStats(app, zone, metrics, now)
		}
		return true
	})

	if err != nil {
		// Soft log it? Potentially a zone might have transient problems / scheduled maintenance.
	}
}

func fetchStats(app models.Application, zone *Zone, metrics chan AppMetrics, now time.Time) {
	stats, err := zone.appRepo.GetAppStats(app)
	if err == nil {
		metrics <- AppMetrics{
			Zone:      zone.name,
			Name:      app.Name,
			Timestamp: now,
			Stats:     stats,
			Space:     zone.GetSpaceName(app.SpaceGUID),
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
		zoneName := getDefaultConfig(prefix+"_NAME", prefix)

		zones = append(zones, CFInfo{
			ZoneName: zoneName,
			Prefix:   prefix,
			API:      api,
			Username: username,
			Password: password,
		})
	}
	return zones
}
