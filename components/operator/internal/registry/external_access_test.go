package registry

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestExternalAccessResolver_Do(t *testing.T) {
	testScheme := runtime.NewScheme()
	testScheme.AddKnownTypes(v1beta1.SchemeGroupVersion, &v1beta1.Gateway{})

	t.Run("return resolved access based on kyma gateway", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(testScheme).WithObjects(&v1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kyma-gateway",
				Namespace: "kyma-system",
			},
			Spec: networkingv1beta1.Gateway{
				Servers: []*networkingv1beta1.Server{
					{
						Hosts: []string{"*.cluster.local"},
					},
				},
			},
		}).Build()

		ear := &externalAccessResolver{
			defaultKymaGatewayHostPrefix: "test-prefix",
		}

		got, err := ear.Do(context.Background(), client, v1alpha1.ExternalAccess{})
		require.NoError(t, err)
		require.Equal(t, &ResolvedAccess{
			Gateway: "kyma-system/kyma-gateway",
			Host:    "test-prefix.cluster.local",
		}, got)
	})

	t.Run("return resolved access based on kyma gateway and custom host", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(testScheme).WithObjects(&v1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kyma-gateway",
				Namespace: "kyma-system",
			},
			Spec: networkingv1beta1.Gateway{
				Servers: []*networkingv1beta1.Server{
					{
						Hosts: []string{"*.cluster.local"},
					},
				},
			},
		}).Build()

		ear := &externalAccessResolver{
			defaultKymaGatewayHostPrefix: "test-prefix",
		}

		got, err := ear.Do(context.Background(), client, v1alpha1.ExternalAccess{
			Host: ptr.To("registry.cluster.local"),
		})
		require.NoError(t, err)
		require.Equal(t, &ResolvedAccess{
			Gateway: "kyma-system/kyma-gateway",
			Host:    "registry.cluster.local",
		}, got)
	})

	t.Run("return resolved access based on custom gateway and host", func(t *testing.T) {
		ear := &externalAccessResolver{}

		got, err := ear.Do(context.Background(), nil, v1alpha1.ExternalAccess{
			Gateway: ptr.To("kyma-system/custom-gateway"),
			Host:    ptr.To("registry.cluster.custom"),
		})
		require.NoError(t, err)
		require.Equal(t, &ResolvedAccess{
			Gateway: "kyma-system/custom-gateway",
			Host:    "registry.cluster.custom",
		}, got)
	})

	t.Run("return error when host is empty with custom gateway", func(t *testing.T) {
		ear := &externalAccessResolver{}

		got, err := ear.Do(context.Background(), nil, v1alpha1.ExternalAccess{
			Gateway: ptr.To("kyma-system/custom-gateway"),
			Host:    nil,
		})
		require.ErrorContains(t, err, "failed to resolve custom gateway because host is empty")
		require.Nil(t, got)
	})

	t.Run("return previously resolved access", func(t *testing.T) {
		ear := &externalAccessResolver{
			resolvedAccess: &ResolvedAccess{
				Gateway: "default/gateway",
				Host:    "test-resolved-address",
			},
		}

		got, err := ear.Do(context.Background(), nil, v1alpha1.ExternalAccess{})
		require.NoError(t, err)
		require.Equal(t, &ResolvedAccess{
			Gateway: "default/gateway",
			Host:    "test-resolved-address",
		}, got)
	})

	t.Run("return error when gateway not found", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(testScheme).Build()
		ear := &externalAccessResolver{}

		got, err := ear.Do(context.Background(), client, v1alpha1.ExternalAccess{})
		require.ErrorContains(t, err, "while fetching cluster address from Istio Gateway")
		require.Nil(t, got)
	})

	t.Run("return previously resolved error", func(t *testing.T) {
		ear := &externalAccessResolver{
			resolvedError: errors.New("test-error"),
		}

		got, err := ear.Do(context.Background(), nil, v1alpha1.ExternalAccess{})
		require.ErrorContains(t, err, "test-error")
		require.Nil(t, got)
	})
}
