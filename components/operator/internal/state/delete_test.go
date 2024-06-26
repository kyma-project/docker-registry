package state

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
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
}
