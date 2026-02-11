package flags

import (
	"testing"

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
			WithLogging("debug", "json", false).
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
			WithLogging("info", "text", true).
			Build()

		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})

	t.Run("skip level and format when empty", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"configData": map[string]interface{}{
				"log": map[string]interface{}{
					"accesslog": map[string]interface{}{
						"disabled": false,
					},
				},
			},
			"rollme": "configData.log.accesslog.disabled=false",
		}

		flags, err := NewBuilder().
			WithLogging("", "", false).
			Build()

		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})
}
