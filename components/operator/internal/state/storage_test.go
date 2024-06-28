package state

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnStorageConfiguration(t *testing.T) {
	t.Run("internal registry using default storage", func(t *testing.T) {
		s := &systemState{
			instance:       v1alpha1.DockerRegistry{},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}
		expectedFlags := map[string]interface{}{
			"configData": map[string]interface{}{
				"storage": map[string]interface{}{
					"filesystem": map[string]interface{}{
						"rootdirectory": "/var/lib/registry",
					},
				},
			},
			"storage": "filesystem",
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			"Configuration ready",
		)
	})

	t.Run("internal registry using azure storage", func(t *testing.T) {
		azureSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "azureSecret",
				Namespace: "kyma-system",
			},
			Data: map[string][]byte{
				"accountName": []byte("accountName"),
				"accountKey":  []byte("accountKey"),
				"container":   []byte("container"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kyma-system",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						Azure: &v1alpha1.StorageAzure{
							SecretName: "azureSecret",
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(azureSecret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
			"storage": "azure",
			"persistence": map[string]interface{}{
				"enabled": false,
			},
			"secrets": map[string]interface{}{
				"azure": map[string]interface{}{
					"accountName": "accountName",
					"accountKey":  "accountKey",
					"container":   "container",
				},
			},
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			"Configuration ready",
		)
	})
	t.Run("internal registry using s3 storage", func(t *testing.T) {
		s3Secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "s3Secret",
				Namespace: "kyma-system",
			},
			Data: map[string][]byte{
				"accessKey": []byte("accessKey"),
				"secretKey": []byte("secretKey"),
			},
		}

		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kyma-system",
				},
				Spec: v1alpha1.DockerRegistrySpec{
					Storage: &v1alpha1.Storage{
						S3: &v1alpha1.StorageS3{
							Bucket:         "bucket",
							Region:         "region",
							RegionEndpoint: "regionEndpoint",
							Encrypt:        false,
							Secure:         true,
							SecretName:     "s3Secret",
						},
					},
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().WithObjects(s3Secret).Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
			"storage": "s3",
			"persistence": map[string]interface{}{
				"enabled": false,
			},
			"s3": map[string]interface{}{
				"bucket":         "bucket",
				"region":         "region",
				"regionEndpoint": "regionEndpoint",
				"encrypt":        false,
				"secure":         true,
			},
			"secrets": map[string]interface{}{
				"s3": map[string]interface{}{
					"accessKey": "accessKey",
					"secretKey": "secretKey",
				},
			},
		}

		next, result, err := sFnStorageConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			"Configuration ready",
		)
	})
}
