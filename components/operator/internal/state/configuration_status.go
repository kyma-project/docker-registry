package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AzureStorageName      = "azure"
	S3StorageName         = "s3"
	FilesystemStorageName = "filesystem"
)

func sFnConfigurationStatus(ctx context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	// TODO: I think we should move this to the end of the reconciliation to not update status with new information when
	// (for example) installation can't be fullied because of any error. we should update status only when everything is done
	err := updateConfigurationStatus(ctx, r, s)
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

func updateConfigurationStatus(ctx context.Context, r *reconciler, s *systemState) error {
	spec := s.instance.Spec
	storageField := getStorageField(spec.Storage, &s.instance)
	addresses, err := getInternalAddresses(ctx, r.client, s)
	if err != nil {
		return err
	}

	fields := fieldsToUpdate{
		{registry.SecretName, &s.instance.Status.InternalAccess.SecretName, "Name of secret with registry access data", ""},
		storageField,
	}

	// initialize addresses slice to not work on empty field
	s.instance.Status.InternalAccess.Addresses = make([]string, len(addresses))
	
	addressesToUpdate := []fieldToUpdate{}
	for i := range addresses {
		addressesToUpdate = append(addressesToUpdate,
			fieldToUpdate{addresses[i], &s.instance.Status.InternalAccess.Addresses[i], "Internal address", ""})
	}
	fields = append(fields, addressesToUpdate...)

	updateStatusFields(r.k8s, &s.instance, fields)
	return nil
}

func getStorageField(storage *v1alpha1.Storage, instance *v1alpha1.DockerRegistry) fieldToUpdate {
	storageName := FilesystemStorageName
	if storage != nil {
		if storage.Azure != nil {
			storageName = AzureStorageName
		} else if storage.S3 != nil {
			storageName = S3StorageName
		}

	}
	return fieldToUpdate{storageName, &instance.Status.Storage, "Storage type", ""}
}

func getInternalAddresses(ctx context.Context, c client.Client, s *systemState) ([]string, error) {
	nodeport, err := s.nodePortResolver.GetNodePort(ctx, c, s.instance.GetNamespace())
	if err != nil {
		return nil, err
	}

	return []string{
		fmt.Sprintf("%s.%s.svc.cluster.local:%d", chart.FullnameOverride, s.instance.GetNamespace(), registry.ServicePort),
		fmt.Sprintf("localhost:%d", nodeport),
	}, nil
}
