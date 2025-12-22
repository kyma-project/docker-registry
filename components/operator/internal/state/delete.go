package state

import (
	"context"
	"time"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"github.com/kyma-project/docker-registry/components/operator/internal/resource"
	toolkit_resource "github.com/kyma-project/manager-toolkit/installation/base/resource"
	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/manager-toolkit/installation/chart/action"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// delete dockerregistry based on previously installed resources
func sFnDeleteResources(_ context.Context, _ *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateDeleting)
	s.instance.UpdateConditionUnknown(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeletion,
		"Uninstalling",
	)

	return nextState(sFnSafeDeletionState)
}

func sFnSafeDeletionState(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	if err := chart.CheckCRDOrphanResources(s.chartConfig); err != nil {
		// stop state machine with a warning and requeue reconciliation in 1min
		// warning state indicates that user intervention would fix it. It's not reconciliation error.
		s.setState(v1alpha1.StateWarning)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeDeleted,
			v1alpha1.ConditionReasonDeletionErr,
			err,
		)
		return stopWithEventualError(err)
	}

	return deleteResourcesWithFilter(ctx, r, s)
}

func deleteResourcesWithFilter(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	done, err := chart.Uninstall(s.chartConfig, &chart.UninstallOpts{
		// cleanup secrets created in all namespaces
		PostActions: []action.PostUninstall{
			action.PostUninstallWithPredicate(
				func(u unstructured.Unstructured) (bool, error) {
					return resource.RemoveResourceFromAllNamespaces(ctx, r.client, r.log, u)
				},
				toolkit_resource.AndPredicates(
					toolkit_resource.HasKind("Secret"),
					toolkit_resource.HasLabel(registry.LabelConfigKey, registry.LabelConfigVal),
				),
			),
		},
	})
	if err != nil {
		return uninstallResourcesError(r, s, err)
	}
	if !done {
		return awaitingSecretsRemoval(s)
	}

	s.setState(v1alpha1.StateDeleting)
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeleted,
		"DockerRegistry module deleted",
	)

	// if resources are ready to be deleted, remove finalizer
	return nextState(sFnRemoveFinalizer)
}

func uninstallResourcesError(r *reconciler, s *systemState, err error) (stateFn, *ctrl.Result, error) {
	r.log.Warnf("error while uninstalling resource %s: %s",
		client.ObjectKeyFromObject(&s.instance), err.Error())
	s.setState(v1alpha1.StateError)
	s.instance.UpdateConditionFalse(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeletionErr,
		err,
	)
	return stopWithEventualError(err)
}

func awaitingSecretsRemoval(s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateDeleting)
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeletion,
		"Deleting module resources",
	)

	// wait one sec until ctrl-mngr remove finalizers from secrets
	return requeueAfter(time.Second)
}
