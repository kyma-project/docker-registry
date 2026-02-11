package state

import (
	"context"
	"testing"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/flags"
	"github.com/stretchr/testify/require"
)

func Test_sFnLoggingConfiguration(t *testing.T) {
	t.Run("use default log level and format when logging is not specified", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{},
			},
			flagsBuilder: flags.NewBuilder(),
		}

		next, result, err := sFnLoggingConfiguration(context.Background(), nil, s)
		require.Nil(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		configData := flags["configData"].(map[string]interface{})
		logConfig := configData["log"].(map[string]interface{})
		require.Equal(t, "info", logConfig["level"])
		require.Equal(t, "json", logConfig["formatter"])
		accesslog := logConfig["accesslog"].(map[string]interface{})
		require.Equal(t, false, accesslog["disabled"])
	})

	t.Run("use custom log level and format from spec", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					Logging: &v1alpha1.Logging{
						Level:  "debug",
						Format: "text",
					},
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}

		next, result, err := sFnLoggingConfiguration(context.Background(), nil, s)
		require.Nil(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		configData := flags["configData"].(map[string]interface{})
		logConfig := configData["log"].(map[string]interface{})
		require.Equal(t, "debug", logConfig["level"])
		require.Equal(t, "text", logConfig["formatter"])
		accesslog := logConfig["accesslog"].(map[string]interface{})
		require.Equal(t, false, accesslog["disabled"])
	})

	t.Run("disable access logs", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					Logging: &v1alpha1.Logging{
						Level:             "info",
						Format:            "json",
						AccessLogDisabled: true,
					},
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}

		next, result, err := sFnLoggingConfiguration(context.Background(), nil, s)
		require.Nil(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		configData := flags["configData"].(map[string]interface{})
		logConfig := configData["log"].(map[string]interface{})
		require.Equal(t, "info", logConfig["level"])
		require.Equal(t, "json", logConfig["formatter"])
		accesslog := logConfig["accesslog"].(map[string]interface{})
		require.Equal(t, true, accesslog["disabled"])
	})

	t.Run("use defaults when logging spec is set but values are empty", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					Logging: &v1alpha1.Logging{
						Level:  "",
						Format: "",
					},
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}

		next, result, err := sFnLoggingConfiguration(context.Background(), nil, s)
		require.Nil(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)
		configData := flags["configData"].(map[string]interface{})
		logConfig := configData["log"].(map[string]interface{})
		require.Equal(t, "info", logConfig["level"])
		require.Equal(t, "json", logConfig["formatter"])
		accesslog := logConfig["accesslog"].(map[string]interface{})
		require.Equal(t, false, accesslog["disabled"])
	})
}
