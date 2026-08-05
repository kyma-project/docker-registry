package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDockerRegistry_StorageSecretName(t *testing.T) {
	testCases := map[string]struct {
		storage  *Storage
		expected string
	}{
		"no storage": {
			storage:  nil,
			expected: "",
		},
		"azure": {
			storage:  &Storage{Azure: &StorageAzure{SecretName: "azure-secret"}},
			expected: "azure-secret",
		},
		"s3": {
			storage:  &Storage{S3: &StorageS3{SecretName: "s3-secret"}},
			expected: "s3-secret",
		},
		"gcs": {
			storage:  &Storage{GCS: &StorageGCS{SecretName: "gcs-secret"}},
			expected: "gcs-secret",
		},
		"btp object store": {
			storage:  &Storage{BTPObjectStore: &StorageBTPObjectStore{SecretName: "btp-secret"}},
			expected: "btp-secret",
		},
		"pvc has no credentials secret": {
			storage:  &Storage{PVC: &StoragePVC{Name: "pvc"}},
			expected: "",
		},
		"s3 without a secret": {
			storage:  &Storage{S3: &StorageS3{Bucket: "bucket"}},
			expected: "",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			instance := &DockerRegistry{Spec: DockerRegistrySpec{Storage: testCase.storage}}

			require.Equal(t, testCase.expected, instance.StorageSecretName())
		})
	}
}
