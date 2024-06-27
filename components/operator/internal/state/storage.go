package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnStorageConfiguration(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateProcessing)
	err := prepareStorage(ctx, r, s)
	if err != nil {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			err,
		)
		return stopWithEventualError(err)
	}

	return nextState(sFnConfigurationStatus)
}

func prepareStorage(ctx context.Context, r *reconciler, s *systemState) error {
	if s.instance.Spec.Storage != nil {
		s.flagsBuilder.WithPVCDisabled()
		if s.instance.Spec.Storage.Azure != nil {
			return prepareAzureStorage(ctx, r, s)
		} else if s.instance.Spec.Storage.S3 != nil {
			return prepareS3Storage(ctx, r, s)
		}
	}
	s.flagsBuilder.WithFilesystem()
	return nil
}

func prepareAzureStorage(ctx context.Context, r *reconciler, s *systemState) error {
	azureSecret, err := registry.GetSecret(ctx, r.client, s.instance.Spec.Storage.Azure.SecretName, s.instance.Namespace)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("while fetching azure storage secret from %s", s.instance.Namespace))
	}
	storageAzureSecret := &v1alpha1.StorageAzureSecrets{
		AccountName: string(azureSecret.Data["accountName"]),
		AccountKey:  string(azureSecret.Data["accountKey"]),
		Container:   string(azureSecret.Data["container"]),
	}
	s.flagsBuilder.WithAzure(storageAzureSecret)
	return nil
}

func prepareS3Storage(ctx context.Context, r *reconciler, s *systemState) error {
	s3Secret, err := registry.GetSecret(ctx, r.client, s.instance.Spec.Storage.S3.SecretName, s.instance.Namespace)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("while fetching s3 storage secret from %s", s.instance.Namespace))
	}
	storageS3Secret := &v1alpha1.StorageS3Secrets{
		AccessKey: string(s3Secret.Data["accessKey"]),
		SecretKey: string(s3Secret.Data["secretKey"]),
	}
	s.flagsBuilder.WithS3(s.instance.Spec.Storage.S3, storageS3Secret)
	return nil
}
