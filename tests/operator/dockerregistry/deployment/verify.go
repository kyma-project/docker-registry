package deployment

import (
	"fmt"
	"github.com/kyma-project/docker-registry/tests/operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyDockerregistryDeployment(testutils *utils.TestUtils) error {
	var deploy appsv1.Deployment
	objectKey := client.ObjectKey{
		Name:      testutils.DockerregistryDeployName,
		Namespace: testutils.Namespace,
	}

	err := testutils.Client.Get(testutils.Ctx, objectKey, &deploy)
	if err != nil {
		return err
	}

	return verifyDeployReadiness(&deploy)
}

func verifyDeployReadiness(deploy *appsv1.Deployment) error {
	if deploy.Status.Replicas != 0 && deploy.Status.Replicas == deploy.Status.ReadyReplicas {
		return nil
	}

	return fmt.Errorf("dockerregistry replicas ready '%d' in total '%d'", deploy.Status.ReadyReplicas, deploy.Status.Replicas)
}
