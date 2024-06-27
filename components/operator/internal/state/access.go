package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/kyma-project/docker-registry/components/operator/internal/istio"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnAccessConfiguration(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateProcessing)
	err := setAccessConfig(ctx, r, s)
	if err != nil {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			err,
		)
		return stopWithEventualError(err)
	}

	return nextState(sFnStorageConfiguration)
}

func setAccessConfig(ctx context.Context, r *reconciler, s *systemState) error {
	if err := setInternalAccessConfig(ctx, r, s); err != nil {
		return err
	}

	return setExternalAccessConfig(ctx, r, s)
}

func setInternalAccessConfig(ctx context.Context, r *reconciler, s *systemState) error {
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

	nodePort, err := s.nodePortResolver.GetNodePort(ctx, r.client, s.instance.Namespace)
	if err != nil {
		return errors.Wrap(err, "while resolving registry node port")
	}
	r.log.Debugf("docker registry node port: %d", nodePort)
	s.flagsBuilder.WithNodePort(int64(nodePort)).
		WithServicePort(registry.ServicePort).
		WithFullname(chart.FullnameOverride)
	return nil
}

func setExternalAccessConfig(ctx context.Context, r *reconciler, s *systemState) error {
	if s.instance.Spec.ExternalAccess == nil ||
		s.instance.Spec.ExternalAccess.Enabled == nil ||
		!*s.instance.Spec.ExternalAccess.Enabled {
		// skip if external access is not enabled
		return nil
	}

	gateway := fmt.Sprintf("%s/%s", istio.GatewayNamespace, istio.GatewayName)
	host, err := resolveRegistryHost(ctx, r, s)
	if err != nil {
		return errors.Wrap(err, "while fetching external access host")
	}

	s.flagsBuilder.WithVirtualService(
		host,
		gateway,
	)

	return nil
}

func resolveRegistryHost(ctx context.Context, r *reconciler, s *systemState) (string, error) {
	hostPrefix := fmt.Sprintf("registry-%s-%s", s.instance.GetName(), s.instance.GetNamespace())
	return s.externalAddressResolver.GetExternalAddress(ctx, r.client, hostPrefix)
}
