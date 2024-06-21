package registry

import (
	"context"
	"fmt"

	"github.com/kyma-project/docker-registry/components/operator/internal/istio"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExternalAccessResolver interface {
	GetExternalAddress(context.Context, client.Client, string) (string, error)
}

type externalAccessResolver struct {
	resolvedAddress string
}

func NewExternalAccessResolver() ExternalAccessResolver {
	return &externalAccessResolver{}
}

func (ear *externalAccessResolver) GetExternalAddress(ctx context.Context, c client.Client, prefix string) (string, error) {
	if ear.resolvedAddress != "" {
		return ear.resolvedAddress, nil
	}

	clusterAddress, err := istio.GetClusterAddressFromGateway(ctx, c)
	if err != nil {
		return "", errors.Wrap(err, "while fetching cluster address from Istio Gateway")
	}

	ear.resolvedAddress = fmt.Sprintf("%s.%s", prefix, clusterAddress)
	return ear.resolvedAddress, nil
}
