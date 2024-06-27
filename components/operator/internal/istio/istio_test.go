package istio

import (
	"context"
	"testing"

	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetClusterAddressFromGateway22(t *testing.T) {
	testScheme := runtime.NewScheme()
	testScheme.AddKnownTypes(v1beta1.SchemeGroupVersion, &v1beta1.Gateway{})

	tests := []struct {
		name    string
		client  client.Client
		want    string
		wantErr bool
	}{
		{
			name: "Should return cluster address",
			client: fake.NewClientBuilder().WithScheme(testScheme).WithObjects(fixTestGateway([]*networkingv1beta1.Server{
				{
					Hosts: []string{"*.cluster.local"},
				},
			})).Build(),
			want: "cluster.local",
		},
		{
			name:    "Should return err when gateway not found",
			client:  fake.NewClientBuilder().WithScheme(testScheme).Build(),
			wantErr: true,
		},
		{
			name:    "Should return err when gateway has no servers",
			client:  fake.NewClientBuilder().WithScheme(testScheme).WithObjects(fixTestGateway([]*networkingv1beta1.Server{})).Build(),
			wantErr: true,
		},
		{
			name: "Should return err when gateway has no hosts",
			client: fake.NewClientBuilder().WithScheme(testScheme).WithObjects(fixTestGateway([]*networkingv1beta1.Server{
				{},
			})).Build(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetClusterAddressFromGateway(context.Background(), tt.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClusterAddressFromGateway() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetClusterAddressFromGateway() = %v, want %v", got, tt.want)
			}
		})
	}
}

func fixTestGateway(servers []*networkingv1beta1.Server) *v1beta1.Gateway {
	return &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kyma-gateway",
			Namespace: "kyma-system",
		},
		Spec: networkingv1beta1.Gateway{
			Servers: servers,
		},
	}
}
