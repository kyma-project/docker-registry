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
			"dockerRegistry": map[string]interface{}{
				"password": "testPassword",
				"username": "testUsername",
			},
			"registryNodePort": int64(1234),
			"commonLabels": map[string]interface{}{
				"app.kubernetes.io/managed-by": "test",
			},
		}

		flags := NewFlagsBuilder().
			WithNodePort(1234).
			WithRegistryCredentials("testUsername", "testPassword").
			WithRegistryHttpSecret("testHttpSecret").
			WithManagedByLabel("test").
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

func Test_flagsBuilder_withRollme(t *testing.T) {
	t.Run("add rollme flag", func(t *testing.T) {
		flags := flagsBuilder{
			flags: map[string]interface{}{},
		}

		_ = flags.withRollme("reason=test")

		expectedFlags := map[string]interface{}{
			"rollme": "reason=test",
		}
		require.Equal(t, expectedFlags, flags.Build())
	})

	t.Run("add value to existing rollme flag", func(t *testing.T) {
		flags := flagsBuilder{
			flags: map[string]interface{}{
				"rollme": "reason=test",
			},
		}

		_ = flags.withRollme("another-reason=test-2")

		expectedFlags := map[string]interface{}{
			"rollme": "reason=test,another-reason=test-2",
		}
		require.Equal(t, expectedFlags, flags.Build())
	})
}
