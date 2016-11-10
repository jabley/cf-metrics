package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"

	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
)

type Zone struct {
	name        string
	config      coreconfig.ReadWriter
	repoLocator api.RepositoryLocator
	appRepo     AppRepo
	spaceRepo   SpaceRepo
	eventRepo   EventRepo

	whitelist map[string]bool

	muSpaces sync.RWMutex
	spaces   map[string]string
}

func (z *Zone) GetSpaceName(guid string) string {
	z.muSpaces.RLock()
	defer z.muSpaces.RUnlock()

	if name, ok := z.spaces[guid]; ok {
		return name
	}

	return guid
}

func (z *Zone) IncludesApp(name string) bool {
	if len(z.whitelist) == 0 {
		return true
	}

	_, isPresent := z.whitelist[name]

	return isPresent
}

func NewZone(info CFInfo, homeDir string, errorHandler func(error), whitelist map[string]bool, envDialTimeout string, ui terminal.UI, writer io.Writer, logger trace.Printer) *Zone {
	config := NewRepositoryConfig(homeDir, info, errorHandler)
	repoLocator, cloudController := NewRepoLocator(config, info, envDialTimeout, ui, logger)

	err := setAPIEndpoint(info.API, config, repoLocator.GetEndpointRepository())

	if err != nil {
		panic(err)
	}

	verifyLogin(repoLocator, info)

	return &Zone{
		name:        info.ZoneName,
		config:      config,
		repoLocator: repoLocator,
		appRepo:     NewAppRepo(config, cloudController),
		spaceRepo:   NewSpaceRepo(config, cloudController),
		eventRepo:   NewEventRepo(config, cloudController),
		whitelist:   whitelist,
	}
}

func parseWhitelist(whitelist string) (res map[string]bool) {
	res = make(map[string]bool)

	if len(whitelist) == 0 {
		return
	}

	apps := strings.Split(whitelist, ",")

	for _, a := range apps {
		if len(a) != 0 {
			res[a] = true
		}
	}
	return
}

func verifyLogin(repoLocator api.RepositoryLocator, info CFInfo) {
	authenticator := repoLocator.GetAuthenticationRepository()

	credentials := make(map[string]string)
	credentials["password"] = info.Password
	credentials["username"] = info.Username

	err := authenticator.Authenticate(credentials)

	if err != nil {
		panic(err)
	}
}

func NewRepositoryConfig(homeDir string, info CFInfo, errorHandler func(error)) coreconfig.Repository {
	configPath := filepath.Join(homeDir, "."+info.Prefix, "config.json")
	config := coreconfig.NewRepositoryFromFilepath(configPath, errorHandler)
	config.SetAPIEndpoint(info.API)
	return config
}

func NewRepoLocator(config coreconfig.Repository, info CFInfo, envDialTimeout string, ui terminal.UI, logger trace.Printer) (api.RepositoryLocator, net.Gateway) {
	gateways := map[string]net.Gateway{
		"cloud-controller": net.NewCloudControllerGateway(config, time.Now, ui, logger, envDialTimeout),
		"uaa":              net.NewUAAGateway(config, ui, logger, envDialTimeout),
		"routing-api":      net.NewRoutingAPIGateway(config, time.Now, ui, logger, envDialTimeout),
	}

	return api.NewRepositoryLocator(config, gateways, logger, envDialTimeout), gateways["cloud-controller"]
}

func setAPIEndpoint(endpoint string, config coreconfig.ReadWriter, endpointRepo coreconfig.EndpointRepository) error {
	if strings.HasSuffix(endpoint, "/") {
		endpoint = strings.TrimSuffix(endpoint, "/")
	}

	config.SetSSLDisabled(false)

	refresher := coreconfig.APIConfigRefresher{
		Endpoint:     endpoint,
		EndpointRepo: endpointRepo,
		Config:       config,
	}

	warning, err := refresher.Refresh()
	if err != nil {
		config.SetAPIEndpoint("")
		config.SetSSLDisabled(false)

		switch typedErr := err.(type) {
		case *errors.InvalidSSLCert:
			tipMessage := fmt.Sprintf("TIP: Every time you disable certificate checks, a kitten dies. Why would you kill a kitten?")
			return errors.New(i18n.T("Invalid SSL Cert for {{.URL}}\n{{.TipMessage}}",
				map[string]interface{}{"URL": typedErr.URL, "TipMessage": tipMessage}))
		default:
			return typedErr
		}
	}

	if warning != nil {
		return errors.New(warning.Warn())
	}
	return nil
}
