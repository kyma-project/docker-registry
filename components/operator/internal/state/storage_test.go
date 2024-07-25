package state

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
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
			flagsBuilder:   chart.NewFlagsBuilder(),
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

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
	})

	t.Run("internal registry using azure storage", func(t *testing.T) {
		azureSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "azureSecret",
				Namespace: "kyma-system",
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
					Namespace: "kyma-system",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						Azure: &v1alpha1.StorageAzure{
							SecretName: "azureSecret",
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(azureSecret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
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

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
	})
	t.Run("internal registry using s3 storage", func(t *testing.T) {
		s3Secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "s3Secret",
				Namespace: "kyma-system",
			},
			Data: map[string][]byte{
				"accessKey": []byte("accessKey"),
				"secretKey": []byte("secretKey"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kyma-system",
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
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(s3Secret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
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

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
	})
	t.Run("internal registry using gcs storage", func(t *testing.T) {
		gcsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gcsSecret",
				Namespace: "kyma-system",
			},
			Data: map[string][]byte{
				"accountkey": []byte("accountkey"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kyma-system",
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
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(gcsSecret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
			"storage": "gcs",
			"persistence": map[string]interface{}{
				"enabled": false,
			},
			"gcs": map[string]interface{}{
				"bucket":        "gcsBucket",
				"rootdirectory": "dir",
				"chunkSize":     10,
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

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
	})

	t.Run("internal registry using btp aws storage", func(t *testing.T) {
		gcsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "btpSecret",
				Namespace: "kyma-system",
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
					Namespace: "kyma-system",
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
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(gcsSecret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
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

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
	})

	t.Run("internal registry using btp azure storage", func(t *testing.T) {
		gcsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "btpSecret",
				Namespace: "kyma-system",
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
					Namespace: "kyma-system",
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
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(gcsSecret).Build()},
			log: zap.NewNop().Sugar(),
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.Error(t, err)
		require.Nil(t, result)
		require.Nil(t, next)
	})

	t.Run("internal registry using btp gcs storage", func(t *testing.T) {
		gcsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "btpSecret",
				Namespace: "kyma-system",
			},
			Data: map[string][]byte{
				"base64EncodedPrivateKeyData": []byte("YWNjb3VudGtleQ=="),
				"bucket":                      []byte("gcsBucket"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kyma-system",
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
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(gcsSecret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
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

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
	})

	t.Run("internal registry using btp pvc storage", func(t *testing.T) {
		gcsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "btpSecret",
				Namespace: "kyma-system",
			},
			Data: map[string][]byte{
				"base64EncodedPrivateKeyData": []byte("YWNjb3VudGtleQ=="),
				"bucket":                      []byte("gcsBucket"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kyma-system",
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
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(gcsSecret).Build()},
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
			"persistence": map[string]interface{}{
				"enabled":       true,
				"existingClaim": "pvc",
			},
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnUpdateConfigurationStatus, next)

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
	})

	t.Run("internal registry using multiple storages", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kyma-system",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						BTPObjectStore: &v1alpha1.StorageBTPObjectStore{},
						Azure:          &v1alpha1.StorageAzure{},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.Error(t, err)
		require.Nil(t, result)
		require.Nil(t, next)
	})

}
