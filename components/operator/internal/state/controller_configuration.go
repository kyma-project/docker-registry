package state

import (
	"context"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

const (
	AzureStorageName      = "azure"
	FilesystemStorageName = "filesystem"
)

func sFnControllerConfiguration(_ context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	err := updateControllerConfigurationStatus(r, &s.instance)
	if err != nil {
		return stopWithEventualError(err)
	}

	configureControllerConfigurationFlags(s)

	s.setState(v1alpha1.StateProcessing)
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeConfigured,
		v1alpha1.ConditionReasonConfigured,
		"Configuration ready",
	)

	return nextState(sFnApplyResources)
}

func updateControllerConfigurationStatus(r *reconciler, instance *v1alpha1.DockerRegistry) error {
	spec := instance.Spec
	storageField := getStorageField(spec.Storage, instance)
	fields := fieldsToUpdate{
		{spec.HealthzLivenessTimeout, &instance.Status.HealthzLivenessTimeout, "Duration of health check", ""},
		{registry.SecretName, &instance.Status.SecretName, "Name of secret with registry access data", ""},
		storageField,
	}

	updateStatusFields(r.k8s, instance, fields)
	return nil
}

func getStorageField(storage *v1alpha1.Storage, instance *v1alpha1.DockerRegistry) fieldToUpdate {
	storageName := FilesystemStorageName
	if storage != nil {
		if storage.Azure != nil {
			storageName = AzureStorageName
		}
	}
	return fieldToUpdate{storageName, &instance.Status.Storage, "Storage type", ""}
}

func configureControllerConfigurationFlags(s *systemState) {
	s.flagsBuilder.
		WithControllerConfiguration(
			s.instance.Status.HealthzLivenessTimeout,
		)
}
