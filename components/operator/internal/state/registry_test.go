package state

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnRegistryConfiguration(t *testing.T) {
	t.Run("internal registry using default storage", func(t *testing.T) {
		s := &systemState{
			instance:       v1alpha1.DockerRegistry{},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}
		expectedFlags := map[string]interface{}{
			"configData": map[string]interface{}{
				"storage": map[string]interface{}{
					"filesystem": map[string]interface{}{
						"rootdirectory": "/var/lib/registry",
					},
				},
			},
			"storage":          "filesystem",
			"registryNodePort": int64(32_137),
		}

		next, result, err := sFnRegistryConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnControllerConfiguration, next)

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
	})

	t.Run("internal registry using azure storage", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						Azure: &v1alpha1.StorageAzure{
							Secrets: &v1alpha1.StorageAzureSecrets{
								AccountName: "accountName",
								AccountKey:  "accountKey",
								Container:   "container",
							},
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}
		expectedFlags := map[string]interface{}{
			"storage": "azure",
			"secrets": map[string]interface{}{
				"azure": map[string]interface{}{
					"accountName": "accountName",
					"accountKey":  "accountKey",
					"container":   "container",
				},
			},
			"registryNodePort": int64(32_137),
		}

		next, result, err := sFnRegistryConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnControllerConfiguration, next)

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
	})
}
