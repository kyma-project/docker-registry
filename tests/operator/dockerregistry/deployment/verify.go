package deployment

import (
	"fmt"

	"github.com/kyma-project/docker-registry/tests/operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyDockerregistryDeployment(testutils *utils.TestUtils) error {
	deploy, err := getDockerregistryDeployment(testutils)
	if err != nil {
		return err
	}

	return verifyDeployReadiness(deploy)
}

// VerifyRegistryRestarted checks that the registry runs in a Pod started after the given one, so that it
// has read the current content of the storage Secret.
func VerifyRegistryRestarted(testutils *utils.TestUtils, previousPodName string) error {
	podName, err := RegistryPodName(testutils)
	if err != nil {
		return err
	}

	if podName == previousPodName {
		return fmt.Errorf("registry pod '%s' was not restarted", podName)
	}

	return VerifyDockerregistryDeployment(testutils)
}

// RegistryPodName returns the name of the Pod currently running the registry.
func RegistryPodName(testutils *utils.TestUtils) (string, error) {
	deploy, err := getDockerregistryDeployment(testutils)
	if err != nil {
		return "", err
	}

	var pods corev1.PodList
	err = testutils.Client.List(testutils.Ctx, &pods,
		client.InNamespace(testutils.Namespace),
		client.MatchingLabels(deploy.Spec.Selector.MatchLabels))
	if err != nil {
		return "", err
	}

	names := []string{}
	for _, pod := range pods.Items {
		if pod.DeletionTimestamp == nil {
			names = append(names, pod.Name)
		}
	}

	if len(names) != 1 {
		return "", fmt.Errorf("expected one registry pod, found %v", names)
	}

	return names[0], nil
}

func getDockerregistryDeployment(testutils *utils.TestUtils) (*appsv1.Deployment, error) {
	var deploy appsv1.Deployment
	objectKey := client.ObjectKey{
		Name:      testutils.DockerregistryDeployName,
		Namespace: testutils.Namespace,
	}

	err := testutils.Client.Get(testutils.Ctx, objectKey, &deploy)
	if err != nil {
		return nil, err
	}

	return &deploy, nil
}

func verifyDeployReadiness(deploy *appsv1.Deployment) error {
	if deploy.Status.Replicas != 0 && deploy.Status.Replicas == deploy.Status.ReadyReplicas {
		return nil
	}

	return fmt.Errorf("dockerregistry replicas ready '%d' in total '%d'", deploy.Status.ReadyReplicas, deploy.Status.Replicas)
}
