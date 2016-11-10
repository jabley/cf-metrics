package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/cf/i18n"
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

// CFInfo is a representation of a Cloud Foundry zone. By Zone, we mean a Cloud
// Foundry provider region. IBM Bluemix has multiple zones in this sense of the
// word - US South, EU GB, and Australia.
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

// Metric is the common header for all documents emitted by this application.
// Each item should have these data points associated with it.
type Metric struct {
	Zone      string    `json:"zone"`
	Space     string    `json:"space"`
	App       string    `json:"app"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// AppMetrics is a document that represents the runtime statistics for a given
// application.
type AppMetrics struct {
	Metric
	Stats Stats `json:"stats,omitempty"`
}

// Usage is a fragment that represents instance-level statistics for an
// application.
type Usage struct {
	CPU       float64 `json:"cpu"`
	Disk      int64   `json:"disk"`
	Mem       int64   `json:"mem"`
	DiskUsage float64 `json:"diskUsage"`
	MemUsage  float64 `json:"memUsage"`
}

// ContainerStats is a fragment that represents instance-level statistics for an
// application.
type ContainerStats struct {
	DiskQuota int64 `json:"diskQuota"`
	MemQuota  int64 `json:"memQuota"`
	Usage     Usage `json:"usage"`
}

// InstanceStats is a fragment that aggregates all instance-level statistics for
// an application.
type InstanceStats struct {
	Stats ContainerStats `json:"stats"`
}

// Stats is a map that contains all of the instance stats for a given
// application. An application can have one or more instances, so we need to
// capture this multiplicity.
type Stats map[string]InstanceStats

// EventInfo is a fragment that describes an application-level event, such as
// restart, update and so on.
type EventInfo struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// Event is a document that represents a single event for an application.
type Event struct {
	Metric
	EventInfo EventInfo `json:"eventInfo"`
}

func main() {
	var (
		verbose   bool
		whitelist string
	)

	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.StringVar(&whitelist, "whitelist", "", "A comma-separated list of app names to collect data about. If none specified, then defaults to all apps that the account can view.")

	flag.Usage = func() {
		basename := filepath.Base(os.Args[0])
		fmt.Printf("Usage: %s\n", basename)
		fmt.Printf("\n"+
			"ZONE_PREFIXES – the environment variable that contains a comma-separated list of\n"+
			"prefixes for ENV vars that can be used for authenticating with a Cloud Foundry\n"+
			"provider.\n"+
			"\nFor example:\n\n> env ZONE_PREFIXES=PWS,CH PWS_USERNAME=user@example.com \\\n"+
			"PWS_PASSWORD='some-passphrase' \\\n"+
			"PWS_API=https://api.run.pivotal.io \\\n"+
			"PWS_NAME=pivotal \\\n"+
			"CH_USERNAME=user@example.com \\\n"+
			"CH_PASSWORD='some-other-passphrase' \\\n"+
			"CH_API=https://api.lyra-836.appcloud.swisscom.com \\\n"+
			"CH_NAME=swisscom %s\n\nOther arguments:\n\n", basename)
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

	logItem := func(item interface{}) {
		if err := encoder.Encode(item); err != nil {
			panic(fmt.Sprintf("Problem encoding %v\nGot %v\n", item, err))
		}
	}

	metrics := make(chan AppMetrics)
	events := make(chan Event)

	writer := os.Stdout

	logger := trace.NewLogger(writer, verbose, os.Getenv("CF_TRACE"), "")

	spawnWorkers(zones, whitelist, metrics, events, writer, logger)

	go func() {
		for e := range events {
			logItem(e)
		}
	}()

	for m := range metrics {
		logItem(m)
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
