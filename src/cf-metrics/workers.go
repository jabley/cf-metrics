package main

import (
	"io"
	"time"

	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/trace"
)

func spawnWorkers(cfInfos []CFInfo,
	metrics chan AppMetrics,
	writer io.Writer,
	logger trace.Printer) {
	zones := NewZones(cfInfos, writer, logger)

	for _, each := range zones {
		zone := each
		go readSpacesLoop(&zone)
		go readAppsLoop(&zone, metrics)
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

func readAppsLoop(zone *Zone, metrics chan AppMetrics) {
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
			pollApps(zone, metrics)
		case <-t.C:
			pollApps(zone, metrics)
		}
	}
}

func pollApps(zone *Zone, metrics chan AppMetrics) {
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
