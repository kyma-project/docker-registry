package dockerregistry

import (
	"fmt"
	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/tests/operator/dockerregistry/deployment"
	"github.com/kyma-project/docker-registry/tests/operator/utils"
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
