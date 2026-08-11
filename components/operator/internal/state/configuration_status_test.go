package state

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/warning"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnUpdateConfigurationStatus(t *testing.T) {
	t.Run("update condition configured", func(t *testing.T) {
		s := &systemState{
			instance:       v1alpha1.DockerRegistry{},
			warningBuilder: warning.NewBuilder(),
		}

		next, result, err := sFnUpdateConfigurationStatus(context.Background(), &reconciler{}, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			"Configuration ready",
		)
	})

	t.Run("report the warnings of the configuration states", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{},
			warningBuilder: warning.NewBuilder().With(
				`failed to set storage configuration: while fetching s3 storage secret from docker-registry: secrets "missing-secret" not found`,
			),
		}

		next, result, err := sFnUpdateConfigurationStatus(context.Background(), &reconciler{}, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonConfigurationErr,
			`Warning: failed to set storage configuration: while fetching s3 storage secret from docker-registry: secrets "missing-secret" not found`,
		)
	})

	t.Run("stop reporting a failure once the configuration succeeds", func(t *testing.T) {
		// the conditions of the previous reconciliation are still on the instance, so a recovered
		// configuration has to overwrite the failure it reported earlier
		s := &systemState{
			instance:       v1alpha1.DockerRegistry{},
			warningBuilder: warning.NewBuilder(),
		}
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			errors.New("secrets \"missing-secret\" not found"),
		)

		_, _, err := sFnUpdateConfigurationStatus(context.Background(), &reconciler{}, s)
		require.NoError(t, err)

		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			"Configuration ready",
		)
	})
}
