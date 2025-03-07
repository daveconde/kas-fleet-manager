package serviceaccounts

import (
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/auth"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/environments"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/flags"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/services/sso"

	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// NewDeleteCommand command for deleting kafkas.
func NewDeleteCommand(env *environments.Env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a serviceaccount",
		Long:  "Delete a serviceaccount.",
		Run: func(cmd *cobra.Command, args []string) {
			runDelete(env, cmd)
		},
	}

	cmd.Flags().String(FlagSaID, "", "Service Account id")
	cmd.Flags().String(FlagOrgID, "", "OCM org id")
	return cmd
}

func runDelete(env *environments.Env, cmd *cobra.Command) {
	id := flags.MustGetDefinedString(FlagSaID, cmd.Flags())
	orgId := flags.MustGetDefinedString(FlagOrgID, cmd.Flags())

	// setup required services
	keycloakConfig := KeycloakConfig(env)
	keycloakService := sso.NewKeycloakServiceBuilder().
		WithConfiguration(keycloakConfig).
		Build()

	ctx := cmd.Context()
	// create jwt with claims and set it in the context
	jwt := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"org_id": orgId,
	})
	ctx = auth.SetTokenInContext(ctx, jwt)

	err := keycloakService.DeleteServiceAccount(ctx, id)
	if err != nil {
		glog.Fatalf("Unable to delete service account: %s", err.Error())
	}

	glog.V(10).Infof("Deleted service account with id %s", id)
}
