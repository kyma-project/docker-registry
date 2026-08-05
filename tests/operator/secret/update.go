package secret

import (
	"github.com/kyma-project/docker-registry/tests/operator/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Update(testutils *utils.TestUtils) error {
	var storageSecret corev1.Secret
	objectKey := client.ObjectKey{
		Name:      testutils.StorageSecretName,
		Namespace: testutils.Namespace,
	}

	if err := testutils.Client.Get(testutils.Ctx, objectKey, &storageSecret); err != nil {
		return err
	}

	storageSecret.StringData = testutils.StorageSecretData

	return testutils.Client.Update(testutils.Ctx, &storageSecret)
}
