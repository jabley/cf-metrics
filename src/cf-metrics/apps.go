package main

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/appinstances"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

type AppRepo struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

func NewAppRepo(config coreconfig.Repository, gateway net.Gateway) (repo AppRepo) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (ar AppRepo) ListApps(callback func(models.Application) bool) error {
	return ar.gateway.ListPaginatedResources(
		ar.config.APIEndpoint(),
		"/v2/apps",
		resources.ApplicationResource{},
		func(resource interface{}) bool {
			return callback(resource.(resources.ApplicationResource).ToModel())
		})
}

func (ar AppRepo) GetAppStats(app models.Application) (statsResponse appinstances.StatsAPIResponse, err error) {
	path := fmt.Sprintf("%s/v2/apps/%s/stats", ar.config.APIEndpoint(), app.GUID)
	statsResponse = appinstances.StatsAPIResponse{}
	err = ar.gateway.GetResource(path, &statsResponse)
	return
}
