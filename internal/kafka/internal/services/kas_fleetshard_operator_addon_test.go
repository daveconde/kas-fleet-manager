package services

import (
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/services/sso"
	"testing"

	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/kafka/internal/clusters"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/kafka/internal/clusters/types"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/kafka/internal/config"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/client/keycloak"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/client/ocm"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/server"

	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/api"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/errors"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

func TestAgentOperatorAddon_Provision(t *testing.T) {
	addonId := "test-id"
	type fields struct {
		providerFactory clusters.ProviderFactory
		ssoService      sso.KeycloakService
	}
	tests := []struct {
		name    string
		fields  fields
		result  bool
		wantErr bool
	}{
		{
			name: "provision is finished successfully",
			fields: fields{
				ssoService: &sso.KeycloakServiceMock{
					RegisterKasFleetshardOperatorServiceAccountFunc: func(agentClusterId string) (*api.ServiceAccount, *errors.ServiceError) {
						return &api.ServiceAccount{}, nil
					},
					GetRealmConfigFunc: func() *keycloak.KeycloakRealmConfig {
						return &keycloak.KeycloakRealmConfig{}
					},
				},
				providerFactory: &clusters.ProviderFactoryMock{GetProviderFunc: func(providerType api.ClusterProviderType) (clusters.Provider, error) {
					return &clusters.ProviderMock{
						InstallKasFleetshardFunc: func(clusterSpec *types.ClusterSpec, params []types.Parameter) (bool, error) {
							return false, nil
						},
					}, nil
				}},
			},
			// we can't change the state of AddOnInstallation to be ready as the field is private
			result:  false,
			wantErr: false,
		},
		{
			name: "provision is failed",
			fields: fields{
				ssoService: &sso.KeycloakServiceMock{
					RegisterKasFleetshardOperatorServiceAccountFunc: func(agentClusterId string) (*api.ServiceAccount, *errors.ServiceError) {
						return nil, errors.GeneralError("error")
					},
				},
				providerFactory: &clusters.ProviderFactoryMock{GetProviderFunc: func(providerType api.ClusterProviderType) (clusters.Provider, error) {
					return &clusters.ProviderMock{
						InstallKasFleetshardFunc: func(clusterSpec *types.ClusterSpec, params []types.Parameter) (bool, error) {
							return false, errors.GeneralError("error")
						},
					}, nil
				}},
			},
			result:  false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterTestingT(t)
			agentOperatorAddon := &kasFleetshardOperatorAddon{
				SsoService:          tt.fields.ssoService,
				ProviderFactory:     tt.fields.providerFactory,
				ServerConfig:        &server.ServerConfig{},
				KasFleetShardConfig: &config.KasFleetshardConfig{},
				OCMConfig:           &ocm.OCMConfig{KasFleetshardAddonID: addonId},
				KeycloakConfig: &keycloak.KeycloakConfig{
					KafkaRealm: &keycloak.KeycloakRealmConfig{},
				},
			}
			ready, _, err := agentOperatorAddon.Provision(api.Cluster{
				ClusterID:    "test-cluster-id",
				ProviderType: api.ClusterProviderOCM,
			})
			if err != nil && !tt.wantErr {
				t.Errorf("Provision() error = %v, want = %v", err, tt.wantErr)
			}
			Expect(ready).To(Equal(tt.result))
		})
	}
}

func TestAgentOperatorAddon_RemoveServiceAccount(t *testing.T) {
	type fields struct {
		ssoService sso.KeycloakService
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "receives error during removal of the service account fails when fleetshard operator is turned on",
			fields: fields{
				ssoService: &sso.KeycloakServiceMock{
					DeRegisterKasFleetshardOperatorServiceAccountFunc: func(agentClusterId string) *errors.ServiceError {
						return &errors.ServiceError{} // an error is returned
					},
				},
			},
			wantErr: true,
		},
		{
			name: "succesful removes the service account when fleetshard operator is turned on",
			fields: fields{
				ssoService: &sso.KeycloakServiceMock{
					DeRegisterKasFleetshardOperatorServiceAccountFunc: func(agentClusterId string) *errors.ServiceError {
						return nil
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterTestingT(t)
			agentOperatorAddon := &kasFleetshardOperatorAddon{
				SsoService: tt.fields.ssoService,
			}
			err := agentOperatorAddon.RemoveServiceAccount(api.Cluster{
				ClusterID:    "test-cluster-id",
				ProviderType: api.ClusterProviderOCM,
			})
			gomega.Expect(err != nil).To(Equal(tt.wantErr))
		})
	}
}

func TestKasFleetshardOperatorAddon_ReconcileParameters(t *testing.T) {
	type fields struct {
		providerFactory clusters.ProviderFactory
		ssoService      sso.KeycloakService
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ReconcileParameters is finished successfully",
			fields: fields{
				ssoService: &sso.KeycloakServiceMock{
					RegisterKasFleetshardOperatorServiceAccountFunc: func(agentClusterId string) (*api.ServiceAccount, *errors.ServiceError) {
						return &api.ServiceAccount{}, nil
					},
					GetRealmConfigFunc: func() *keycloak.KeycloakRealmConfig {
						return &keycloak.KeycloakRealmConfig{}
					},
				},
				providerFactory: &clusters.ProviderFactoryMock{GetProviderFunc: func(providerType api.ClusterProviderType) (clusters.Provider, error) {
					return &clusters.ProviderMock{
						InstallKasFleetshardFunc: func(clusterSpec *types.ClusterSpec, params []types.Parameter) (bool, error) {
							return true, nil
						},
					}, nil
				}},
			},
			wantErr: false,
		},
		{
			name: "ReconcileParameters is failed because UpdateAddonParameters failed",
			fields: fields{
				ssoService: &sso.KeycloakServiceMock{
					RegisterKasFleetshardOperatorServiceAccountFunc: func(agentClusterId string) (*api.ServiceAccount, *errors.ServiceError) {
						return &api.ServiceAccount{}, nil
					},
					GetRealmConfigFunc: func() *keycloak.KeycloakRealmConfig {
						return &keycloak.KeycloakRealmConfig{}
					},
				},
				providerFactory: &clusters.ProviderFactoryMock{GetProviderFunc: func(providerType api.ClusterProviderType) (clusters.Provider, error) {
					return &clusters.ProviderMock{
						InstallKasFleetshardFunc: func(clusterSpec *types.ClusterSpec, params []types.Parameter) (bool, error) {
							return false, errors.GeneralError("test error")
						},
					}, nil
				}},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterTestingT(t)
			agentOperatorAddon := &kasFleetshardOperatorAddon{
				SsoService:          tt.fields.ssoService,
				ProviderFactory:     tt.fields.providerFactory,
				ServerConfig:        &server.ServerConfig{},
				KasFleetShardConfig: &config.KasFleetshardConfig{},
				OCMConfig:           &ocm.OCMConfig{KasFleetshardAddonID: "kas-fleetshard"},
				KeycloakConfig: &keycloak.KeycloakConfig{
					KafkaRealm: &keycloak.KeycloakRealmConfig{},
				},
			}
			_, err := agentOperatorAddon.ReconcileParameters(api.Cluster{
				ClusterID:    "test-cluster-id",
				ProviderType: api.ClusterProviderOCM,
			})
			if err != nil && !tt.wantErr {
				t.Errorf("Provision() error = %v, want = %v", err, tt.wantErr)
			}
		})
	}
}
