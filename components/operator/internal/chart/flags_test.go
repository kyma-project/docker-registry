package chart

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_flagsBuilder_Build(t *testing.T) {
	t.Run("build empty flags", func(t *testing.T) {
		flags, err := NewFlagsBuilder().Build()
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

		flags, err := NewFlagsBuilder().
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

		flags, err := NewFlagsBuilder().
			WithRegistryCredentials("testUsername", "testPassword").
			Build()

		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})
}

func Test_flagsBuilder_withRollme(t *testing.T) {
	t.Run("add rollme flag", func(t *testing.T) {
		builder := flagsBuilder{
			flags: map[string]interface{}{},
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
		builder := flagsBuilder{
			flags: map[string]interface{}{
				"rollme": "reason=test",
			},
		}

		_ = builder.withRollme("another-reason=test-2")

		expectedFlags := map[string]interface{}{
			"rollme": "reason=test,another-reason=test-2",
		}

		flags, err := builder.Build()
		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})
}
