package main

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/appinstances"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

// AppRepo provides functionality to query a Cloud Foundry API for applications.
type AppRepo struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

// NewAppRepo returns a new SpaceRepo which internally uses the specified
// coreconfig.Repository and net.Gateway.
func NewAppRepo(config coreconfig.Repository, gateway net.Gateway) (repo AppRepo) {
	repo.config = config
	repo.gateway = gateway
	return
}

// ListApps lists all applicaitons in the organisation. For each application, it
// calls the provided callback function.
func (ar AppRepo) ListApps(callback func(models.Application) bool) error {
	return ar.gateway.ListPaginatedResources(
		ar.config.APIEndpoint(),
		"/v2/apps",
		resources.ApplicationResource{},
		func(resource interface{}) bool {
			return callback(resource.(resources.ApplicationResource).ToModel())
		})
}

// GetAppStats fetches the stats for the specified application. The application
// needs to be started.
func (ar AppRepo) GetAppStats(app models.Application) (statsResponse appinstances.StatsAPIResponse, err error) {
	path := fmt.Sprintf("%s/v2/apps/%s/stats", ar.config.APIEndpoint(), app.GUID)
	statsResponse = appinstances.StatsAPIResponse{}
	err = ar.gateway.GetResource(path, &statsResponse)
	return
}
