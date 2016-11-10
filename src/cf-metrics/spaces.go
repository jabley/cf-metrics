package main

import (
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

// SpaceRepo provides functionality to query a Cloud Foundry API for spaces
type SpaceRepo struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

// NewSpaceRepo returns a new SpaceRepo which internally uses the specified
// coreconfig.Repository and net.Gateway.
func NewSpaceRepo(config coreconfig.Repository, gateway net.Gateway) (repo SpaceRepo) {
	repo.config = config
	repo.gateway = gateway
	return
}

// ListSpaces lists all spaces in the organisation. For each space, it calls the
// provided callback function.
func (sr SpaceRepo) ListSpaces(callback func(models.Space) bool) error {
	return sr.gateway.ListPaginatedResources(
		sr.config.APIEndpoint(),
		"/v2/spaces",
		resources.SpaceResource{},
		func(resource interface{}) bool {
			return callback(resource.(resources.SpaceResource).ToModel())
		})
}
