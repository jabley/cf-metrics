package main

import (
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

type SpaceRepo struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

func NewSpaceRepo(config coreconfig.Repository, gateway net.Gateway) (repo SpaceRepo) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (sr SpaceRepo) ListSpaces(callback func(models.Space) bool) error {
	return sr.gateway.ListPaginatedResources(
		sr.config.APIEndpoint(),
		"/v2/spaces",
		resources.SpaceResource{},
		func(resource interface{}) bool {
			return callback(resource.(resources.SpaceResource).ToModel())
		})
}
