package state

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/internal/registry"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnConfigurationStatus(t *testing.T) {
	configurationReadyMsg := "Configuration ready"

	t.Run("update status additional configuration overrides", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-name",
					Namespace: "test-namespace",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					ExternalAccess: &v1alpha1.ExternalAccess{
						Enabled: ptr.To(true),
					},
				},
			},
			flagsBuilder:            chart.NewFlagsBuilder(),
			nodePortResolver:        registry.NewNodePortResolver(registry.RandomNodePort),
			externalAddressResolver: &testExternalAddressResolver{expectedAddress: "registry-test-name-test-namespace.cluster.local"},
		}

		c := fake.NewClientBuilder().Build()
		eventRecorder := record.NewFakeRecorder(10)
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: eventRecorder}}
		next, result, err := sFnConfigurationStatus(context.TODO(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		status := s.instance.Status
		require.Equal(t, registry.SecretName, status.InternalAccess.SecretName)
		require.Equal(t, "localhost:32137", status.PullAddress)
		require.Equal(t, "dockerregistry.test-namespace.svc.cluster.local:5000", status.InternalAccess.PushAddress)
		require.Equal(t, registry.SecretName, status.ExternalAccess.SecretName)
		require.Equal(t, "registry-test-name-test-namespace.cluster.local", status.ExternalAccess.PushAddress)

		require.Equal(t, FilesystemStorageName, status.Storage)

		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			configurationReadyMsg,
		)
	})

	t.Run("update status additional storage configuration overrides", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						Azure: &v1alpha1.StorageAzure{
							SecretName: "azureSecret",
						},
					},
				},
			},
			flagsBuilder:            chart.NewFlagsBuilder(),
			nodePortResolver:        registry.NewNodePortResolver(registry.RandomNodePort),
			externalAddressResolver: &testExternalAddressResolver{expectedAddress: "registry-test-name-test-namespace.cluster.local"},
		}

		c := fake.NewClientBuilder().Build()
		eventRecorder := record.NewFakeRecorder(10)
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: eventRecorder}}
		next, result, err := sFnConfigurationStatus(context.TODO(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		status := s.instance.Status
		require.Equal(t, "localhost:32137", status.PullAddress)
		require.Equal(t, "dockerregistry.test-namespace.svc.cluster.local:5000", status.InternalAccess.PushAddress)

		require.Equal(t, AzureStorageName, status.Storage)

		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			configurationReadyMsg,
		)

	})

	t.Run("reconcile from configurationError", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
				},
				Status: v1alpha1.DockerRegistryStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(v1alpha1.ConditionTypeConfigured),
							Status: metav1.ConditionFalse,
							Reason: string(v1alpha1.ConditionReasonConfigurationErr),
						},
						{
							Type:   string(v1alpha1.ConditionTypeInstalled),
							Status: metav1.ConditionTrue,
							Reason: string(v1alpha1.ConditionReasonInstallation),
						},
					},
					State: v1alpha1.StateError,
				},
			},
			statusSnapshot:   v1alpha1.DockerRegistryStatus{},
			flagsBuilder:     chart.NewFlagsBuilder(),
			nodePortResolver: registry.NewNodePortResolver(registry.RandomNodePort),
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "boo",
			},
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
			k8s: k8s{
				client:        fake.NewClientBuilder().WithObjects(secret).Build(),
				EventRecorder: record.NewFakeRecorder(4),
			},
		}

		next, result, err := sFnConfigurationStatus(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)
		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			configurationReadyMsg)
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
	})
}

type testExternalAddressResolver struct {
	expectedAddress string
	expectedError   error
}

func (r *testExternalAddressResolver) GetExternalAddress(_ context.Context, _ client.Client, _ string) (string, error) {
	return r.expectedAddress, r.expectedError
}
