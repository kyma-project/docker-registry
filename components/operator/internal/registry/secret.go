package registry

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	InternalAccessSecretName = "dockerregistry-config"
	ExternalAccessSecretName = "dockerregistry-config-external"
	LabelConfigKey           = "dockerregistry.kyma-project.io/config"
	LabelConfigVal           = "credentials"
	DeploymentName           = "dockerregistry"
	HttpEnvKey               = "REGISTRY_HTTP_SECRET"
	ConfigSecretFinalizer    = "dockerregistry.kyma-project.io/finalizer-registry-config"
)

// ReleaseConfigSecretFinalizers removes ConfigSecretFinalizer from every terminating config
// Secret in the given namespace. The Secret controller that normally owns this finalizer is
// deployed into that same namespace, so it can be torn down before the Secrets it guards are
// finalized. Nothing would then be left to release them and the namespace could never finish
// terminating, which also blocks any later reinstallation of the module.
func ReleaseConfigSecretFinalizers(ctx context.Context, c client.Client, namespace string) error {
	var secrets corev1.SecretList
	err := c.List(ctx, &secrets,
		client.InNamespace(namespace),
		client.MatchingLabels{LabelConfigKey: LabelConfigVal},
	)
	if err != nil {
		return err
	}

	for i := range secrets.Items {
		secret := &secrets.Items[i]
		if secret.GetDeletionTimestamp().IsZero() {
			continue
		}
		if !controllerutil.RemoveFinalizer(secret, ConfigSecretFinalizer) {
			continue
		}
		if err := c.Update(ctx, secret); client.IgnoreNotFound(err) != nil {
			return err
		}
	}

	return nil
}

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
