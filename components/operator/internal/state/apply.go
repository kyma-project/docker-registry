package state

import (
	"context"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"github.com/kyma-project/manager-toolkit/installation/base/resource"
	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/manager-toolkit/installation/chart/action"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// run dockerregistry chart installation
func sFnApplyResources(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	// set condition Installed if it does not exist
	if !s.instance.IsCondition(v1alpha1.ConditionTypeInstalled) {
		s.instance.UpdateConditionUnknown(v1alpha1.ConditionTypeInstalled, v1alpha1.ConditionReasonInstallation,
			"Installing for configuration")
	}

	s.flagsBuilder.WithManagedByLabel("dockerregistry-operator")

	// install component
	err := install(ctx, r, s)
	if err != nil {
		r.log.Warnf("error while installing resource %s: %s",
			client.ObjectKeyFromObject(&s.instance), err.Error())
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstallationErr,
			err,
		)
		return stopWithEventualError(err)
	}

	// switch state verify
	return nextState(sFnVerifyResources)
}

func install(ctx context.Context, r *reconciler, s *systemState) error {
	flags, err := s.flagsBuilder.Build()
	if err != nil {
		return err
	}

	return chart.Install(s.chartConfig, &chart.InstallOpts{
		CustomFlags: flags,
		PreActions: []action.PreApply{
			action.PreApplyWithPredicate(
				adjustPVCPreApplyAction(ctx, r.client),
				resource.HasKind("PersistentVolumeClaim"),
			),
		},
	})
}

func adjustPVCPreApplyAction(ctx context.Context, c client.Client) action.PreApply {
	return func(u *unstructured.Unstructured) error {
		adjusted, err := registry.AdjustDockerRegToClusterPVCSize(ctx, c, *u)
		*u = adjusted
		return err
	}
}
