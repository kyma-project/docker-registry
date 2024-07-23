package state

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnStorageConfiguration(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
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

	return nextState(sFnUpdateConfigurationStatus)
}

func prepareStorage(ctx context.Context, r *reconciler, s *systemState) error {
	if s.instance.Spec.Storage != nil {
		if err := prepareStorageUnique(s); err != nil {
			return err
		}
		s.flagsBuilder.WithPVCDisabled()
		if s.instance.Spec.Storage.Azure != nil {
			return prepareAzureStorage(ctx, r, s)
		} else if s.instance.Spec.Storage.S3 != nil {
			return prepareS3Storage(ctx, r, s)
		} else if s.instance.Spec.Storage.GCS != nil {
			return prepareGCSStorage(ctx, r, s)
		} else if s.instance.Spec.Storage.BTPObjectStore != nil {
			return prepareBTPStorage(ctx, r, s)
		} else if s.instance.Spec.Storage.PVC != nil {
			return preparePVCStorage(s)
		}
	}
	s.flagsBuilder.WithFilesystem()
	return nil
}

func prepareStorageUnique(s *systemState) error {
	// make sure only one of the storage options is used
	storages := 0
	if s.instance.Spec.Storage.Azure != nil {
		storages++
	}
	if s.instance.Spec.Storage.S3 != nil {
		storages++
	}
	if s.instance.Spec.Storage.GCS != nil {
		storages++
	}
	if s.instance.Spec.Storage.BTPObjectStore != nil {
		storages++
	}
	if s.instance.Spec.Storage.PVC != nil {
		storages++
	}
	if storages > 1 {
		return errors.New("only one storage option can be used")
	}
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

func prepareGCSStorage(ctx context.Context, r *reconciler, s *systemState) error {
	gcsSecret, err := registry.GetSecret(ctx, r.client, s.instance.Spec.Storage.GCS.SecretName, s.instance.Namespace)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("while fetching gcs storage secret from %s", s.instance.Namespace))
	}
	storageGCSSecret := &v1alpha1.StorageGCSSecrets{
		AccountKey: string(gcsSecret.Data["accountkey"]),
	}
	s.flagsBuilder.WithGCS(s.instance.Spec.Storage.GCS, storageGCSSecret)
	return nil
}

func prepareBTPStorage(ctx context.Context, r *reconciler, s *systemState) error {
	btpSecret, err := registry.GetSecret(ctx, r.client, s.instance.Spec.Storage.BTPObjectStore.SecretName, s.instance.Namespace)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("while fetching btp storage secret from %s", s.instance.Namespace))
	}
	storageType := getBTPStorageHyperscaler(btpSecret.Data)

	switch storageType {
	case "aws":
		storage := &v1alpha1.StorageS3{
			Bucket: string(btpSecret.Data["bucket"]),
			Region: string(btpSecret.Data["region"]),
			Secure: true,
		}
		storageSecret := &v1alpha1.StorageS3Secrets{
			AccessKey: string(btpSecret.Data["access_key_id"]),
			SecretKey: string(btpSecret.Data["secret_access_key"]),
		}
		s.flagsBuilder.WithS3(storage, storageSecret)
	case "azure":
		// Azure storage uses Azure DNS zone endpoints, which are not supported by distribution
		return errors.New("Azure storage is not supported for BTPObjectStore")
	case "gcp":
		storage := &v1alpha1.StorageGCS{
			Bucket: string(btpSecret.Data["bucket"]),
		}
		// the key is base64-encoded, we're expecting a JSON string
		decodedKey, err := base64.StdEncoding.DecodeString(string(btpSecret.Data["base64EncodedPrivateKeyData"]))
		if err != nil {
			return errors.Wrap(err, "while decoding GCP private key")
		}
		storageSecret := &v1alpha1.StorageGCSSecrets{
			AccountKey: string(decodedKey),
		}
		s.flagsBuilder.WithGCS(storage, storageSecret)
	default:
		return errors.New("unknown storage type")
	}
	return nil
}

func preparePVCStorage(s *systemState) error {
	s.flagsBuilder.WithFilesystem()
	s.flagsBuilder.WithPVC(s.instance.Spec.Storage.PVC)
	return nil
}
