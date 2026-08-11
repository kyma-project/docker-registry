package state

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/flags"
	"github.com/kyma-project/docker-registry/components/operator/internal/warning"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnStorageConfiguration(t *testing.T) {
	t.Run("internal registry using default storage", func(t *testing.T) {
		s := &systemState{
			instance:       v1alpha1.DockerRegistry{},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}
		expectedFlags := map[string]interface{}{
			"configData": map[string]interface{}{
				"storage": map[string]interface{}{
					"filesystem": map[string]interface{}{
						"rootdirectory": "/var/lib/registry",
					},
				},
			},
			"storage": "filesystem",
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnUpdateConfigurationStatus, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		require.EqualValues(t, expectedFlags, flags)
	})

	t.Run("internal registry using azure storage with deleteEnabled", func(t *testing.T) {
		azureSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "azureSecret",
				Namespace: "docker-registry",
			},
			Data: map[string][]byte{
				"accountName": []byte("accountName"),
				"accountKey":  []byte("accountKey"),
				"container":   []byte("container"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "docker-registry",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						DeleteEnabled: true,
						Azure: &v1alpha1.StorageAzure{
							SecretName: "azureSecret",
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(azureSecret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
			"rollme": "configData.storage.delete.enabled=true,secrets.azure=e6c00036b8c46818",
			"configData": map[string]interface{}{
				"storage": map[string]interface{}{
					"delete": map[string]interface{}{
						"enabled": true,
					},
				},
			},
			"storage": "azure",
			"persistence": map[string]interface{}{
				"enabled": false,
			},
			"secrets": map[string]interface{}{
				"azure": map[string]interface{}{
					"accountName": "accountName",
					"accountKey":  "accountKey",
					"container":   "container",
				},
			},
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnUpdateConfigurationStatus, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		require.EqualValues(t, expectedFlags, flags)
	})
	t.Run("internal registry using s3 storage", func(t *testing.T) {
		s3Secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "s3Secret",
				Namespace: "docker-registry",
			},
			Data: map[string][]byte{
				"accessKey": []byte("accessKey"),
				"secretKey": []byte("secretKey"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "docker-registry",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						S3: &v1alpha1.StorageS3{
							Bucket:         "bucket",
							Region:         "region",
							RegionEndpoint: "regionEndpoint",
							Encrypt:        false,
							Secure:         true,
							SecretName:     "s3Secret",
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(s3Secret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
			"rollme": "configData.storage.delete.enabled=false,secrets.s3=911be0030de8f63e",
			"configData": map[string]interface{}{
				"storage": map[string]interface{}{
					"delete": map[string]interface{}{
						"enabled": false,
					},
				},
			},
			"storage": "s3",
			"persistence": map[string]interface{}{
				"enabled": false,
			},
			"s3": map[string]interface{}{
				"bucket":         "bucket",
				"region":         "region",
				"regionEndpoint": "regionEndpoint",
				"encrypt":        false,
				"secure":         true,
			},
			"secrets": map[string]interface{}{
				"s3": map[string]interface{}{
					"accessKey": "accessKey",
					"secretKey": "secretKey",
				},
			},
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnUpdateConfigurationStatus, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		require.EqualValues(t, expectedFlags, flags)
	})
	t.Run("internal registry using gcs storage", func(t *testing.T) {
		gcsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gcsSecret",
				Namespace: "docker-registry",
			},
			Data: map[string][]byte{
				"accountkey": []byte("accountkey"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "docker-registry",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						GCS: &v1alpha1.StorageGCS{
							Bucket:        "gcsBucket",
							SecretName:    "gcsSecret",
							Rootdirectory: "dir",
							Chunksize:     10,
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(gcsSecret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
			"rollme": "configData.storage.delete.enabled=false,secrets.gcs=444580958f2541b1",
			"configData": map[string]interface{}{
				"storage": map[string]interface{}{
					"delete": map[string]interface{}{
						"enabled": false,
					},
				},
			},
			"storage": "gcs",
			"persistence": map[string]interface{}{
				"enabled": false,
			},
			"gcs": map[string]interface{}{
				"bucket":        "gcsBucket",
				"rootdirectory": "dir",
				"chunkSize":     int64(10),
			},
			"secrets": map[string]interface{}{
				"gcs": map[string]interface{}{
					"accountkey": "accountkey",
				},
			},
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnUpdateConfigurationStatus, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		require.EqualValues(t, expectedFlags, flags)
	})

	t.Run("internal registry using btp aws storage", func(t *testing.T) {
		gcsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "btpSecret",
				Namespace: "docker-registry",
			},
			Data: map[string][]byte{
				"host":              []byte("host"),
				"region":            []byte("region"),
				"bucket":            []byte("bucket"),
				"access_key_id":     []byte("accessKey"),
				"secret_access_key": []byte("secretKey"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "docker-registry",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						BTPObjectStore: &v1alpha1.StorageBTPObjectStore{
							SecretName: "btpSecret",
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(gcsSecret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
			"rollme": "configData.storage.delete.enabled=false,secrets.s3=911be0030de8f63e",
			"configData": map[string]interface{}{
				"storage": map[string]interface{}{
					"delete": map[string]interface{}{
						"enabled": false,
					},
				},
			},
			"storage": "s3",
			"persistence": map[string]interface{}{
				"enabled": false,
			},
			"s3": map[string]interface{}{
				"bucket":  "bucket",
				"region":  "region",
				"encrypt": false,
				"secure":  true,
			},
			"secrets": map[string]interface{}{
				"s3": map[string]interface{}{
					"accessKey": "accessKey",
					"secretKey": "secretKey",
				},
			},
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnUpdateConfigurationStatus, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		require.EqualValues(t, expectedFlags, flags)
	})

	t.Run("internal registry using btp azure storage", func(t *testing.T) {
		gcsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "btpSecret",
				Namespace: "docker-registry",
			},
			Data: map[string][]byte{
				"account_name":   []byte("accountName"),
				"sas_token":      []byte("accountKey"),
				"container_name": []byte("container"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "docker-registry",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						BTPObjectStore: &v1alpha1.StorageBTPObjectStore{
							SecretName: "btpSecret",
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(gcsSecret).Build()},
			log: zap.NewNop().Sugar(),
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)

		// Check that warning was recorded
		warnings := s.warningBuilder.Build()
		require.Contains(t, warnings, "Azure storage is not supported for BTPObjectStore")
	})

	t.Run("internal registry using btp gcs storage", func(t *testing.T) {
		gcsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "btpSecret",
				Namespace: "docker-registry",
			},
			Data: map[string][]byte{
				"base64EncodedPrivateKeyData": []byte("YWNjb3VudGtleQ=="),
				"bucket":                      []byte("gcsBucket"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "docker-registry",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						BTPObjectStore: &v1alpha1.StorageBTPObjectStore{
							SecretName: "btpSecret",
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(gcsSecret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
			"rollme": "configData.storage.delete.enabled=false,secrets.gcs=444580958f2541b1",
			"configData": map[string]interface{}{
				"storage": map[string]interface{}{
					"delete": map[string]interface{}{
						"enabled": false,
					},
				},
			},
			"storage": "gcs",
			"persistence": map[string]interface{}{
				"enabled": false,
			},
			"gcs": map[string]interface{}{
				"bucket": "gcsBucket",
			},
			"secrets": map[string]interface{}{
				"gcs": map[string]interface{}{
					"accountkey": "accountkey",
				},
			},
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnUpdateConfigurationStatus, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		require.EqualValues(t, expectedFlags, flags)
	})

	t.Run("internal registry using pvc storage", func(t *testing.T) {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pvc",
				Namespace: "docker-registry",
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "docker-registry",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						PVC: &v1alpha1.StoragePVC{
							Name: "pvc",
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(pvc).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
			"rollme": "configData.storage.delete.enabled=false",
			"configData": map[string]interface{}{
				"storage": map[string]interface{}{
					"delete": map[string]interface{}{
						"enabled": false,
					},
					"filesystem": map[string]interface{}{
						"rootdirectory": "/var/lib/registry",
					},
				},
			},
			"storage": "filesystem",
			"persistence": map[string]interface{}{
				"enabled":       true,
				"existingClaim": "pvc",
			},
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnUpdateConfigurationStatus, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		require.EqualValues(t, expectedFlags, flags)
	})

	t.Run("error when pvc does not exist", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "docker-registry",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						PVC: &v1alpha1.StoragePVC{
							Name: "not-existing-pvc",
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)

		// Check that warning was recorded
		warnings := s.warningBuilder.Build()
		require.Contains(t, warnings, "pvc specified to store images can't be reached because of the error")
	})

	t.Run("internal registry using multiple storages", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "docker-registry",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						BTPObjectStore: &v1alpha1.StorageBTPObjectStore{},
						Azure:          &v1alpha1.StorageAzure{},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)

		// Check that warning was recorded
		warnings := s.warningBuilder.Build()
		require.Contains(t, warnings, "only one storage option can be used")
	})

	t.Run("retry when the storage secret does not exist", func(t *testing.T) {
		testCases := map[string]*v1alpha1.Storage{
			"azure": {Azure: &v1alpha1.StorageAzure{SecretName: "missing-secret"}},
			"s3":    {S3: &v1alpha1.StorageS3{Bucket: "bucket", Region: "region", SecretName: "missing-secret"}},
			"gcs":   {GCS: &v1alpha1.StorageGCS{Bucket: "bucket", SecretName: "missing-secret"}},
			"btp":   {BTPObjectStore: &v1alpha1.StorageBTPObjectStore{SecretName: "missing-secret"}},
		}

		for name, storage := range testCases {
			t.Run(name, func(t *testing.T) {
				s := &systemState{
					instance: v1alpha1.DockerRegistry{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "docker-registry",
						},
						Spec: v1alpha1.DockerRegistrySpec{
							Storage: storage,
						},
					},
					statusSnapshot: v1alpha1.DockerRegistryStatus{},
					flagsBuilder:   flags.NewBuilder(),
					warningBuilder: warning.NewBuilder(),
				}
				r := &reconciler{
					k8s: k8s{client: fake.NewClientBuilder().Build()},
					log: zap.NewNop().Sugar(),
				}

				next, result, err := sFnStorageConfiguration(context.Background(), r, s)
				require.NoError(t, err)
				require.Nil(t, result)
				requireEqualFunc(t, sFnUpdateConfigurationStatus, next)

				require.Contains(t, s.warningBuilder.Build(), `secrets "missing-secret" not found`)
				require.Equal(t, storageRetryInterval, s.retryAfter)
			})
		}
	})

	t.Run("reports Configured false when the s3 storage secret is missing", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "docker-registry",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						S3: &v1alpha1.StorageS3{
							Bucket:     "bucket",
							Region:     "region",
							SecretName: "missing-secret",
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnUpdateConfigurationStatus, next)

		// the reported condition is what the CR ends up showing, so the failure has to survive the
		// state that reports the configuration status
		next, result, err = next(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonConfigurationErr,
			`Warning: failed to set storage configuration: while fetching s3 storage secret from docker-registry: secrets "missing-secret" not found`,
		)
	})

	t.Run("no retry when the storage configuration succeeds", func(t *testing.T) {
		s := &systemState{
			instance:       v1alpha1.DockerRegistry{},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   flags.NewBuilder(),
			warningBuilder: warning.NewBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}

		_, _, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Zero(t, s.retryAfter)
	})
}
