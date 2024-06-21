package istio

import (
	"context"
	"fmt"

	"istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	GatewayName      = "kyma-gateway"
	GatewayNamespace = "kyma-system"
)

func GetClusterAddressFromGateway(ctx context.Context, c client.Client) (string, error) {
	gateway := &v1beta1.Gateway{}

	err := c.Get(ctx, client.ObjectKey{
		Name:      GatewayName,
		Namespace: GatewayNamespace,
	}, gateway)
	if err != nil {
		return "", fmt.Errorf("while getting Gateway %s in namespace %s: %w", GatewayName, GatewayNamespace, err)
	}

	// kyma gateway can't be modified by the user so we can assume that it has at least one server and host
	// this `if` should protect us when this situation changes
	if len(gateway.Spec.Servers) < 1 || len(gateway.Spec.Servers[0].Hosts) < 1 {
		return "", fmt.Errorf("the Gateway %s in namespace %s does not have any hosts defined", GatewayName, GatewayNamespace)
	}

	host := gateway.Spec.Servers[0].Hosts[0]

	// host is always in format '*.<address>' so we need to remove the first two characters
	return host[2:], nil
}
