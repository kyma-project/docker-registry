package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testBaseNamespace = "docker-registry"

func TestReleaseConfigSecretFinalizers(t *testing.T) {
	testCases := map[string]struct {
		secret          *corev1.Secret
		terminating     bool
		expectDeleted   bool
		expectFinalizer bool
	}{
		"terminating internal config secret is released": {
			secret:        fixFinalizedSecret(InternalAccessSecretName, testBaseNamespace, true),
			terminating:   true,
			expectDeleted: true,
		},
		"terminating external config secret is released": {
			secret:        fixFinalizedSecret(ExternalAccessSecretName, testBaseNamespace, true),
			terminating:   true,
			expectDeleted: true,
		},
		"config secret that is not terminating keeps its finalizer": {
			secret:          fixFinalizedSecret(InternalAccessSecretName, testBaseNamespace, true),
			terminating:     false,
			expectFinalizer: true,
		},
		"secret without the config label is left alone": {
			secret:          fixFinalizedSecret("unrelated-secret", testBaseNamespace, false),
			terminating:     true,
			expectFinalizer: true,
		},
		"config secret in another namespace is left alone": {
			secret:          fixFinalizedSecret(InternalAccessSecretName, "other-namespace", true),
			terminating:     true,
			expectFinalizer: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			//GIVEN
			c := fake.NewClientBuilder().WithObjects(testCase.secret).Build()
			if testCase.terminating {
				require.NoError(t, c.Delete(context.TODO(), testCase.secret))
			}

			//WHEN
			err := ReleaseConfigSecretFinalizers(context.TODO(), c, testBaseNamespace)

			//THEN
			require.NoError(t, err)

			var actual corev1.Secret
			getErr := c.Get(context.TODO(), client.ObjectKeyFromObject(testCase.secret), &actual)
			if testCase.expectDeleted {
				require.True(t, apierrors.IsNotFound(getErr),
					"secret should be gone once the finalizer is released, got %v", getErr)
				return
			}

			require.NoError(t, getErr)
			if testCase.expectFinalizer {
				require.Contains(t, actual.GetFinalizers(), ConfigSecretFinalizer)
			} else {
				require.NotContains(t, actual.GetFinalizers(), ConfigSecretFinalizer)
			}
		})
	}
}

func TestReleaseConfigSecretFinalizersWithoutSecrets(t *testing.T) {
	c := fake.NewClientBuilder().Build()

	require.NoError(t, ReleaseConfigSecretFinalizers(context.TODO(), c, testBaseNamespace))
}

func fixFinalizedSecret(name, namespace string, configLabel bool) *corev1.Secret {
	labels := map[string]string{}
	if configLabel {
		labels[LabelConfigKey] = LabelConfigVal
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Labels:     labels,
			Finalizers: []string{ConfigSecretFinalizer},
		},
	}
}
