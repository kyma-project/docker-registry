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

func sFnConfigurationStatus(_ context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	// TODO: I think we should move this to the end of the reconciliation to not update status with new information when
	// (for example) installation can't be fullied because of any error. we should update status only when everything is done
	err := updateConfigurationStatus(r, &s.instance)
	if err != nil {
		return stopWithEventualError(err)
	}

	s.setState(v1alpha1.StateProcessing)
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeConfigured,
		v1alpha1.ConditionReasonConfigured,
		"Configuration ready",
	)

	return nextState(sFnApplyResources)
}

func updateConfigurationStatus(r *reconciler, instance *v1alpha1.DockerRegistry) error {
	spec := instance.Spec
	storageField := getStorageField(spec.Storage, instance)
	fields := fieldsToUpdate{
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
