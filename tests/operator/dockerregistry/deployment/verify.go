package deployment

import (
	"fmt"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/tests/operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyCtrlMngrEnvs(testutils *utils.TestUtils, dockerRegistry *v1alpha1.DockerRegistry) error {
	var deploy appsv1.Deployment
	objectKey := client.ObjectKey{
		Name:      testutils.CtrlDeployName,
		Namespace: testutils.Namespace,
	}

	err := testutils.Client.Get(testutils.Ctx, objectKey, &deploy)
	if err != nil {
		return err
	}

	return verifyDeployEnvs(&deploy, dockerRegistry)
}

func verifyDeployEnvs(deploy *appsv1.Deployment, dockerRegistry *v1alpha1.DockerRegistry) error {
	expectedEnvs := []corev1.EnvVar{
		{
			Name:  "APP_HEALTHZ_LIVENESS_TIMEOUT",
			Value: dockerRegistry.Status.HealthzLivenessTimeout,
		},
	}
	for _, expectedEnv := range expectedEnvs {
		if !isEnvReflected(expectedEnv, &deploy.Spec.Template.Spec.Containers[0]) {
			return fmt.Errorf("env '%s' with value '%s' not found in deployment", expectedEnv.Name, expectedEnv.Value)
		}
	}

	return nil
}

func isEnvReflected(expected corev1.EnvVar, in *corev1.Container) bool {
	if expected.Value == "" {
		// return true if value is not overrided
		return true
	}

	for _, env := range in.Env {
		if env.Name == expected.Name {
			// return true if value is the same
			return env.Value == expected.Value
		}
	}

	return false
}
