package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"reflect"

	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/connector/internal/api/dbapi"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/connector/internal/api/public"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/connector/internal/config"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/connector/internal/presenters"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/connector/internal/services/phase"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/api"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/db"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/errors"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/services"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/services/queryparser"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/services/signalbus"
	"gorm.io/gorm"
)

type ConnectorNamespaceService interface {
	Create(ctx context.Context, request *dbapi.ConnectorNamespace) *errors.ServiceError
	Update(ctx context.Context, request *dbapi.ConnectorNamespace) *errors.ServiceError
	Get(ctx context.Context, namespaceID string) (*dbapi.ConnectorNamespace, *errors.ServiceError)
	List(ctx context.Context, clusterIDs []string, listArguments *services.ListArguments, gtVersion int64) (dbapi.ConnectorNamespaceList, *api.PagingMeta, *errors.ServiceError)
	Delete(ctx context.Context, namespaceId string) *errors.ServiceError
	SetEvalClusterId(request *dbapi.ConnectorNamespace) *errors.ServiceError
	CreateDefaultNamespace(ctx context.Context, connectorCluster *dbapi.ConnectorCluster) *errors.ServiceError
	GetExpiredNamespaceIds() ([]string, *errors.ServiceError)
	UpdateConnectorNamespaceStatus(ctx context.Context, namespaceID string, status *dbapi.ConnectorNamespaceStatus) *errors.ServiceError
	DeleteNamespaceAndConnectorDeployments(ctx context.Context, dbConn *gorm.DB, query interface{}, values ...interface{}) (bool, bool, *errors.ServiceError)
	ReconcileUnusedDeletingNamespaces() (int, []*errors.ServiceError)
	ReconcileDeletedNamespaces() (int, []*errors.ServiceError)
	GetNamespaceTenant(namespaceId string) (*dbapi.ConnectorNamespace, *errors.ServiceError)
	CheckConnectorQuota(namespaceId string) *errors.ServiceError
	CanCreateEvalNamespace(userId string) *errors.ServiceError
	GetEmptyDeletingNamespaces(clusterId string) (dbapi.ConnectorNamespaceList, *errors.ServiceError)
}

var _ ConnectorNamespaceService = &connectorNamespaceService{}

type connectorNamespaceService struct {
	connectionFactory *db.ConnectionFactory
	connectorsConfig  *config.ConnectorsConfig
	quotaConfig       *config.ConnectorsQuotaConfig
	bus               signalbus.SignalBus
}

func init() {
	// random seed
	rand.Seed(time.Now().UnixNano())
}

func NewConnectorNamespaceService(factory *db.ConnectionFactory, config *config.ConnectorsConfig,
	quotaConfig *config.ConnectorsQuotaConfig, bus signalbus.SignalBus) *connectorNamespaceService {
	return &connectorNamespaceService{
		connectionFactory: factory,
		connectorsConfig:  config,
		quotaConfig:       quotaConfig,
		bus:               bus,
	}
}

func (k *connectorNamespaceService) SetEvalClusterId(request *dbapi.ConnectorNamespace) *errors.ServiceError {
	var availableClusters []string

	// get eval clusters
	dbConn := k.connectionFactory.New()
	if err := dbConn.Raw("SELECT id FROM connector_clusters WHERE deleted_at is null AND organisation_id IN ?",
		k.connectorsConfig.ConnectorEvalOrganizations).
		Scan(&availableClusters).Error; err != nil {
		return errors.Unauthorized("failed to get eval cluster id: %v", err)
	}

	numOrgClusters := len(availableClusters)
	if numOrgClusters == 0 {
		return errors.Unauthorized("no eval clusters")
	} else if numOrgClusters == 1 {
		request.ClusterId = availableClusters[0]
	} else {
		// TODO add support for load balancing strategies
		// pick a cluster at random
		request.ClusterId = availableClusters[rand.Intn(numOrgClusters)]
	}

	// also set expiration duration
	expiry := time.Now().Add(k.connectorsConfig.ConnectorEvalDuration)
	request.Expiration = &expiry

	// also set profile in annotations to ConnectorsQuotaConfig.EvalNamespaceQuotaProfile
	found := false
	for i := 0; i < len(request.Annotations); i++ {
		ann := &request.Annotations[i]
		if ann.Key == config.AnnotationProfileKey {
			if ann.Value != k.quotaConfig.EvalNamespaceQuotaProfile {
				ann.Value = k.quotaConfig.EvalNamespaceQuotaProfile
			}
			found = true
			break
		}
	}
	if !found {
		request.Annotations = append(request.Annotations, dbapi.ConnectorNamespaceAnnotation{
			NamespaceId: request.ID,
			Key:         config.AnnotationProfileKey,
			Value:       k.quotaConfig.EvalNamespaceQuotaProfile,
		})
	}

	return nil
}

