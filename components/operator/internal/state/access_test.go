package state

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"github.com/kyma-project/docker-registry/components/operator/internal/warning"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	istiov1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnAccessConfiguration(t *testing.T) {
	t.Run("setup node port only when registry secret does not exist", func(t *testing.T) {
		s := &systemState{
			instance:         v1alpha1.DockerRegistry{},
			statusSnapshot:   v1alpha1.DockerRegistryStatus{},
			flagsBuilder:     chart.NewFlagsBuilder(),
			nodePortResolver: registry.NewNodePortResolver(registry.RandomNodePort),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}
		expectedFlags := map[string]interface{}{
			"FullnameOverride": "dockerregistry",
			"configData": map[string]interface{}{
				"http": map[string]interface{}{
					"addr": ":5000",
				},
			},
			"registryNodePort": int64(32_137),
			"service": map[string]interface{}{
				"port": int64(5_000),
			},
		}

		next, result, err := sFnAccessConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnStorageConfiguration, next)

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
	})

	t.Run("setup node port and use existing username and password", func(t *testing.T) {
		registrySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registry.InternalAccessSecretName,
				Namespace: "kyma",
				Labels: map[string]string{
					registry.LabelConfigKey: registry.LabelConfigVal,
				},
			},
			Data: map[string][]byte{
				"username": []byte("ala"),
				"password": []byte("makota"),
			},
		}

		registryDeploy := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registry.DeploymentName,
				Namespace: "kyma",
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Env: []corev1.EnvVar{
									{
										Name:  registry.HttpEnvKey,
										Value: "httpEnvKeyVal",
									},
								},
							},
						},
					},
				},
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kyma",
				},
			},
			statusSnapshot:   v1alpha1.DockerRegistryStatus{},
			flagsBuilder:     chart.NewFlagsBuilder(),
			nodePortResolver: registry.NewNodePortResolver(registry.RandomNodePort),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(registrySecret, registryDeploy).Build()},
			log: zap.NewNop().Sugar(),
		}
		expectedFlags := map[string]interface{}{
			"FullnameOverride": "dockerregistry",
			"configData": map[string]interface{}{
				"http": map[string]interface{}{
					"addr": ":5000",
				},
			},
			"registryNodePort": int64(32_137),
			"service": map[string]interface{}{
				"port": int64(5_000),
			},
			"dockerRegistry": map[string]interface{}{
				"username": "ala",
				"password": "makota",
			},
			"registryHTTPSecret": "httpEnvKeyVal",
			"rollme":             "dontrollplease",
		}

		next, result, err := sFnAccessConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnStorageConfiguration, next)

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
	})

	t.Run("setup external access", func(t *testing.T) {
		testScheme := runtime.NewScheme()
		require.NoError(t, istiov1beta1.AddToScheme(testScheme))
		require.NoError(t, clientgoscheme.AddToScheme(testScheme))

		testGateway := &istiov1beta1.Gateway{
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
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-name",
					Namespace: "test-namespace",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					ExternalAccess: &v1alpha1.ExternalAccess{
						Enabled: ptr.To(true),
					},
				},
			},
			statusSnapshot:      v1alpha1.DockerRegistryStatus{},
			flagsBuilder:        chart.NewFlagsBuilder(),
			nodePortResolver:    registry.NewNodePortResolver(registry.RandomNodePort),
			gatewayHostResolver: registry.NewExternalAccessResolver("registry-test-name-test-namespace"),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithScheme(testScheme).WithObjects(testGateway).Build()},
			log: zap.NewNop().Sugar(),
		}
		expectedFlags := map[string]interface{}{
			"FullnameOverride": "dockerregistry",
			"configData": map[string]interface{}{
				"http": map[string]interface{}{
					"addr": ":5000",
				},
			},
			"registryNodePort": int64(32_137),
			"service": map[string]interface{}{
				"port": int64(5_000),
			},
			"virtualService": map[string]interface{}{
				"enabled": true,
				"gateway": "kyma-system/kyma-gateway",
				"host":    "registry-test-name-test-namespace.cluster.local",
			},
		}

		next, result, err := sFnAccessConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnStorageConfiguration, next)

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
	})

	t.Run("external access gateway not found error", func(t *testing.T) {
		testScheme := runtime.NewScheme()
		require.NoError(t, istiov1beta1.AddToScheme(testScheme))
		require.NoError(t, clientgoscheme.AddToScheme(testScheme))

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					ExternalAccess: &v1alpha1.ExternalAccess{
						Enabled: ptr.To(true),
					},
				},
			},
			statusSnapshot:      v1alpha1.DockerRegistryStatus{},
			flagsBuilder:        chart.NewFlagsBuilder(),
			nodePortResolver:    registry.NewNodePortResolver(registry.RandomNodePort),
			gatewayHostResolver: registry.NewExternalAccessResolver(""),
			warningBuilder:      warning.NewBuilder(),
		}

		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithScheme(testScheme).Build()},
			log: zap.NewNop().Sugar(),
		}
		expectedFlags := map[string]interface{}{
			"FullnameOverride": "dockerregistry",
			"configData": map[string]interface{}{
				"http": map[string]interface{}{
					"addr": ":5000",
				},
			},
			"registryNodePort": int64(32_137),
			"service": map[string]interface{}{
				"port": int64(5_000),
			},
		}

		next, result, err := sFnAccessConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnStorageConfiguration, next)

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())

		require.Equal(t, "Warning: .spec.externalAccess.enabled is true but the kyma-gateway Gateway in the kyma-system namespace is not found", s.warningBuilder.Build())
	})
}
