package utils

import (
	"context"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TestUtils struct {
	Ctx    context.Context
	Logger *zap.SugaredLogger
	Client client.Client

	Namespace                string
	Name                     string
	DockerregistryDeployName string
	RegistryName             string
	CreateSpec               v1alpha1.DockerRegistrySpec
	UpdateSpec               v1alpha1.DockerRegistrySpec

	// StorageSecretName is the Secret with the external storage credentials referenced by CreateSpec.
	StorageSecretName string
	// StorageSecretData holds the credentials the storage Secret is created and rotated with.
	StorageSecretData map[string]string
}
