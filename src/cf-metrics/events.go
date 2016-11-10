package main

import (
	"fmt"
	"time"

	"net/url"

	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

// EventRepo provides functionality to query a Cloud Foundry API for events
// related to an application.
type EventRepo struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

// NewEventRepo returns a new EventRepo which internally uses the specified
// coreconfig.Repository and net.Gateway.
func NewEventRepo(config coreconfig.Repository, gateway net.Gateway) (repo EventRepo) {
	repo.config = config
	repo.gateway = gateway
	return
}

// GetAppEvents fetches all events for the specified app that have happened
// since the specified time. For each event, it calls the provided callback function.
func (r EventRepo) GetAppEvents(app models.Application, since time.Time, callback func(models.EventFields) bool) error {
	return r.gateway.ListPaginatedResources(
		r.config.APIEndpoint(),
		fmt.Sprintf("/v2/events?q=actee:%s&q=timestamp%s", app.GUID, url.QueryEscape(">"+since.Format(time.RFC3339))),
		resources.EventResourceNewV2{},
		func(resource interface{}) bool {
			return callback(resource.(resources.EventResourceNewV2).ToFields())
		})
}
