package dockerregistry

import "github.com/kyma-project/docker-registry/tests/operator/utils"

func Delete(utils *utils.TestUtils) error {
	dockerRegistry := fixDockerRegistry(utils)

	return utils.Client.Delete(utils.Ctx, dockerRegistry)
}
