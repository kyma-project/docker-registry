package controllers

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_keepIdentityOnly(t *testing.T) {
	t.Run("keeps only what the storage secret watch reads", func(t *testing.T) {
		metadata := &metav1.PartialObjectMetadata{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "storage-secret",
				Namespace:       "docker-registry",
				ResourceVersion: "42",
				Labels:          map[string]string{"app": "storage"},
				Annotations:     map[string]string{"kubectl.kubernetes.io/last-applied-configuration": "{}"},
				ManagedFields:   []metav1.ManagedFieldsEntry{{Manager: "kubectl"}},
				OwnerReferences: []metav1.OwnerReference{{Name: "owner"}},
				Finalizers:      []string{"some-finalizer"},
			},
		}

		result, err := keepIdentityOnly(metadata)

		require.NoError(t, err)
		stripped, ok := result.(*metav1.PartialObjectMetadata)
		require.True(t, ok)
		require.Equal(t, "storage-secret", stripped.Name)
		require.Equal(t, "docker-registry", stripped.Namespace)
		require.Equal(t, "42", stripped.ResourceVersion)
		require.Nil(t, stripped.Labels)
		require.Nil(t, stripped.Annotations)
		require.Nil(t, stripped.ManagedFields)
		require.Nil(t, stripped.OwnerReferences)
		require.Nil(t, stripped.Finalizers)
	})

	t.Run("passes other objects through", func(t *testing.T) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "storage-secret", Namespace: "docker-registry"},
			Data:       map[string][]byte{"accessKey": []byte("key")},
		}

		result, err := keepIdentityOnly(secret)

		require.NoError(t, err)
		require.Equal(t, secret, result)
	})
}
