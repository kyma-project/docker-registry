package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

const (
	AzureStorageName      = "azure"
	GCSStorageName        = "gcs"
	S3StorageName         = "s3"
	FilesystemStorageName = "filesystem"
)

func sFnUpdateFinalStatus(ctx context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	err := updateStatus(ctx, r, s)
	if err != nil {
		return stopWithEventualError(err)
	}

	warning := s.warningBuilder.Build()
	if warning != "" {
		s.setState(v1alpha1.StateWarning)
		s.instance.UpdateConditionTrue(
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstalled,
			warning,
		)
	} else {
		s.setState(v1alpha1.StateReady)
		s.instance.UpdateConditionTrue(
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstalled,
			"DockerRegistry installed",
		)
	}

	return stop()
}

func updateStatus(ctx context.Context, r *reconciler, s *systemState) error {
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
		{registry.InternalAccessSecretName, &s.instance.Status.InternalAccess.SecretName, "Name of secret with registry access data", ""},
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
		{registry.ExternalAccessSecretName, &s.instance.Status.ExternalAccess.SecretName, "Name of secret with registry external access data", ""},
	}, nil
}

func getStorageField(storage *v1alpha1.Storage, instance *v1alpha1.DockerRegistry) fieldToUpdate {
	storageName := FilesystemStorageName
	if storage != nil {
		if storage.Azure != nil {
			storageName = AzureStorageName
		} else if storage.S3 != nil {
			storageName = S3StorageName
		} else if storage.GCS != nil {
			storageName = GCSStorageName
		}

	}
	return fieldToUpdate{storageName, &instance.Status.Storage, "Storage type", ""}
}

type fieldsToUpdate []fieldToUpdate

type fieldToUpdate struct {
	specField    string
	statusField  *string
	fieldName    string
	defaultValue string
}

func updateStatusFields(eventRecorder record.EventRecorder, instance *v1alpha1.DockerRegistry, fields fieldsToUpdate) {
	for _, field := range fields {
		// set default value if spec field is empty
		if field.specField == "" {
			field.specField = field.defaultValue
		}

		if field.specField != *field.statusField {
			oldStatusValue := *field.statusField
			*field.statusField = field.specField
			eventRecorder.Eventf(
				instance,
				"Normal",
				string(v1alpha1.ConditionReasonConfiguration),
				"%s set from '%s' to '%s'",
				field.fieldName,
				oldStatusValue,
				field.specField,
			)
		}
	}
}
