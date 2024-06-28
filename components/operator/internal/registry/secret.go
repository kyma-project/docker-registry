package registry

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	InternalAccessSecretName = "dockerregistry-config"
	ExternalAccessSecretName = "dockerregistry-config-external"
	LabelConfigKey           = "dockerregistry.kyma-project.io/config"
	LabelConfigVal           = "credentials"
	DeploymentName           = "dockerregistry"
	HttpEnvKey               = "REGISTRY_HTTP_SECRET"
)

func GetDockerRegistryInternalRegistrySecret(ctx context.Context, c client.Client, namespace string) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      InternalAccessSecretName,
	}
	err := c.Get(ctx, key, &secret)
	if err != nil {
		return nil, client.IgnoreNotFound(err)
	}

	if val, ok := secret.GetLabels()[LabelConfigKey]; !ok || val != LabelConfigVal {
		return nil, nil
	}

	return &secret, nil
}

func GetRegistryHTTPSecretEnvValue(ctx context.Context, c client.Client, namespace string) (string, error) {
	deployment := appsv1.Deployment{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      DeploymentName,
	}
	err := c.Get(ctx, key, &deployment)
	if err != nil {
		return "", client.IgnoreNotFound(err)
	}

	envs := deployment.Spec.Template.Spec.Containers[0].Env
	for _, v := range envs {
		if v.Name == HttpEnvKey && v.Value != "" {
			return v.Value, nil
		}
	}

	return "", nil
}

func GetSecret(ctx context.Context, c client.Client, name, namespace string) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}
	err := c.Get(ctx, key, &secret)
	if err != nil {
		return nil, err
	}

	return &secret, nil
}
