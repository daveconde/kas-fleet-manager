package routes

import (
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/services/sso"
	"net/http"

	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/connector/internal/config"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/connector/internal/generated"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/connector/internal/handlers"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/acl"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/api"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/auth"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/db"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/environments"
	kerrors "github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/errors"
	coreHandlers "github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/handlers"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/server"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/shared"
	"github.com/goava/di"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type options struct {
	di.Inject
	ConnectorsConfig          *config.ConnectorsConfig
	ServerConfig              *server.ServerConfig
	ErrorsHandler             *coreHandlers.ErrorHandler
	AuthorizeMiddleware       *acl.AccessControlListMiddleware
	KeycloakService           sso.KafkaKeycloakService
	AuthAgentService          auth.AuthAgentService
	ConnectorAdminHandler     *handlers.ConnectorAdminHandler
	ConnectorTypesHandler     *handlers.ConnectorTypesHandler
	ConnectorsHandler         *handlers.ConnectorsHandler
	ConnectorClusterHandler   *handlers.ConnectorClusterHandler
	ConnectorNamespaceHandler *handlers.ConnectorNamespaceHandler
	DB                        *db.ConnectionFactory
}

func NewRouteLoader(s options) environments.RouteLoader {
	return &s
}

func (s *options) AddRoutes(mainRouter *mux.Router) error {

	authorizeMiddleware := s.AuthorizeMiddleware.Authorize
	requireOrgID := auth.NewRequireOrgIDMiddleware().RequireOrgID(kerrors.ErrorUnauthenticated)

	openAPIDefinitions, err := shared.LoadOpenAPISpec(generated.Asset, "connector_mgmt.yaml")
	if err != nil {
		return errors.Wrap(err, "Can't load OpenAPI specification")
	}

	//  /api/connector_mgmt
	apiRouter := mainRouter.PathPrefix("/api/connector_mgmt").Subrouter()

	//  /api/connector_mgmt/v1
	apiV1Router := apiRouter.PathPrefix("/v1").Subrouter()

	//  /api/connector_mgmt/v1/openapi
	apiV1Router.HandleFunc("/openapi", coreHandlers.NewOpenAPIHandler(openAPIDefinitions).Get).Methods(http.MethodGet)

	//  /api/connector_mgmt/v1/errors
	apiV1ErrorsRouter := apiV1Router.PathPrefix("/errors").Subrouter()
	apiV1ErrorsRouter.HandleFunc("", s.ErrorsHandler.List).Methods(http.MethodGet)
	apiV1ErrorsRouter.HandleFunc("/{id}", s.ErrorsHandler.Get).Methods(http.MethodGet)

	v1Collections := []api.CollectionMetadata{}

	//  /api/connector_mgmt/v1/kafka_connector_types
	v1Collections = append(v1Collections, api.CollectionMetadata{
		ID:   "kafka_connector_types",
		Kind: "ConnectorTypeList",
	})
	apiV1ConnectorTypesRouter := apiV1Router.PathPrefix("/kafka_connector_types").Subrouter()
	apiV1ConnectorTypesRouter.HandleFunc("/{connector_type_id}", s.ConnectorTypesHandler.Get).Methods(http.MethodGet)
	apiV1ConnectorTypesRouter.HandleFunc("", s.ConnectorTypesHandler.List).Methods(http.MethodGet)
	apiV1ConnectorTypesRouter.Use(authorizeMiddleware)
	apiV1ConnectorTypesRouter.Use(requireOrgID)

	//  /api/connector_mgmt/v1/kafka_connectors
	v1Collections = append(v1Collections, api.CollectionMetadata{
		ID:   "kafka_connectors",
		Kind: "ConnectorList",
	})

	apiV1ConnectorsRouter := apiV1Router.PathPrefix("/kafka_connectors").Subrouter()
	apiV1ConnectorsRouter.HandleFunc("", s.ConnectorsHandler.Create).Methods(http.MethodPost)
	apiV1ConnectorsRouter.HandleFunc("", s.ConnectorsHandler.List).Methods(http.MethodGet)
	apiV1ConnectorsRouter.HandleFunc("/{connector_id}", s.ConnectorsHandler.Get).Methods(http.MethodGet)
	apiV1ConnectorsRouter.HandleFunc("/{connector_id}", s.ConnectorsHandler.Patch).Methods(http.MethodPatch)
	apiV1ConnectorsRouter.HandleFunc("/{connector_id}", s.ConnectorsHandler.Delete).Methods(http.MethodDelete)
	apiV1ConnectorsRouter.Use(authorizeMiddleware)
	apiV1ConnectorsRouter.Use(requireOrgID)

	//  /api/connector_mgmt/v1/kafka_connector_clusters
	v1Collections = append(v1Collections, api.CollectionMetadata{
		ID:   "kafka_connector_clusters",
		Kind: "ConnectorClusterList",
	})

	apiV1ConnectorClustersRouter := apiV1Router.PathPrefix("/kafka_connector_clusters").Subrouter()
	apiV1ConnectorClustersRouter.HandleFunc("", s.ConnectorClusterHandler.Create).Methods(http.MethodPost)
	apiV1ConnectorClustersRouter.HandleFunc("", s.ConnectorClusterHandler.List).Methods(http.MethodGet)
	apiV1ConnectorClustersRouter.HandleFunc("/{connector_cluster_id}", s.ConnectorClusterHandler.Get).Methods(http.MethodGet)
	apiV1ConnectorClustersRouter.HandleFunc("/{connector_cluster_id}", s.ConnectorClusterHandler.Update).Methods(http.MethodPut)
	apiV1ConnectorClustersRouter.HandleFunc("/{connector_cluster_id}", s.ConnectorClusterHandler.Delete).Methods(http.MethodDelete)
	apiV1ConnectorClustersRouter.HandleFunc("/{connector_cluster_id}/addon_parameters", s.ConnectorClusterHandler.GetAddonParameters).Methods(http.MethodGet)
	apiV1ConnectorClustersRouter.HandleFunc("/{connector_cluster_id}/namespaces", s.ConnectorClusterHandler.GetNamespaces).Methods(http.MethodGet)
	apiV1ConnectorClustersRouter.Use(authorizeMiddleware)
	apiV1ConnectorClustersRouter.Use(requireOrgID)

	//  /api/connector_mgmt/v1/kafka_connector_namespaces
	v1Collections = append(v1Collections, api.CollectionMetadata{
		ID:   "kafka_connector_namespaces",
		Kind: "ConnectorNamespaceList",
	})

	apiV1ConnectorNamespacesRouter := apiV1Router.PathPrefix("/kafka_connector_namespaces").Subrouter()
	apiV1ConnectorNamespacesRouter.HandleFunc("", s.ConnectorNamespaceHandler.List).Methods(http.MethodGet)
	apiV1ConnectorNamespacesRouter.HandleFunc("/eval", s.ConnectorNamespaceHandler.CreateEvaluation).Methods(http.MethodPost)
	apiV1ConnectorNamespacesRouter.HandleFunc("/{connector_namespace_id}", s.ConnectorNamespaceHandler.Get).Methods(http.MethodGet)
	if s.ConnectorsConfig.ConnectorNamespaceLifecycleAPI {
		apiV1ConnectorNamespacesRouter.HandleFunc("", s.ConnectorNamespaceHandler.Create).Methods(http.MethodPost)
		apiV1ConnectorNamespacesRouter.HandleFunc("/{connector_namespace_id}", s.ConnectorNamespaceHandler.Update).Methods(http.MethodPatch)
		apiV1ConnectorNamespacesRouter.HandleFunc("/{connector_namespace_id}", s.ConnectorNamespaceHandler.Delete).Methods(http.MethodDelete)
	} else {
		apiV1ConnectorNamespacesRouter.HandleFunc("", api.SendMethodNotAllowed).Methods(http.MethodPost)
		apiV1ConnectorNamespacesRouter.HandleFunc("/{connector_namespace_id}", api.SendMethodNotAllowed).Methods(http.MethodPatch)
		apiV1ConnectorNamespacesRouter.HandleFunc("/{connector_namespace_id}", api.SendMethodNotAllowed).Methods(http.MethodDelete)
	}
	apiV1ConnectorNamespacesRouter.Use(authorizeMiddleware)
	apiV1ConnectorNamespacesRouter.Use(requireOrgID)

	// This section adds the API's accessed by the connector agent...
	{
		//  /api/connector_mgmt/v1/kafka_connector_clusters/{id}
		agentRouter := apiV1Router.PathPrefix("/agent/kafka_connector_clusters/{connector_cluster_id}").Subrouter()
		agentRouter.HandleFunc("/status", s.ConnectorClusterHandler.UpdateConnectorClusterStatus).Methods(http.MethodPut)
		agentRouter.HandleFunc("/deployments", s.ConnectorClusterHandler.ListDeployments).Methods(http.MethodGet)
		agentRouter.HandleFunc("/deployments/{deployment_id}", s.ConnectorClusterHandler.GetDeployment).Methods(http.MethodGet)
		agentRouter.HandleFunc("/namespaces", s.ConnectorClusterHandler.GetAgentNamespaces).Methods(http.MethodGet)
		agentRouter.HandleFunc("/namespaces/{namespace_id}", s.ConnectorClusterHandler.GetNamespace).Methods(http.MethodGet)
		agentRouter.HandleFunc("/namespaces/{namespace_id}/status", s.ConnectorClusterHandler.UpdateNamespaceStatus).Methods(http.MethodPut)
		agentRouter.HandleFunc("/deployments/{deployment_id}/status", s.ConnectorClusterHandler.UpdateDeploymentStatus).Methods(http.MethodPut)
		auth.UseOperatorAuthorisationMiddleware(agentRouter, s.KeycloakService.GetRealmConfig().ValidIssuerURI, "connector_cluster_id", s.AuthAgentService)
	}

	// This section adds APIs accessed by connector admins
	adminRouter := apiV1Router.PathPrefix("/admin").Subrouter()
	rolesMapping := map[string][]string{
		http.MethodDelete: {auth.ConnectorFleetManagerAdminWriteRole, auth.ConnectorFleetManagerAdminFullRole},
		http.MethodGet:    {auth.ConnectorFleetManagerAdminReadRole, auth.ConnectorFleetManagerAdminWriteRole, auth.ConnectorFleetManagerAdminFullRole},
		http.MethodPost:   {auth.ConnectorFleetManagerAdminWriteRole, auth.ConnectorFleetManagerAdminFullRole},
		http.MethodPut:    {auth.ConnectorFleetManagerAdminWriteRole, auth.ConnectorFleetManagerAdminFullRole},
	}
	adminRouter.Use(auth.NewRequireIssuerMiddleware().RequireIssuer([]string{s.KeycloakService.GetConfig().OSDClusterIDPRealm.ValidIssuerURI}, kerrors.ErrorNotFound))
	adminRouter.Use(auth.NewRolesAuhzMiddleware().RequireRolesForMethods(rolesMapping, kerrors.ErrorNotFound))
	adminRouter.Use(auth.NewAuditLogMiddleware().AuditLog(kerrors.ErrorNotFound))
	adminRouter.HandleFunc("/kafka_connector_clusters", s.ConnectorAdminHandler.ListConnectorClusters).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connector_clusters/{connector_cluster_id}/namespaces", s.ConnectorAdminHandler.GetClusterNamespaces).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connector_clusters/{connector_cluster_id}/connectors", s.ConnectorAdminHandler.GetClusterConnectors).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connector_clusters/{connector_cluster_id}/deployments", s.ConnectorAdminHandler.GetClusterDeployments).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connector_clusters/{connector_cluster_id}/deployments/{deployment_id}", s.ConnectorAdminHandler.GetConnectorDeployment).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connector_clusters/{connector_cluster_id}/upgrades/type", s.ConnectorAdminHandler.GetConnectorUpgradesByType).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connector_clusters/{connector_cluster_id}/upgrades/type", s.ConnectorAdminHandler.UpgradeConnectorsByType).Methods(http.MethodPut)
	adminRouter.HandleFunc("/kafka_connector_clusters/{connector_cluster_id}/upgrades/operator", s.ConnectorAdminHandler.GetConnectorUpgradesByOperator).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connector_clusters/{connector_cluster_id}/upgrades/operator", s.ConnectorAdminHandler.UpgradeConnectorsByOperator).Methods(http.MethodPut)
	adminRouter.HandleFunc("/kafka_connector_namespaces", s.ConnectorAdminHandler.GetConnectorNamespaces).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connector_namespaces", s.ConnectorAdminHandler.CreateConnectorNamespace).Methods(http.MethodPost)
	adminRouter.HandleFunc("/kafka_connector_namespaces/{namespace_id}", s.ConnectorAdminHandler.GetConnectorNamespace).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connector_namespaces/{namespace_id}", s.ConnectorAdminHandler.DeleteConnectorNamespace).Methods(http.MethodDelete)
	adminRouter.HandleFunc("/kafka_connector_namespaces/{namespace_id}/connectors", s.ConnectorAdminHandler.GetNamespaceConnectors).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connector_namespaces/{namespace_id}/deployments", s.ConnectorAdminHandler.GetNamespaceDeployments).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connectors/{connector_id}", s.ConnectorAdminHandler.GetConnector).Methods(http.MethodGet)
	adminRouter.HandleFunc("/kafka_connectors/{connector_id}", s.ConnectorAdminHandler.DeleteConnector).Methods(http.MethodDelete)

	v1Metadata := api.VersionMetadata{
		ID:          "v1",
		Collections: v1Collections,
	}
	apiMetadata := api.Metadata{
		ID: "connector_mgmt",
		Versions: []api.VersionMetadata{
			v1Metadata,
		},
	}

	apiRouter.HandleFunc("", apiMetadata.ServeHTTP).Methods(http.MethodGet)
	apiV1Router.HandleFunc("", v1Metadata.ServeHTTP).Methods(http.MethodGet)

	apiRouter.Use(coreHandlers.MetricsMiddleware)
	apiRouter.Use(db.TransactionMiddleware(s.DB))
	apiRouter.Use(gorillaHandlers.CompressHandler)
	return nil
}