func (k *connectorNamespaceService) Create(ctx context.Context, request *dbapi.ConnectorNamespace) *errors.ServiceError {

	if err := k.validateAnnotations(request); err != nil {
		return err
	}

	dbConn := k.connectionFactory.New()
	if err := dbConn.Create(request).Error; err != nil {
		return services.HandleCreateError("Connector namespace", err)
	}

	// reload namespace to get version
	if err := dbConn.Preload("Annotations").Preload("TenantUser").Preload("TenantOrganisation").
		First(request, "id = ?", request.ID).Error; err != nil {
		return services.HandleGetError("Connector namespace", "id", request.ID, err)
	}

	return nil
}

func (k *connectorNamespaceService) validateAnnotations(request *dbapi.ConnectorNamespace) *errors.ServiceError {
	for _, a := range request.Annotations {
		if a.Key == config.AnnotationProfileKey {
			if _, ok := k.quotaConfig.GetNamespaceQuota(a.Value); !ok {
				return errors.BadRequest(`invalid profile %s`, a.Value)
			}
		}
	}
	return nil
}

func (k *connectorNamespaceService) Update(ctx context.Context, request *dbapi.ConnectorNamespace) *errors.ServiceError {

	if request.Version == 0 {
		return errors.BadRequest("resource version is required")
	}

	dbConn := k.connectionFactory.New()
	updates := dbConn.Where(`id = ? AND version = ?`, request.ID, request.Version).
		Updates(request)
	if err := updates.Error; err != nil {
		return services.HandleUpdateError(`Connector namespace`, err)
	}
	if updates.RowsAffected == 0 {
		return errors.Conflict(`resource version changed`)
	}

	// reload namespace to get version update
	if err := dbConn.Select(`version`).First(request).Error; err != nil {
		return services.HandleGetError(`Connector namespace`, `id`, request.ID, err)
	}

	return nil
}

func (k *connectorNamespaceService) Get(ctx context.Context, namespaceID string) (*dbapi.ConnectorNamespace, *errors.ServiceError) {
	dbConn := k.connectionFactory.New()
	result := &dbapi.ConnectorNamespace{
		Model: db.Model{
			ID: namespaceID,
		},
	}
	if err := dbConn.Preload("Annotations").Preload("TenantUser").Preload("TenantOrganisation").
		Unscoped().First(result).Error; err != nil {
		return nil, services.HandleGetError(`Connector namespace`, `id`, namespaceID, err)
	}
	if result.DeletedAt.Valid {
		return nil, services.HandleGoneError("Connector namespace", "id", namespaceID)
	}
	if err := k.setConnectorsDeployed(dbapi.ConnectorNamespaceList{result}); err != nil {
		return result, err
	}

	return result, nil
}

