package state

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnResourcesConfiguration(_ context.Context, _ *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	if s.instance.Spec.Resources != nil {
		s.flagsBuilder.WithResources(*s.instance.Spec.Resources)
	}

	if s.instance.Spec.Replicas != nil {
		s.flagsBuilder.WithReplicas(*s.instance.Spec.Replicas)
	}

	return nextState(sFnAccessConfiguration)
}
