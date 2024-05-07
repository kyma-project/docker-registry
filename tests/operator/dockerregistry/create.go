package dockerregistry

import (
	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/tests/operator/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Create(utils *utils.TestUtils) error {
	dockerRegistryObj := fixDockerRegistry(utils)

	return utils.Client.Create(utils.Ctx, dockerRegistryObj)
}

func fixDockerRegistry(testUtils *utils.TestUtils) *v1alpha1.DockerRegistry {
	return &v1alpha1.DockerRegistry{
		ObjectMeta: v1.ObjectMeta{
			Name:      testUtils.Name,
			Namespace: testUtils.Namespace,
		},
		Spec: v1alpha1.DockerRegistrySpec{},
	}
}