func (k *connectorNamespaceService) setConnectorsDeployed(namespaces dbapi.ConnectorNamespaceList) *errors.ServiceError {
	if len(namespaces) == 0 {
		return nil
	}

	idMap := make(map[string]*dbapi.ConnectorNamespace)
	ids := make([]string, len(namespaces))
	i := 0
	for _, ns := range namespaces {
		idMap[ns.ID] = ns
		ids[i] = ns.ID
		i++
	}

	result := make([]struct {
		Id    string
		Count int32
	}, 0)
	if err := k.connectionFactory.New().Model(&dbapi.ConnectorDeployment{}).
		Select("namespace_id as id, count(*) as count").
		Group("namespace_id").
		Having("namespace_id in ?", ids).
		Find(&result).Error; err != nil {
		return services.HandleGetError(`Connector namespace`, `id`, ids, err)
	}

	// set counts for non-empty ns
	for _, row := range result {
		ns := idMap[row.Id]
		ns.Status.ConnectorsDeployed = row.Count
		delete(idMap, ns.ID)
	}
	// set remaining ns to 0
	for _, ns := range idMap {
		ns.Status.ConnectorsDeployed = 0
	}

	return nil
}

var validNamespaceColumns = []string{`name`, `cluster_id`, `owner`, `expiration`, `tenant_user_id`, `tenant_organisation_id`, `status_phase`}

func (k *connectorNamespaceService) List(ctx context.Context, clusterIDs []string, listArguments *services.ListArguments, gtVersion int64) (dbapi.ConnectorNamespaceList, *api.PagingMeta, *errors.ServiceError) {
	var resourceList dbapi.ConnectorNamespaceList
	pagingMeta := api.PagingMeta{
		Page:  listArguments.Page,
		Size:  listArguments.Size,
		Total: 0,
	}
	dbConn := k.connectionFactory.New().Model(&resourceList)
	if len(clusterIDs) != 0 {
		dbConn = dbConn.Where("cluster_id IN ?", clusterIDs)
	}

	// Apply search query
	if len(listArguments.Search) > 0 {
		queryParser := queryparser.NewQueryParser(validNamespaceColumns...)
		searchDbQuery, err := queryParser.Parse(listArguments.Search)
		if err != nil {
			return resourceList, &pagingMeta, errors.NewWithCause(errors.ErrorFailedToParseSearch, err, "Unable to list connector namespace requests: %s", err.Error())
		}
		dbConn = dbConn.Where(searchDbQuery.Query, searchDbQuery.Values...)
	}

	// check if a minimum resource version is provided
	if gtVersion != 0 {
		dbConn = dbConn.Where("connector_namespaces.version > ?", gtVersion)
	}

	// set total, limit and paging (based on https://gitlab.cee.redhat.com/service/api-guidelines#user-content-paging)
	total := int64(pagingMeta.Total)
	dbConn.Count(&total)
	pagingMeta.Total = int(total)
	if pagingMeta.Size > pagingMeta.Total {
		pagingMeta.Size = pagingMeta.Total
	}
	dbConn = dbConn.Offset((pagingMeta.Page - 1) * pagingMeta.Size).Limit(pagingMeta.Size)

	if len(listArguments.OrderBy) == 0 {
		// default orderBy name
		dbConn = dbConn.Order("name ASC")
	}

	// Set the order by arguments if any
	for _, orderByArg := range listArguments.OrderBy {
		dbConn = dbConn.Order(orderByArg)
	}

	// execute query
	if err := dbConn.Preload("Annotations").Preload("TenantUser").Preload("TenantOrganisation").
		Find(&resourceList).Error; err != nil {
		return nil, nil, errors.GeneralError("failed to get connector namespaces: %v", err)
	}

	if err := k.setConnectorsDeployed(resourceList); err != nil {
		return resourceList, &pagingMeta, err
	}
	return resourceList, &pagingMeta, nil
}

