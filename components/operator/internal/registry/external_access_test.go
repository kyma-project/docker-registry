package registry

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestExternalAccessResolver_GetExternalAddress(t *testing.T) {
	testScheme := runtime.NewScheme()
	testScheme.AddKnownTypes(v1beta1.SchemeGroupVersion, &v1beta1.Gateway{})

	type fields struct {
		resolvedAddress string
		resolvedError   error
	}
	type args struct {
		c      client.Client
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "return resolved address based on gateway",
			args: args{
				c: fake.NewClientBuilder().WithScheme(testScheme).WithObjects(&v1beta1.Gateway{
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
				}).Build(),
				prefix: "test-prefix",
			},
			want: "test-prefix.cluster.local",
		},
		{
			name: "return previously resolved address",
			fields: fields{
				resolvedAddress: "test-resolved-address",
			},
			want: "test-resolved-address",
		},
		{
			name: "return error when gateway not found",
			args: args{
				c:      fake.NewClientBuilder().WithScheme(testScheme).Build(),
				prefix: "test-prefix",
			},
			wantErr: true,
		},
		{
			name: "return previously resolved error",
			fields: fields{
				resolvedError: errors.New("test-error"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ear := &externalAccessResolver{
				resolvedAddress: tt.fields.resolvedAddress,
				resolvedError:   tt.fields.resolvedError,
			}
			got, err := ear.GetExternalAddress(context.Background(), tt.args.c, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExternalAccessResolver.GetExternalAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExternalAccessResolver.GetExternalAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
