package state

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"github.com/kyma-project/docker-registry/components/operator/internal/warning"

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
			flagsBuilder:     chart.NewFlagsBuilder(),
			nodePortResolver: registry.NewNodePortResolver(registry.RandomNodePort),
			gatewayHostResolver: &testExternalAddressResolver{expectedAccess: &registry.ResolvedAccess{
				Host: "registry-test-name-test-namespace.cluster.local",
			}},
			warningBuilder: warning.NewBuilder(),
		}

		c := fake.NewClientBuilder().Build()
		eventRecorder := record.NewFakeRecorder(10)
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: eventRecorder}}
		next, result, err := sFnUpdateFinalStatus(context.TODO(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		require.Nil(t, next)

		status := s.instance.Status
		require.Equal(t, "True", status.InternalAccess.Enabled)
		require.Equal(t, registry.InternalAccessSecretName, status.InternalAccess.SecretName)
		require.Equal(t, "localhost:32137", status.InternalAccess.PullAddress)
		require.Equal(t, "dockerregistry.test-namespace.svc.cluster.local:5000", status.InternalAccess.PushAddress)
		require.Equal(t, "True", status.ExternalAccess.Enabled)
		require.Equal(t, registry.ExternalAccessSecretName, status.ExternalAccess.SecretName)
		require.Equal(t, "registry-test-name-test-namespace.cluster.local", status.ExternalAccess.PushAddress)

		require.Equal(t, FilesystemStorageName, status.Storage)

		require.Equal(t, v1alpha1.StateReady, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonInstalled,
			"DockerRegistry installed",
		)
	})

	t.Run("update status additional storage configuration overrides and warning", func(t *testing.T) {
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
			flagsBuilder:        chart.NewFlagsBuilder(),
			nodePortResolver:    registry.NewNodePortResolver(registry.RandomNodePort),
			gatewayHostResolver: &testExternalAddressResolver{expectedError: errors.New("test-error")},
			warningBuilder:      warning.NewBuilder(),
		}

		s.warningBuilder.With("test warning")
		c := fake.NewClientBuilder().Build()
		eventRecorder := record.NewFakeRecorder(10)
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: eventRecorder}}
		next, result, err := sFnUpdateFinalStatus(context.TODO(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		require.Nil(t, next)

		status := s.instance.Status
		require.Equal(t, "True", status.InternalAccess.Enabled)
		require.Equal(t, "localhost:32137", status.InternalAccess.PullAddress)
		require.Equal(t, "dockerregistry.test-namespace.svc.cluster.local:5000", status.InternalAccess.PushAddress)
		require.Equal(t, "False", status.ExternalAccess.Enabled)

		require.Equal(t, AzureStorageName, status.Storage)

		require.Equal(t, v1alpha1.StateWarning, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonInstalled,
			"Warning: test warning",
		)
	})

	t.Run("update status pvc storage configuration", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						PVC: &v1alpha1.StoragePVC{
							Name: "test-pvc",
						},
					},
				},
			},
			flagsBuilder:        chart.NewFlagsBuilder(),
			nodePortResolver:    registry.NewNodePortResolver(registry.RandomNodePort),
			gatewayHostResolver: &testExternalAddressResolver{expectedError: errors.New("test-error")},
			warningBuilder:      warning.NewBuilder(),
		}

		c := fake.NewClientBuilder().Build()
		eventRecorder := record.NewFakeRecorder(10)
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: eventRecorder}}
		next, result, err := sFnUpdateFinalStatus(context.TODO(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		require.Nil(t, next)

		status := s.instance.Status
		require.Equal(t, "True", status.InternalAccess.Enabled)
		require.Equal(t, "localhost:32137", status.InternalAccess.PullAddress)
		require.Equal(t, "dockerregistry.test-namespace.svc.cluster.local:5000", status.InternalAccess.PushAddress)
		require.Equal(t, "False", status.ExternalAccess.Enabled)

		require.Equal(t, PVCStorageName, status.Storage)
		require.Equal(t, "test-pvc", status.PVC)
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
			warningBuilder:   warning.NewBuilder(),
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
				EventRecorder: record.NewFakeRecorder(10),
			},
		}

		next, result, err := sFnUpdateFinalStatus(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		require.Nil(t, next)
		require.Equal(t, v1alpha1.StateReady, s.instance.Status.State)
		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonInstalled,
			"DockerRegistry installed",
		)
	})
}

type testExternalAddressResolver struct {
	expectedAccess *registry.ResolvedAccess
	expectedError  error
}

func (r *testExternalAddressResolver) Do(_ context.Context, _ client.Client, _ v1alpha1.ExternalAccess) (*registry.ResolvedAccess, error) {
	return r.expectedAccess, r.expectedError
}
