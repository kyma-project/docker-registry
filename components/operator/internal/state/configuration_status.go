package state

import (
	"context"
	"errors"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func sFnUpdateConfigurationStatus(_ context.Context, _ *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	// the configuration states record a warning for everything that did not end up configured the way
	// the CR asks for, so reporting the configuration as ready while any of them is set would hide it
	if warning := s.warningBuilder.Build(); warning != "" {
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			errors.New(warning),
		)

		return nextState(sFnApplyResources)
	}

	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeConfigured,
		v1alpha1.ConditionReasonConfigured,
		"Configuration ready",
	)

	return nextState(sFnApplyResources)
}
