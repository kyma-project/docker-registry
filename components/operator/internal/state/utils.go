package state

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	requeueDuration = time.Second * 3
)

func GetDockerRegistryOrServed(ctx context.Context, req ctrl.Request, c client.Client) (*v1alpha1.DockerRegistry, error) {
	instance := &v1alpha1.DockerRegistry{}
	err := c.Get(ctx, req.NamespacedName, instance)
	if err == nil {
		return instance, nil
	}
	if !k8serrors.IsNotFound(err) {
		return nil, errors.Wrap(err, "while fetching dockerregistry instance")
	}

	instance, err = GetServedDockerRegistry(ctx, c)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching served dockerregistry instance")
	}
	return instance, nil
}

func GetServedDockerRegistry(ctx context.Context, c client.Client) (*v1alpha1.DockerRegistry, error) {
	var dockerRegistryList v1alpha1.DockerRegistryList

	err := c.List(ctx, &dockerRegistryList)

	if err != nil {
		return nil, err
	}

	for _, item := range dockerRegistryList.Items {
		if !item.IsServedEmpty() && item.Status.Served == v1alpha1.ServedTrue {
			return &item, nil
		}
	}

	return nil, nil
}

func updateDockerRegistryWithoutStatus(ctx context.Context, r *reconciler, s *systemState) error {
	return r.client.Update(ctx, &s.instance)
}

func updateDockerRegistryStatus(ctx context.Context, r *reconciler, s *systemState) error {
	if !reflect.DeepEqual(s.instance.Status, s.statusSnapshot) {
		err := r.client.Status().Update(ctx, &s.instance)
		emitEvent(r, s)
		s.saveStatusSnapshot()
		return err
	}
	return nil
}

func resolveRegistryHost(ctx context.Context, r *reconciler, s *systemState) (string, error) {
	hostPrefix := fmt.Sprintf("registry-%s-%s", s.instance.GetName(), s.instance.GetNamespace())

	externalAccess := s.instance.Spec.ExternalAccess
	if externalAccess != nil && externalAccess.HostPrefix != nil {
		hostPrefix = *externalAccess.HostPrefix
	}

	return s.externalAddressResolver.GetExternalAddress(ctx, r.client, hostPrefix)
}
