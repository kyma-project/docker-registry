package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnRegistryConfiguration(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateProcessing)
	// setup status.dockerRegistry and set possible warnings
	err := configureRegistry(ctx, r, s)
	if err != nil {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			err,
		)
		return stopWithEventualError(err)
	}

	return nextState(sFnControllerConfiguration)
}

func configureRegistry(ctx context.Context, r *reconciler, s *systemState) error {
	err := setInternalRegistryConfig(ctx, r, s)
	if err != nil {
		return err
	}

	return nil
}

func setInternalRegistryConfig(ctx context.Context, r *reconciler, s *systemState) error {
	existingIntRegSecret, err := registry.GetDockerRegistryInternalRegistrySecret(ctx, r.client, s.instance.Namespace)
	if err != nil {
		return errors.Wrap(err, "while fetching existing internal docker registry secret")
	}
	if existingIntRegSecret != nil {
		r.log.Debugf("reusing existing credentials for internal docker registry to avoiding docker registry  rollout")
		registryHttpSecretEnvValue, getErr := registry.GetRegistryHTTPSecretEnvValue(ctx, r.client, s.instance.Namespace)
		if getErr != nil {
			return errors.Wrap(getErr, "while reading env value registryHttpSecret from internal docker registry deployment")
		}
		s.flagsBuilder.
			WithRegistryCredentials(
				string(existingIntRegSecret.Data["username"]),
				string(existingIntRegSecret.Data["password"]),
			).
			WithRegistryHttpSecret(
				registryHttpSecretEnvValue,
			)
	}

	err = prepareStorage(ctx, r, s) //s.instance.Spec.Storage, s.flagsBuilder)
	if err != nil {
		return errors.Wrap(err, "while preparing storage")
	}

	resolver := registry.NewNodePortResolver(registry.RandomNodePort)
	nodePort, err := resolver.ResolveDockerRegistryNodePortFn(ctx, r.client, s.instance.Namespace)
	if err != nil {
		return errors.Wrap(err, "while resolving registry node port")
	}
	r.log.Debugf("docker registry node port: %d", nodePort)
	s.flagsBuilder.WithNodePort(int64(nodePort))
	return nil
}

func prepareStorage(ctx context.Context, r *reconciler, s *systemState) error { //storage *v1alpha1.Storage, flagsBuilder chart.FlagsBuilder, s *systemState) {
	if s.instance.Spec.Storage != nil {
		if s.instance.Spec.Storage.Azure != nil {
			azureSecret, err := registry.GetSecret(ctx, r.client, s.instance.Spec.Storage.Azure.SecretName, s.instance.Namespace)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("while fetching azure storage secret from %s", s.instance.Namespace))
			}
			storageAzureSecret := &v1alpha1.StorageAzureSecrets{
				AccountName: string(azureSecret.Data["accountName"]),
				AccountKey:  string(azureSecret.Data["accountKey"]),
				Container:   string(azureSecret.Data["container"]),
			}
			s.flagsBuilder.WithAzure(storageAzureSecret)
			return nil
		}
	}
	s.flagsBuilder.WithFilesystem()
	return nil
}
