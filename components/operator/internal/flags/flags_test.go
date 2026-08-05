package flags

import (
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/stretchr/testify/require"
)

func Test_flagsBuilder_Build(t *testing.T) {
	t.Run("build empty flags", func(t *testing.T) {
		flags, err := NewBuilder().Build()
		require.NoError(t, err)
		require.Equal(t, map[string]interface{}{}, flags)
	})

	t.Run("build flags", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"registryHTTPSecret": "testHttpSecret",
			"dockerRegistry": map[string]interface{}{
				"password": "testPassword",
				"username": "testUsername",
			},
			"registryNodePort": int64(1234),
			"commonLabels": map[string]interface{}{
				"app.kubernetes.io/managed-by": "test",
			},
		}

		flags, err := NewBuilder().
			WithNodePort(1234).
			WithRegistryCredentials("testUsername", "testPassword").
			WithRegistryHttpSecret("testHttpSecret").
			WithManagedByLabel("test").
			Build()

		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})

	t.Run("build registry flags only", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"password": "testPassword",
				"username": "testUsername",
			},
		}

		flags, err := NewBuilder().
			WithRegistryCredentials("testUsername", "testPassword").
			Build()

		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})
}

func Test_flagsBuilder_withRollme(t *testing.T) {
	t.Run("add rollme flag", func(t *testing.T) {
		builder := Builder{
			FlagsBuilder: chart.NewFlagsBuilder(),
		}

		_ = builder.withRollme("reason=test")

		expectedFlags := map[string]interface{}{
			"rollme": "reason=test",
		}

		flags, err := builder.Build()
		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})

	t.Run("add value to existing rollme flag", func(t *testing.T) {
		builder := Builder{
			FlagsBuilder: chart.NewFlagsBuilder(),
		}

		_ = builder.
			withRollme("reason=test").
			withRollme("another-reason=test-2").
			withRollme("test=test2")

		expectedFlags := map[string]interface{}{
			"rollme": "reason=test,another-reason=test-2,test=test2",
		}

		flags, err := builder.Build()
		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})
}

func Test_flagsBuilder_WithLogging(t *testing.T) {
	t.Run("set log level and format with access logs enabled", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"configData": map[string]interface{}{
				"log": map[string]interface{}{
					"level":     "debug",
					"formatter": "json",
					"accesslog": map[string]interface{}{
						"disabled": false,
					},
				},
			},
			"rollme": "configData.log.level=debug,configData.log.formatter=json,configData.log.accesslog.disabled=false",
		}

		flags, err := NewBuilder().
			WithLogging("debug", "json", true).
			Build()

		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})

	t.Run("set log level and format with access logs disabled", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"configData": map[string]interface{}{
				"log": map[string]interface{}{
					"level":     "info",
					"formatter": "text",
					"accesslog": map[string]interface{}{
						"disabled": true,
					},
				},
			},
			"rollme": "configData.log.level=info,configData.log.formatter=text,configData.log.accesslog.disabled=true",
		}

		flags, err := NewBuilder().
			WithLogging("info", "text", false).
			Build()

		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})

	t.Run("skip level and format when empty", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"configData": map[string]interface{}{
				"log": map[string]interface{}{
					"accesslog": map[string]interface{}{
						"disabled": true,
					},
				},
			},
			"rollme": "configData.log.accesslog.disabled=true",
		}

		flags, err := NewBuilder().
			WithLogging("", "", false).
			Build()

		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})
}

func Test_flagsBuilder_credentialsRollme(t *testing.T) {
	rollmeOf := func(t *testing.T, build func(*Builder) *Builder) string {
		t.Helper()
		flags, err := build(NewBuilder()).Build()
		require.NoError(t, err)
		rollme, found := flags["rollme"]
		require.True(t, found, "expected storage credentials to produce a rollme value")
		return rollme.(string)
	}

	t.Run("rotated credentials change the rollme value", func(t *testing.T) {
		testCases := map[string]struct {
			before func(*Builder) *Builder
			after  func(*Builder) *Builder
		}{
			"azure": {
				before: func(b *Builder) *Builder {
					return b.WithAzure(&v1alpha1.StorageAzureSecrets{AccountName: "name", AccountKey: "old-key", Container: "container"})
				},
				after: func(b *Builder) *Builder {
					return b.WithAzure(&v1alpha1.StorageAzureSecrets{AccountName: "name", AccountKey: "new-key", Container: "container"})
				},
			},
			"s3": {
				before: func(b *Builder) *Builder {
					return b.WithS3(&v1alpha1.StorageS3{Bucket: "bucket", Region: "region"}, &v1alpha1.StorageS3Secrets{AccessKey: "access", SecretKey: "old-key"})
				},
				after: func(b *Builder) *Builder {
					return b.WithS3(&v1alpha1.StorageS3{Bucket: "bucket", Region: "region"}, &v1alpha1.StorageS3Secrets{AccessKey: "access", SecretKey: "new-key"})
				},
			},
			"gcs": {
				before: func(b *Builder) *Builder {
					return b.WithGCS(&v1alpha1.StorageGCS{Bucket: "bucket"}, &v1alpha1.StorageGCSSecrets{AccountKey: "old-key"})
				},
				after: func(b *Builder) *Builder {
					return b.WithGCS(&v1alpha1.StorageGCS{Bucket: "bucket"}, &v1alpha1.StorageGCSSecrets{AccountKey: "new-key"})
				},
			},
		}

		for name, testCase := range testCases {
			t.Run(name, func(t *testing.T) {
				before := rollmeOf(t, testCase.before)
				after := rollmeOf(t, testCase.after)

				require.NotEqual(t, before, after)
				// the same credentials have to produce the same value, otherwise the registry
				// would be restarted on every reconciliation
				require.Equal(t, before, rollmeOf(t, testCase.before))
			})
		}
	})

	t.Run("rollme does not leak credentials", func(t *testing.T) {
		secret := "super-secret-value"

		rollmes := []string{
			rollmeOf(t, func(b *Builder) *Builder {
				return b.WithAzure(&v1alpha1.StorageAzureSecrets{AccountName: "name", AccountKey: secret, Container: "container"})
			}),
			rollmeOf(t, func(b *Builder) *Builder {
				return b.WithS3(&v1alpha1.StorageS3{Bucket: "bucket", Region: "region"}, &v1alpha1.StorageS3Secrets{AccessKey: "access", SecretKey: secret})
			}),
			rollmeOf(t, func(b *Builder) *Builder {
				return b.WithGCS(&v1alpha1.StorageGCS{Bucket: "bucket"}, &v1alpha1.StorageGCSSecrets{AccountKey: secret})
			}),
		}

		for _, rollme := range rollmes {
			require.NotContains(t, rollme, secret)
		}
	})

	t.Run("storage without credentials does not set rollme", func(t *testing.T) {
		flags, err := NewBuilder().
			WithS3(&v1alpha1.StorageS3{Bucket: "bucket", Region: "region"}, nil).
			Build()

		require.NoError(t, err)
		require.NotContains(t, flags, "rollme")
	})
}
