package controllers

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_mapStorageSecretToDockerRegistryCRs(t *testing.T) {
	dockerRegistry := func(name, namespace string, storage *v1alpha1.Storage) *v1alpha1.DockerRegistry {
		return &v1alpha1.DockerRegistry{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Spec:       v1alpha1.DockerRegistrySpec{Storage: storage},
		}
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "storage-secret", Namespace: "docker-registry"},
	}

	testCases := map[string]struct {
		objects  []client.Object
		expected []ctrl.Request
	}{
		"enqueues the CR referencing the secret": {
			objects: []client.Object{
				dockerRegistry("default", "docker-registry", &v1alpha1.Storage{
					S3: &v1alpha1.StorageS3{SecretName: "storage-secret"},
				}),
			},
			expected: []ctrl.Request{
				{NamespacedName: client.ObjectKey{Namespace: "docker-registry", Name: "default"}},
			},
		},
		"enqueues every backend referencing the secret": {
			objects: []client.Object{
				dockerRegistry("azure", "docker-registry", &v1alpha1.Storage{
					Azure: &v1alpha1.StorageAzure{SecretName: "storage-secret"},
				}),
				dockerRegistry("btp", "docker-registry", &v1alpha1.Storage{
					BTPObjectStore: &v1alpha1.StorageBTPObjectStore{SecretName: "storage-secret"},
				}),
			},
			expected: []ctrl.Request{
				{NamespacedName: client.ObjectKey{Namespace: "docker-registry", Name: "azure"}},
				{NamespacedName: client.ObjectKey{Namespace: "docker-registry", Name: "btp"}},
			},
		},
		"skips a CR referencing another secret": {
			objects: []client.Object{
				dockerRegistry("default", "docker-registry", &v1alpha1.Storage{
					GCS: &v1alpha1.StorageGCS{SecretName: "other-secret"},
				}),
			},
			expected: []ctrl.Request{},
		},
		"skips a CR in another namespace": {
			objects: []client.Object{
				dockerRegistry("default", "other-namespace", &v1alpha1.Storage{
					S3: &v1alpha1.StorageS3{SecretName: "storage-secret"},
				}),
			},
			expected: []ctrl.Request{},
		},
		"skips a CR without external storage": {
			objects: []client.Object{
				dockerRegistry("filesystem", "docker-registry", nil),
				dockerRegistry("pvc", "docker-registry", &v1alpha1.Storage{
					PVC: &v1alpha1.StoragePVC{Name: "pvc"},
				}),
			},
			expected: []ctrl.Request{},
		},
	}

	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			reconciler := &dockerRegistryReconciler{
				client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(testCase.objects...).Build(),
				log:    zap.NewNop().Sugar(),
			}

			requests := reconciler.mapStorageSecretToDockerRegistryCRs(context.Background(), secret)

			require.ElementsMatch(t, testCase.expected, requests)
		})
	}
}