func (k *connectorNamespaceService) Delete(ctx context.Context, namespaceId string) *errors.ServiceError {

	var namespaceDeleted, connectorsDeleted bool
	if err := k.connectionFactory.New().Transaction(func(dbConn *gorm.DB) error {

		var resource dbapi.ConnectorNamespace
		if err := dbConn.Where("id = ?", namespaceId).Select("id", "cluster_id", "status_phase").
			First(&resource).Error; err != nil {
			return services.HandleGetError("Connector namespace", "id", namespaceId, err)
		}

		var cluster dbapi.ConnectorCluster
		if err := dbConn.Where("id = ?", resource.ClusterId).Select("id", "status_phase").
			First(&cluster).Error; err != nil {
			return services.HandleGetError("Connector namespace", "id", namespaceId, err)
		}

		if _, err := phase.PerformNamespaceOperation(&cluster, &resource, phase.DeleteNamespace,
			func(ns *dbapi.ConnectorNamespace) *errors.ServiceError {

				var serr *errors.ServiceError
				namespaceDeleted, connectorsDeleted, serr = k.DeleteNamespaceAndConnectorDeployments(ctx, dbConn, "connector_namespaces.id = ?", namespaceId)
				if !namespaceDeleted {
					return errors.NotFound("Connector namespace %s not found", namespaceId)
				}
				return serr

			}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return services.HandleDeleteError("Connector namespace", "id", namespaceId, err)
	}

	// reconcile deleted namespaces and connectors
	defer func() {
		if namespaceDeleted {
			k.bus.Notify("reconcile:connector_namespace")
		}
	}()
	defer func() {
		if connectorsDeleted {
			k.bus.Notify("reconcile:connector")
		}
	}()

	return nil
}

// TODO make this a configurable property in the future
const defaultNamespaceName = "default-connector-namespace"

func (k *connectorNamespaceService) CreateDefaultNamespace(ctx context.Context, connectorCluster *dbapi.ConnectorCluster) *errors.ServiceError {
	owner := connectorCluster.Owner
	organisationId := connectorCluster.OrganisationId

	kind := public.CONNECTORNAMESPACETENANTKIND_USER
	if organisationId != "" {
		kind = public.CONNECTORNAMESPACETENANTKIND_ORGANISATION
	}

	namespaceRequest, err := presenters.ConvertConnectorNamespaceRequest(&public.ConnectorNamespaceRequest{
		Name: defaultNamespaceName,
		Annotations: []public.ConnectorNamespaceRequestMetaAnnotations{
			{
				Key:   config.AnnotationProfileKey,
				Value: "default-profile",
			},
		},
		ClusterId: connectorCluster.ID,
		Kind:      kind,
	}, owner, organisationId)

	if err != nil {
		return err
	}
	return k.Create(ctx, namespaceRequest)
}

func (k *connectorNamespaceService) GetExpiredNamespaceIds() ([]string, *errors.ServiceError) {
	var result []string
	dbConn := k.connectionFactory.New()
	now := time.Now()
	if err := dbConn.Model(&dbapi.ConnectorNamespace{}).Select("id").
		Where("expiration < ? AND status_phase NOT IN ?", now, []string{string(dbapi.ConnectorNamespacePhaseDeleting), string(dbapi.ConnectorNamespacePhaseDeleted)}).Find(&result).Error; err != nil {
		return nil, services.HandleGetError("Connector namespace", "expiration", now, err)
	}
	// update phase to 'deleting' for all expired namespaces
	if err := dbConn.Model(&dbapi.ConnectorNamespace{}).
		Where("id in ?", result).
		Update("status_phase", dbapi.ConnectorNamespacePhaseDeleting).Error; err != nil {
		return nil, services.HandleUpdateError("Connector namespace", err)
	}
	return result, nil
}

func (k *connectorNamespaceService) UpdateConnectorNamespaceStatus(ctx context.Context, namespaceID string, status *dbapi.ConnectorNamespaceStatus) *errors.ServiceError {

	if err := k.connectionFactory.New().Transaction(func(dbConn *gorm.DB) error {
		var namespace dbapi.ConnectorNamespace
		if err := dbConn.Unscoped().Where(`id = ?`, namespaceID).First(&namespace).Error; err != nil {
			return services.HandleGetError("Connector namespace", "id", namespaceID, err)
		}
		if namespace.DeletedAt.Valid {
			return services.HandleGoneError(`Connector namespace`, `id`, namespaceID)
		}
		var cluster dbapi.ConnectorCluster
		if err := dbConn.Select(`id`, `status_phase`).Where(`id = ?`, namespace.ClusterId).
			First(&cluster).Error; err != nil {
			return services.HandleGetError("Connector namespace", "id", namespaceID, err)
		}

		updated, serr := phase.PerformNamespaceOperation(&cluster, &namespace, phase.ConnectNamespace)
		if serr != nil {
			return serr
		}

		// take new phase from fsm operation above
		status.Phase = namespace.Status.Phase
		if updated || !reflect.DeepEqual(namespace.Status, *status) {
			namespace.Status.Version = status.Version
			namespace.Status.ConnectorsDeployed = status.ConnectorsDeployed
			namespace.Status.Conditions = status.Conditions
			if err := k.Update(ctx, &namespace); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return services.HandleUpdateError("Connector namespace", err)
	}

	return nil
}

func (k *connectorNamespaceService) DeleteNamespaceAndConnectorDeployments(ctx context.Context, dbConn *gorm.DB, query interface{}, values ...interface{}) (bool, bool, *errors.ServiceError) {
	var namespaces []dbapi.ConnectorNamespace

	if err := dbConn.Model(&dbapi.ConnectorNamespace{}).Select("id", "cluster_id").
		Where(query, values).Find(&namespaces).Error; err != nil {
		return false, false, services.HandleGetError("Connector namespace", fmt.Sprintf("%s", query), values, err)
	}

	// no namespaces
	if len(namespaces) == 0 {
		return false, false, nil
	}
	var namespaceIds []string
	clusterIds := make(map[string]struct{})
	var empty struct{}
	for _, ns := range namespaces {
		namespaceIds = append(namespaceIds, ns.ID)
		clusterIds[ns.ClusterId] = empty
	}

	// Set namespace phase to "deleting"
	if err := dbConn.Model(&dbapi.ConnectorNamespace{}).
		Where("id IN ? AND status_phase NOT IN ?", namespaceIds,
			[]string{string(dbapi.ConnectorNamespacePhaseDeleting), string(dbapi.ConnectorNamespacePhaseDeleted)}).
		Update("status_phase", dbapi.ConnectorNamespacePhaseDeleting).Error; err != nil {
		return false, false, services.HandleUpdateError("Connector namespace", err)
	}

	// get all connectors that are currently not being deleted
	var connectorIds []string
	if err := dbConn.Model(&dbapi.Connector{}).Select("id").
		Where("namespace_id IN ? AND desired_state != ?", namespaceIds, dbapi.ConnectorDeleted).
		Find(&connectorIds).Error; err != nil {
		return false, false, services.HandleGetError("Connector", "namespace_id", namespaces, err)
	}

	// no connectors
	if len(connectorIds) == 0 {
		return true, false, nil
	}

	// set connectors' desired state to 'deleted' by default
	connectorDesiredState := dbapi.ConnectorDeleted
	if k.connectorsConfig.ConnectorEnableUnassignedConnectors {
		// set connectors' state to 'unassigned' if it's supported, i.e. cascade delete is disabled
		connectorDesiredState = dbapi.ConnectorUnassigned
	}
	// set connector desired state to connectorDesiredState and status to "deleting" to remove from namespaces
	if err := dbConn.Where("deleted_at IS NULL AND id IN ?", connectorIds).
		Updates(&dbapi.Connector{DesiredState: connectorDesiredState}).Error; err != nil {
		return false, false, services.HandleUpdateError("Connector", err)
	}
	if err := dbConn.Where("deleted_at IS NULL AND id IN ?", connectorIds).
		Updates(&dbapi.ConnectorStatus{Phase: dbapi.ConnectorStatusPhaseDeleting}).Error; err != nil {
		return false, false, services.HandleUpdateError("Connector", err)
	}
	// mark all deployment statuses as "deleting"
	if err := dbConn.Where("deleted_at IS NULL AND id IN ?", connectorIds).
		Updates(&dbapi.ConnectorDeploymentStatus{Phase: dbapi.ConnectorStatusPhaseDeleting}).Error; err != nil {
		return false, false, services.HandleUpdateError("Connector", err)
	}

	// notify deployment status update
	_ = db.AddPostCommitAction(ctx, func() {
		for cid := range clusterIds {
			k.bus.Notify(fmt.Sprintf("/kafka_connector_clusters/%s/deployments", cid))
		}
	})

	return true, true, nil
}

func (k *connectorNamespaceService) ReconcileUnusedDeletingNamespaces() (int, []*errors.ServiceError) {
	count := 0
	var errs []*errors.ServiceError
	if err := k.connectionFactory.New().Transaction(func(dbConn *gorm.DB) error {

		// get ids of all namespaces in deleting phase that never had connectors
		var namespaceIds []string
		if err := dbConn.Table("connector_namespaces").Select("connector_namespaces.id").
			Joins("LEFT JOIN connector_statuses ON connector_statuses.namespace_id = connector_namespaces.id").
			Group("connector_namespaces.id").
			Having("connector_namespaces.status_phase = ? AND "+
				"connector_namespaces.deleted_at IS NULL AND count(namespace_id) = 0", dbapi.ConnectorNamespacePhaseDeleting).
			Find(&namespaceIds).Error; err != nil {
			return services.HandleGetError("Connector namespace",
				"status_phase", dbapi.ConnectorNamespacePhaseDeleting, err)
		}
		count = len(namespaceIds)
		if count == 0 {
			return nil
		}

		// mark empty unused namespaces as deleted
		if err := dbConn.Where("id IN ?", namespaceIds).
			Updates(&dbapi.ConnectorNamespace{
				Status: dbapi.ConnectorNamespaceStatus{
					Phase: dbapi.ConnectorNamespacePhaseDeleted,
				},
			}).Error; err != nil {
			count = 0
			return services.HandleUpdateError("Connector namespace", err)
		}

		return nil

	}); err != nil {
		errs = append(errs, services.HandleUpdateError("Connector namespace", err))
	}

	return count, errs
}

// ReconcileDeletedNamespaces deletes empty namespaces in phase deleted with no connectors,
func (k *connectorNamespaceService) ReconcileDeletedNamespaces() (int, []*errors.ServiceError) {
	count := 0
	var errs []*errors.ServiceError
	if err := k.connectionFactory.New().Transaction(func(dbConn *gorm.DB) error {

		// get ids of all namespaces in deleted phase that have no connectors
		var namespaceIds []string
		if err := dbConn.Table("connector_namespaces").Select("connector_namespaces.id").
			Joins("LEFT JOIN connector_statuses ON connector_statuses.namespace_id = connector_namespaces.id AND "+
				"connector_statuses.deleted_at IS NULL").
			Where("connector_namespaces.status_phase = ? AND "+
				"connector_namespaces.deleted_at IS NULL AND namespace_id IS NULL", dbapi.ConnectorNamespacePhaseDeleted).
			Find(&namespaceIds).Error; err != nil {
			return services.HandleGetError("Connector namespace",
				"status_phase", dbapi.ConnectorNamespacePhaseDeleted, err)
		}
		count = len(namespaceIds)
		if count == 0 {
			return nil
		}

		// delete all the empty deleted namespaces
		if err := dbConn.Where("id IN ?", namespaceIds).
			Delete(&dbapi.ConnectorNamespace{}).Error; err != nil {
			count = 0
			return services.HandleDeleteError("Connector namespace", "id", namespaceIds, err)
		}

		return nil

	}); err != nil {
		errs = append(errs, services.HandleDeleteError("Connector namespace", "status_phase",
			dbapi.ConnectorNamespacePhaseDeleted, err))
	}

	return count, errs
}

func (k *connectorNamespaceService) GetNamespaceTenant(namespaceId string) (*dbapi.ConnectorNamespace, *errors.ServiceError) {
	dbConn := k.connectionFactory.New()
	var namespace dbapi.ConnectorNamespace
	if err := dbConn.Unscoped().Where("id = ?", namespaceId).
		Select(`id`, `tenant_user_id`, `tenant_organisation_id`).First(&namespace).Error; err != nil {
		return nil, services.HandleGetError("Connector namespace", "id", namespaceId, err)
	}
	if namespace.DeletedAt.Valid {
		return nil, services.HandleGoneError("Connector cluster", "id", namespaceId)
	}
	return &namespace, nil
}

func (k *connectorNamespaceService) CheckConnectorQuota(namespaceId string) *errors.ServiceError {
	dbConn := k.connectionFactory.New()
	var profileName string
	var quota config.NamespaceQuota
	if err := dbConn.Model(&dbapi.ConnectorNamespaceAnnotation{}).
		Where("namespace_id = ? AND key = ?", namespaceId, config.AnnotationProfileKey).
		Select(`value`).First(&profileName).Error; err != nil {
		return errors.FailedToCheckQuota("Error reading Connector namespace annotation with namespace id %s: %s", namespaceId, err)
	}
	quota, _ = k.quotaConfig.GetNamespaceQuota(profileName)
	if quota.Connectors > 0 {
		// get number of connectors using this namespace
		var count int64
		if err := dbConn.Model(&dbapi.Connector{}).Where("namespace_id = ?", namespaceId).
			Count(&count).Error; err != nil {
			return services.HandleGetError("Connector", "namespace_id", namespaceId, err)
		}
		if count >= int64(quota.Connectors) {
			return errors.InsufficientQuotaError("The maximum number of allowed connectors has been reached")
		}
	}
	return nil
}

func (k *connectorNamespaceService) CanCreateEvalNamespace(userId string) *errors.ServiceError {
	dbConn := k.connectionFactory.New()
	var count int64
	if err := dbConn.Table("connector_namespaces").
		Where("tenant_user_id = ? AND expiration IS NOT NULL "+
			"AND connector_namespaces.deleted_at IS NULL AND connector_clusters.deleted_at IS NULL", userId).
		Joins("JOIN connector_clusters ON connector_clusters.id = connector_namespaces.cluster_id AND connector_clusters.organisation_id IN ?",
			k.connectorsConfig.ConnectorEvalOrganizations).
		Count(&count).Error; err != nil {
		return errors.FailedToCheckQuota("Error reading connector namespace with tenant user id %s: %s", userId, err)
	}

	if count > 0 {
		return errors.InsufficientQuotaError("Evaluation Connector Namespace already exists for user %s", userId)
	}
	return nil
}

func (k *connectorNamespaceService) GetEmptyDeletingNamespaces(clusterId string) (dbapi.ConnectorNamespaceList, *errors.ServiceError) {

	var namespaces dbapi.ConnectorNamespaceList
	dbConn := k.connectionFactory.New()
	if err := dbConn.Table("connector_namespaces").Select("connector_namespaces.id", "connector_namespaces.version").
		Joins("LEFT JOIN connector_statuses ON connector_statuses.namespace_id = connector_namespaces.id AND "+
			"connector_statuses.deleted_at IS NULL").
		Where("connector_namespaces.cluster_id = ? AND connector_namespaces.status_phase = ? AND "+
			"connector_namespaces.deleted_at IS NULL AND namespace_id IS NULL",
			clusterId, dbapi.ConnectorNamespacePhaseDeleting).
		Find(&namespaces).Error; err != nil {
		return namespaces, services.HandleGetError("Connector namespace",
			"cluster_id", clusterId, err)
	}

	return namespaces, nil
}
