package dockerregistry

import (
	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/tests/operator/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Update(testutils *utils.TestUtils) error {
	var dockerregistry v1alpha1.DockerRegistry
	objectKey := client.ObjectKey{
		Name:      testutils.Name,
		Namespace: testutils.Namespace,
	}

	if err := testutils.Client.Get(testutils.Ctx, objectKey, &dockerregistry); err != nil {
		return err
	}

	dockerregistry.Spec = testutils.UpdateSpec

	return testutils.Client.Update(testutils.Ctx, &dockerregistry)
}
