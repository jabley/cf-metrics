package main

import (
	"io"
	"os"
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/cf/api/appinstances"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
)

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

func spawnWorkers(cfInfos []CFInfo, whitelist string, metrics chan AppMetrics, events chan Event, writer io.Writer, logger trace.Printer) {
	homeDir := userHomeDir()
	errorHandler := func(err error) {
	}

	appWhitelist := parseWhitelist(whitelist)

	teePrinter := terminal.NewTeePrinter(writer)
	envDialTimeout := getDefaultConfig("CF_DIAL_TIMEOUT", "5")
	ui := terminal.NewUI(os.Stdin, writer, teePrinter, logger)

	for _, info := range cfInfos {
		zone := NewZone(info, homeDir, errorHandler, appWhitelist, envDialTimeout, ui, writer, logger)
		go readSpacesLoop(zone)
		go readAppsLoop(zone, metrics, events)
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

func readAppsLoop(zone *Zone, metrics chan AppMetrics, events chan Event) {
	t := time.NewTicker(time.Duration(10) * time.Second)

	since := time.Now()

	// channel used to do the initial poll
	start := make(chan struct{})

	// This one weird trick to do the initial poll
	go func() {
		start <- struct{}{}
	}()

	for {
		select {
		case <-start:
			pollApps(zone, since, metrics, events)
			since = time.Now()
		case <-t.C:
			pollApps(zone, since, metrics, events)
			since = time.Now()
		}
	}
}

func pollApps(zone *Zone, since time.Time, metrics chan AppMetrics, events chan Event) {
	now := time.Now()
	err := zone.appRepo.ListApps(func(app models.Application) bool {
		if !zone.IncludesApp(app.Name) {
			return true
		}
		if app.State == models.ApplicationStateStarted {
			go fetchStats(app, zone, metrics, now)
		}
		go fetchEvents(app, zone, events, now, since)
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
			Metric: Metric{
				Zone:      zone.name,
				Space:     zone.GetSpaceName(app.SpaceGUID),
				App:       app.Name,
				Type:      "metric",
				Timestamp: now,
			},
			Stats: cfStatsToExternalStats(stats),
		}
	}
}

func cfStatsToExternalStats(stats appinstances.StatsAPIResponse) (res Stats) {
	res = make(Stats)
	for k, v := range stats {
		diskUsage := calculateUsage(v.Stats.Usage.Disk, v.Stats.DiskQuota)
		memUsage := calculateUsage(v.Stats.Usage.Mem, v.Stats.MemQuota)
		res[k] = InstanceStats{
			Stats: ContainerStats{
				DiskQuota: v.Stats.DiskQuota,
				MemQuota:  v.Stats.MemQuota,
				Usage: Usage{
					CPU:       v.Stats.Usage.CPU,
					Disk:      v.Stats.Usage.Disk,
					Mem:       v.Stats.Usage.Mem,
					DiskUsage: diskUsage,
					MemUsage:  memUsage,
				},
			},
		}
	}
	return
}

func calculateUsage(usage, quota int64) float64 {
	if usage == 0 || quota == 0 {
		return 0.0
	}

	return float64(usage) / float64(quota)
}

func fetchEvents(app models.Application, zone *Zone, events chan Event, now time.Time, since time.Time) {
	err := zone.eventRepo.GetAppEvents(app, since, func(e models.EventFields) bool {
		events <- Event{
			Metric: Metric{
				Zone:      zone.name,
				Space:     zone.GetSpaceName(app.SpaceGUID),
				App:       app.Name,
				Type:      "event",
				Timestamp: now,
			},
			EventInfo: EventInfo{
				Type:      e.Name,
				Timestamp: e.Timestamp,
			},
		}
		return true
	})
	if err == nil {
		// Soft log it? Potentially a zone might have transient problems / scheduled maintenance.
	}
}
