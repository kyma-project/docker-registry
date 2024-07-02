package state

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnUpdateConfigurationStatus(t *testing.T) {
	t.Run("update condition configured", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{},
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
}
