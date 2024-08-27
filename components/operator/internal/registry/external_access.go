package registry

import (
	"context"
	"fmt"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/istio"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResolvedAccess struct {
	Gateway string
	Host    string
}

type ExternalAccessResolver interface {
	Do(context.Context, client.Client, v1alpha1.ExternalAccess) (*ResolvedAccess, error)
}

type externalAccessResolver struct {
	defaultKymaGatewayHostPrefix string
	resolvedAccess               *ResolvedAccess
	resolvedError                error
}

func NewExternalAccessResolver(defaultKymaGatewayHostPrefix string) ExternalAccessResolver {
	return &externalAccessResolver{
		defaultKymaGatewayHostPrefix: defaultKymaGatewayHostPrefix,
	}
}

// Do returns host that can be used to access registry from outside of the cluster of error if host is not operational
func (ear *externalAccessResolver) Do(ctx context.Context, client client.Client, externalAccess v1alpha1.ExternalAccess) (*ResolvedAccess, error) {
	if ear.resolvedAccess != nil || ear.resolvedError != nil {
		return ear.resolvedAccess, ear.resolvedError
	}

	ear.resolvedAccess, ear.resolvedError = ear.resolveAccess(
		ctx,
		client,
		externalAccess.Gateway,
		externalAccess.Host,
	)

	return ear.resolvedAccess, ear.resolvedError
}

func (ear *externalAccessResolver) resolveAccess(ctx context.Context, client client.Client, gateway, customHost *string) (*ResolvedAccess, error) {
	if gateway != nil {
		// resolve host for custom gateway - not kyma gateway
		return resolveAccessWithCustomGateway(gateway, customHost)
	}

	clusterAddress, err := istio.GetClusterAddressFromGateway(ctx, client)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching cluster address from Istio Gateway")
	}

	registryHost := fmt.Sprintf("%s.%s", ear.defaultKymaGatewayHostPrefix, clusterAddress)
	if customHost != nil {
		registryHost = *customHost
	}

	return &ResolvedAccess{
		Host:    registryHost,
		Gateway: fmt.Sprintf("%s/%s", istio.GatewayNamespace, istio.GatewayName),
	}, nil
}

func resolveAccessWithCustomGateway(gateway, host *string) (*ResolvedAccess, error) {
	if host == nil {
		return nil, errors.New("failed to resolve custom gateway because host is empty")
	}

	return &ResolvedAccess{
		Host:    *host,
		Gateway: *gateway,
	}, nil
}
