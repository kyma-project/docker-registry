package state

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
	})

	t.Run("setup node port and use existing username and password", func(t *testing.T) {
		registrySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registry.SecretName,
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
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
	})
}
