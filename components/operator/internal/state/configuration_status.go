package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	controllerruntime "sigs.k8s.io/controller-runtime"
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

	externalAddressFields, err := getExternalAccessFields(ctx, r, s)
	if err != nil {
		return err
	}

	nodeport, err := s.nodePortResolver.GetNodePort(ctx, r.client, s.instance.GetNamespace())
	if err != nil {
		return err
	}

	pulladdress := fmt.Sprintf("localhost:%d", nodeport)
	pushAddress := fmt.Sprintf("%s.%s.svc.cluster.local:%d", chart.FullnameOverride, s.instance.GetNamespace(), registry.ServicePort)

	fields := append(externalAddressFields, fieldsToUpdate{
		{"True", &s.instance.Status.InternalAccess.Enabled, "Internal access enabled", ""},
		{pulladdress, &s.instance.Status.InternalAccess.PullAddress, "Internal pull address", ""},
		{pushAddress, &s.instance.Status.InternalAccess.PushAddress, "Internal push address", ""},
		{registry.SecretName, &s.instance.Status.InternalAccess.SecretName, "Name of secret with registry access data", ""},
		storageField,
	}...)

	updateStatusFields(r.k8s, &s.instance, fields)
	return nil
}

func getExternalAccessFields(ctx context.Context, r *reconciler, s *systemState) (fieldsToUpdate, error) {
	externalConfigured := s.instance.Spec.ExternalAccess != nil && s.instance.Spec.ExternalAccess.Enabled != nil

	if !externalConfigured || !*s.instance.Spec.ExternalAccess.Enabled {
		// skip if its disabled
		return fieldsToUpdate{
			{"False", &s.instance.Status.ExternalAccess.Enabled, "External access disabled", ""},
			{"", &s.instance.Status.ExternalAccess.PullAddress, "Internal pull address", ""},
			{"", &s.instance.Status.ExternalAccess.PushAddress, "External push address", ""},
			{"", &s.instance.Status.ExternalAccess.SecretName, "Name of secret with registry external access data", ""},
		}, nil
	}

	externalPushAddress, err := resolveRegistryHost(ctx, r, s)
	if err != nil {
		// gateway is not operational but we should continue the reconciliation with old status configuration
		return nil, nil
	}

	return fieldsToUpdate{
		{"True", &s.instance.Status.ExternalAccess.Enabled, "External access enabled", ""},
		{externalPushAddress, &s.instance.Status.ExternalAccess.PullAddress, "External pull address", ""},
		{externalPushAddress, &s.instance.Status.ExternalAccess.PushAddress, "External push address", ""},
		{registry.ExposedSecretName, &s.instance.Status.ExternalAccess.SecretName, "Name of secret with registry external access data", ""},
	}, nil
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
