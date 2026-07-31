package state

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testDeletingDockerRegistry = func() v1alpha1.DockerRegistry {
		dockerRegistry := testInstalledDockerRegistry
		dockerRegistry.Status.State = v1alpha1.StateDeleting
		dockerRegistry.Status.Conditions = []metav1.Condition{
			{
				Type:   string(v1alpha1.ConditionTypeDeleted),
				Reason: string(v1alpha1.ConditionReasonDeletion),
				Status: metav1.ConditionUnknown,
			},
		}
		return dockerRegistry
	}()
)

func Test_sFnDeleteResources(t *testing.T) {
	ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}}

	t.Run("update condition", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{},
		}

		next, result, err := sFnDeleteResources(context.Background(), nil, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnSafeDeletionState, next)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateDeleting, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			metav1.ConditionUnknown,
			v1alpha1.ConditionReasonDeletion,
			"Uninstalling",
		)
	})

	t.Run("deletion error while checking orphan resources", func(t *testing.T) {
		s := &systemState{
			instance: *testDeletingDockerRegistry.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixManifestCache("\t"),
				CacheKey: types.NamespacedName{
					Name:      testInstalledDockerRegistry.GetName(),
					Namespace: testInstalledDockerRegistry.GetNamespace(),
				},
			},
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		next, result, err := sFnSafeDeletionState(context.TODO(), r, s)
		require.EqualError(t, err, "could not parse chart manifest: yaml: found character that cannot start any token")
		require.Nil(t, result)
		require.Nil(t, next)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateWarning, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonDeletionErr,
			"could not parse chart manifest: yaml: found character that cannot start any token",
		)
	})

	t.Run("deletion", func(t *testing.T) {
		s := &systemState{
			instance: *testDeletingDockerRegistry.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testDeletingDockerRegistry.GetName(),
					Namespace: testDeletingDockerRegistry.GetNamespace(),
				},
				Cluster: chart.Cluster{
					Client: fake.NewClientBuilder().
						WithScheme(scheme.Scheme).
						WithObjects(&ns).
						Build(),
				},
			},
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		next, result, err := sFnSafeDeletionState(context.TODO(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnRemoveFinalizer, next)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateDeleting, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonDeleted,
			"DockerRegistry module deleted",
		)
	})

	t.Run("config secret finalizer is released while awaiting uninstall", func(t *testing.T) {
		namespace := testDeletingDockerRegistry.GetNamespace()
		configSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:       registry.InternalAccessSecretName,
				Namespace:  namespace,
				Labels:     map[string]string{registry.LabelConfigKey: registry.LabelConfigVal},
				Finalizers: []string{registry.ConfigSecretFinalizer},
			},
		}

		c := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithObjects(&ns, configSecret).
			Build()

		s := &systemState{
			instance: *testDeletingDockerRegistry.DeepCopy(),
			chartConfig: &chart.Config{
				Ctx:   context.TODO(),
				Log:   zap.NewNop().Sugar(),
				Cache: fixManifestCache(fixConfigSecretManifest(namespace)),
				CacheKey: types.NamespacedName{
					Name:      testDeletingDockerRegistry.GetName(),
					Namespace: namespace,
				},
				Cluster: chart.Cluster{Client: c},
			},
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
			k8s: k8s{client: c},
		}

		next, result, err := sFnSafeDeletionState(context.TODO(), r, s)
		require.Nil(t, err)
		require.Nil(t, next)
		require.NotNil(t, result)
		require.Equal(t, time.Second, result.RequeueAfter)

		// the Secret controller is not running here, so the state machine has to release the
		// finalizer itself, otherwise the namespace could never finish terminating
		getErr := c.Get(context.TODO(), client.ObjectKeyFromObject(configSecret), &corev1.Secret{})
		require.True(t, apierrors.IsNotFound(getErr),
			"config secret should be gone once the finalizer is released, got %v", getErr)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateDeleting, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonDeletion,
			"Deleting module resources",
		)
	})
}

func fixConfigSecretManifest(namespace string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
  labels:
    %s: %s
`, registry.InternalAccessSecretName, namespace, registry.LabelConfigKey, registry.LabelConfigVal)
}
