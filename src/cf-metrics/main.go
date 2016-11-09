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

type Metric struct {
	Zone      string    `json:"zone"`
	Space     string    `json:"space"`
	App       string    `json:"app"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

type AppMetrics struct {
	Metric
	Stats Stats `json:"stats,omitempty"`
}

type Usage struct {
	CPU       float64 `json:"cpu"`
	Disk      int64   `json:"disk"`
	Mem       int64   `json:"mem"`
	DiskUsage float64 `json:"disk-usage"`
	MemUsage  float64 `json:"mem-usage"`
}

type ContainerStats struct {
	DiskQuota int64 `json:"disk-quota"`
	MemQuota  int64 `json:"mem-quota"`
	Usage     Usage `json:"usage"`
}

type InstanceStats struct {
	Stats ContainerStats `json:"stats"`
}

type Stats map[string]InstanceStats

type EventInfo struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

type Event struct {
	Metric
	EventInfo EventInfo
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

	logItem := func(item interface{}) {
		if err := encoder.Encode(item); err != nil {
			panic(err)
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
