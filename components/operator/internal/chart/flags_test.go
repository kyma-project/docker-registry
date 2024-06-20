package chart

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_flagsBuilder_Build(t *testing.T) {
	t.Run("build empty flags", func(t *testing.T) {
		flags := NewFlagsBuilder().Build()
		require.Equal(t, map[string]interface{}{}, flags)
	})

	t.Run("build flags", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"registryHTTPSecret": "testHttpSecret",
			"rollme":             "dontrollplease",
			"dockerRegistry": map[string]interface{}{
				"password": "testPassword",
				"username": "testUsername",
			},
			"registryNodePort": int64(1234),
		}

		flags := NewFlagsBuilder().
			WithNodePort(1234).
			WithRegistryCredentials("testUsername", "testPassword").
			WithRegistryHttpSecret("testHttpSecret").
			Build()

		require.Equal(t, expectedFlags, flags)
	})

	t.Run("build registry flags only", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"password": "testPassword",
				"username": "testUsername",
			},
		}

		flags := NewFlagsBuilder().
			WithRegistryCredentials("testUsername", "testPassword").
			Build()

		require.Equal(t, expectedFlags, flags)
	})
}
