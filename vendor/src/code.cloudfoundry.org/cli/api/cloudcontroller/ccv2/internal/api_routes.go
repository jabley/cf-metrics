package internal

import (
	"net/http"

	"github.com/tedsuo/rata"
)

const (
	AppsFromRouteRequest          = "AppsFromRoute"
	AppsRequest                   = "Apps"
	DeleteRouteRequest            = "DeleteRoute"
	DeleteServiceBindingRequest   = "DeleteServiceBinding"
	InfoRequest                   = "Info"
	PrivateDomainRequest          = "PrivateDomain"
	RouteMappingsFromRouteRequest = "RouteMappingsFromRoute"
	RoutesFromSpaceRequest        = "RoutesFromSpace"
	ServiceBindingsRequest        = "ServiceBindings"
	ServiceInstancesRequest       = "ServiceInstances"
	SharedDomainRequest           = "SharedDomain"
	SpaceServiceInstancesRequest  = "SpaceServiceInstances"
)

var APIRoutes = rata.Routes{
	{Path: "/v2/apps", Method: http.MethodGet, Name: AppsRequest},
	{Path: "/v2/info", Method: http.MethodGet, Name: InfoRequest},
	{Path: "/v2/private_domains/:private_domain_guid", Method: http.MethodGet, Name: PrivateDomainRequest},
	{Path: "/v2/routes/:route_guid", Method: http.MethodDelete, Name: DeleteRouteRequest},
	{Path: "/v2/routes/:route_guid/apps", Method: http.MethodGet, Name: AppsFromRouteRequest},
	{Path: "/v2/routes/:route_guid/route_mappings", Method: http.MethodGet, Name: RouteMappingsFromRouteRequest},
	{Path: "/v2/service_bindings", Method: http.MethodGet, Name: ServiceBindingsRequest},
	{Path: "/v2/service_bindings/:service_binding_guid", Method: http.MethodDelete, Name: DeleteServiceBindingRequest},
	{Path: "/v2/service_instances", Method: http.MethodGet, Name: ServiceInstancesRequest},
	{Path: "/v2/shared_domains/:shared_domain_guid", Method: http.MethodGet, Name: SharedDomainRequest},
	{Path: "/v2/spaces/:guid/service_instances", Method: http.MethodGet, Name: SpaceServiceInstancesRequest},
	{Path: "/v2/spaces/:space_guid/routes", Method: http.MethodGet, Name: RoutesFromSpaceRequest},
}
