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

type EventRepo struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

func NewEventRepo(config coreconfig.Repository, gateway net.Gateway) (repo EventRepo) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (r EventRepo) GetAppEvents(app models.Application, since time.Time, callback func(models.EventFields) bool) error {
	return r.gateway.ListPaginatedResources(
		r.config.APIEndpoint(),
		fmt.Sprintf("/v2/events?q=actee:%s&q=timestamp%3E%s", app.GUID, url.QueryEscape(since.Format(time.RFC3339))),
		resources.EventResourceNewV2{},
		func(resource interface{}) bool {
			return callback(resource.(resources.EventResourceNewV2).ToFields())
		})
}
