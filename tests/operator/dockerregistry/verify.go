package dockerregistry

import (
	"fmt"
	"strings"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/tests/operator/dockerregistry/deployment"
	"github.com/kyma-project/docker-registry/tests/operator/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyDeletion(utils *utils.TestUtils) error {
	err := Verify(utils)
	if !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func Verify(utils *utils.TestUtils) error {
	var dockerRegistry v1alpha1.DockerRegistry
	objectKey := client.ObjectKey{
		Name:      utils.Name,
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &dockerRegistry); err != nil {
		return err
	}

	if err := verifyState(utils, &dockerRegistry); err != nil {
		return err
	}

	if err := deployment.VerifyDockerregistryDeployment(utils); err != nil {
		return err
	}

	return nil
}

func verifyState(utils *utils.TestUtils, dockerRegistry *v1alpha1.DockerRegistry) error {
	if dockerRegistry.Status.State != v1alpha1.StateReady {
		return fmt.Errorf("dockerregistry '%s' in '%s' state", utils.Name, dockerRegistry.Status.State)
	}

	return nil
}

// VerifyMissingStorageSecretWarning checks that the CR complains about the storage Secret it references
// instead of reporting any other state.
func VerifyMissingStorageSecretWarning(testutils *utils.TestUtils) error {
	var dockerRegistry v1alpha1.DockerRegistry
	objectKey := client.ObjectKey{
		Name:      testutils.Name,
		Namespace: testutils.Namespace,
	}

	if err := testutils.Client.Get(testutils.Ctx, objectKey, &dockerRegistry); err != nil {
		return err
	}

	if dockerRegistry.Status.State != v1alpha1.StateWarning {
		return fmt.Errorf("dockerregistry '%s' in '%s' state, expected '%s'",
			testutils.Name, dockerRegistry.Status.State, v1alpha1.StateWarning)
	}

	for _, condition := range dockerRegistry.Status.Conditions {
		if strings.Contains(condition.Message, testutils.StorageSecretName) {
			return nil
		}
	}

	return fmt.Errorf("dockerregistry '%s' does not report the missing storage secret '%s'",
		testutils.Name, testutils.StorageSecretName)
}

// VerifyStorageCredentials checks that the credentials the registry Pods read come from the current
// content of the storage Secret.
func VerifyStorageCredentials(testutils *utils.TestUtils) error {
	var registrySecret corev1.Secret
	objectKey := client.ObjectKey{
		// the chart names the generated Secret after the registry Deployment
		Name:      fmt.Sprintf("%s-secret", testutils.DockerregistryDeployName),
		Namespace: testutils.Namespace,
	}

	if err := testutils.Client.Get(testutils.Ctx, objectKey, &registrySecret); err != nil {
		return err
	}

	for storageKey, registryKey := range map[string]string{"accessKey": "s3AccessKey", "secretKey": "s3SecretKey"} {
		if string(registrySecret.Data[registryKey]) != testutils.StorageSecretData[storageKey] {
			return fmt.Errorf("secret '%s' does not contain the current '%s'", objectKey.Name, storageKey)
		}
	}

	return nil
}
