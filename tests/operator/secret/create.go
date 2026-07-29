package secret

import (
	"github.com/kyma-project/docker-registry/tests/operator/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Create(testutils *utils.TestUtils) error {
	storageSecret := fixStorageSecret(testutils)

	return testutils.Client.Create(testutils.Ctx, storageSecret)
}

func fixStorageSecret(testutils *utils.TestUtils) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      testutils.StorageSecretName,
			Namespace: testutils.Namespace,
		},
		StringData: testutils.StorageSecretData,
	}
}
