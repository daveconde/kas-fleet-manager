package kafka

import (
	"context"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/kafka/internal/services"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/auth"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/environments"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/flags"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// NewDeleteCommand command for deleting kafkas.
func NewDeleteCommand(env *environments.Env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a kafka request",
		Long:  "Delete a kafka request.",
		Run: func(cmd *cobra.Command, args []string) {
			runDelete(env, cmd, args)
		},
	}

	cmd.Flags().String(FlagID, "", "Kafka id")
	cmd.Flags().String(FlagOwner, "test-user", "Username")
	return cmd
}

func runDelete(env *environments.Env, cmd *cobra.Command, _ []string) {
	id := flags.MustGetDefinedString(FlagID, cmd.Flags())
	owner := flags.MustGetDefinedString(FlagOwner, cmd.Flags())
	var kafkaService services.KafkaService
	env.MustResolveAll(&kafkaService)

	// create jwt with claims and set it in the context
	jwt := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"username": owner,
	})
	ctx := auth.SetTokenInContext(context.TODO(), jwt)

	if err := kafkaService.RegisterKafkaDeprovisionJob(ctx, id); err != nil {
		glog.Fatalf("Unable to register the deprovisioning request: %s", err.Error())
	} else {
		glog.V(10).Infof("Deprovisioning request accepted for kafka cluster with id %s", id)
	}
}
