package state

import (
	"context"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// choose right scenario to start (installation/deletion)
func sFnInitialize(_ context.Context, _ *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateProcessing)

	// in case instance is being deleted and has finalizer - delete all resources
	instanceIsBeingDeleted := !s.instance.GetDeletionTimestamp().IsZero()
	if instanceIsBeingDeleted {
		return nextState(sFnDeleteResources)
	}

	return nextState(sFnAccessConfiguration)
}
