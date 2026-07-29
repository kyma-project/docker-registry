package controllers

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func Test_storageSecretChanged(t *testing.T) {
	secretWithResourceVersion := func(resourceVersion string) *metav1.PartialObjectMetadata {
		return &metav1.PartialObjectMetadata{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "storage-secret",
				Namespace:       "docker-registry",
				ResourceVersion: resourceVersion,
			},
		}
	}

	t.Run("skips an informer resync", func(t *testing.T) {
		resync := event.TypedUpdateEvent[*metav1.PartialObjectMetadata]{
			ObjectOld: secretWithResourceVersion("42"),
			ObjectNew: secretWithResourceVersion("42"),
		}

		require.False(t, storageSecretChanged.Update(resync))
	})

	t.Run("passes a rotation", func(t *testing.T) {
		rotation := event.TypedUpdateEvent[*metav1.PartialObjectMetadata]{
			ObjectOld: secretWithResourceVersion("42"),
			ObjectNew: secretWithResourceVersion("43"),
		}

		require.True(t, storageSecretChanged.Update(rotation))
	})

	t.Run("passes a secret showing up late", func(t *testing.T) {
		creation := event.TypedCreateEvent[*metav1.PartialObjectMetadata]{
			Object: secretWithResourceVersion("42"),
		}

		require.True(t, storageSecretChanged.Create(creation))
	})
}

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
