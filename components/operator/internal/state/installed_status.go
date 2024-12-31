package state

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	BTPStorageName        = "btp-objectstore"
	AzureStorageName      = "azure"
	GCSStorageName        = "gcs"
	S3StorageName         = "s3"
	FilesystemStorageName = "filesystem"
	PVCStorageName        = "pvc"
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
	storageFields, err := getStorageFields(ctx, spec.Storage, &s.instance, r.client)
	if err != nil {
		return err
	}

	pvcField := getPVCField(spec.Storage, &s.instance)

	externalAddressFields := getExternalAccessFields(ctx, r, s)

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
		pvcField,
	}...)
	fields = append(fields, storageFields...)

	updateStatusFields(r.k8s, &s.instance, fields)
	return nil
}

func getExternalAccessFields(ctx context.Context, r *reconciler, s *systemState) fieldsToUpdate {
	externalConfigured := s.instance.Spec.ExternalAccess != nil && s.instance.Spec.ExternalAccess.Enabled != nil

	if !externalConfigured || !*s.instance.Spec.ExternalAccess.Enabled {
		// skip if its disabled
		return fieldsToUpdate{
			{"False", &s.instance.Status.ExternalAccess.Enabled, "External access disabled", ""},
			{"", &s.instance.Status.ExternalAccess.PullAddress, "Internal pull address", ""},
			{"", &s.instance.Status.ExternalAccess.PushAddress, "External push address", ""},
			{"", &s.instance.Status.ExternalAccess.Gateway, "External gateway namespaced name", ""},
			{"", &s.instance.Status.ExternalAccess.SecretName, "Name of secret with registry external access data", ""},
		}
	}

	resolvedAccess, err := s.gatewayHostResolver.Do(ctx, r.client, *s.instance.Spec.ExternalAccess)
	if err != nil {
		// gateway is not operational but we should continue the reconciliation with old status configuration
		return nil
	}

	return fieldsToUpdate{
		{"True", &s.instance.Status.ExternalAccess.Enabled, "External access enabled", ""},
		{resolvedAccess.Host, &s.instance.Status.ExternalAccess.PullAddress, "External pull address", ""},
		{resolvedAccess.Host, &s.instance.Status.ExternalAccess.PushAddress, "External push address", ""},
		{resolvedAccess.Gateway, &s.instance.Status.ExternalAccess.Gateway, "External gateway namespaced name", ""},
		{registry.ExternalAccessSecretName, &s.instance.Status.ExternalAccess.SecretName, "Name of secret with registry external access data", ""},
	}
}

func getStorageFields(ctx context.Context, storage *v1alpha1.Storage, instance *v1alpha1.DockerRegistry, client client.Client) (fieldsToUpdate, error) {
	storageName := FilesystemStorageName
	deleteEnabled := "False"
	if storage != nil {
		deleteEnabled = cases.Title(language.Und).String(strconv.FormatBool(instance.Spec.Storage.DeleteEnabled))

		if storage.Azure != nil {
			storageName = AzureStorageName
		} else if storage.S3 != nil {
			storageName = S3StorageName
		} else if storage.GCS != nil {
			storageName = GCSStorageName
		} else if storage.BTPObjectStore != nil {
			btpSecret, err := registry.GetSecret(ctx, client, instance.Spec.Storage.BTPObjectStore.SecretName, instance.Namespace)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("while fetching btp storage secret from %s", instance.Namespace))
			}
			storageType := getBTPStorageHyperscaler(btpSecret.Data)
			storageName = fmt.Sprintf("%s-%s", BTPStorageName, storageType)
		} else if storage.PVC != nil {
			storageName = PVCStorageName
		}
	}
	return fieldsToUpdate{
		{storageName, &instance.Status.Storage, "Storage type", ""},
		{deleteEnabled, &instance.Status.DeleteEnabled, "Enable image blobs and manifests by digest", ""},
	}, nil
}

func getPVCField(storage *v1alpha1.Storage, instance *v1alpha1.DockerRegistry) fieldToUpdate {
	if storage != nil && storage.PVC != nil {
		return fieldToUpdate{storage.PVC.Name, &instance.Status.PVC, "PVC name", ""}
	}
	return fieldToUpdate{"", &instance.Status.PVC, "PVC name", ""}
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
