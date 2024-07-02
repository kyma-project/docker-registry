package state

import (
	"context"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func sFnUpdateConfigurationStatus(_ context.Context, _ *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeConfigured,
		v1alpha1.ConditionReasonConfigured,
		"Configuration ready",
	)

	return nextState(sFnApplyResources)
}
