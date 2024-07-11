package kubernetes

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ConfigLabel           = "dockerregistry.kyma-project.io/config"
	CredentialsLabelValue = "credentials"
)

type Config struct {
	BaseNamespace                 string        `envconfig:"default=kyma-system"`
	BaseInternalSecretName        string        `envconfig:"default=dockerregistry-config"`
	BaseExternalSecretName        string        `envconfig:"default=dockerregistry-config-external"`
	ExcludedNamespaces            []string      `envconfig:"default=kyma-system"`
	ConfigMapRequeueDuration      time.Duration `envconfig:"default=1m"`
	SecretRequeueDuration         time.Duration `envconfig:"default=1m"`
	ServiceAccountRequeueDuration time.Duration `envconfig:"default=1m"`
}

func getNamespaces(ctx context.Context, client client.Client, base string, excluded []string) ([]string, error) {
	var namespaces corev1.NamespaceList
	if err := client.List(ctx, &namespaces); err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, namespace := range namespaces.Items {
		if !isExcludedNamespace(namespace.GetName(), base, excluded) && namespace.Status.Phase != corev1.NamespaceTerminating {
			names = append(names, namespace.GetName())
		}
	}

	return names, nil
}

func isExcludedNamespace(name, base string, excluded []string) bool {
	if name == base {
		return true
	}

	for _, namespace := range excluded {
		if name == namespace {
			return true
		}
	}

	return false
}
