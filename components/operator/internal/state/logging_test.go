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
		require.Equal(t, true, accesslog["disabled"])
	})

	t.Run("use custom log level and format from spec", func(t *testing.T) {
		debugLevel := "debug"
		textFormat := "text"
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					Logging: &v1alpha1.Logging{
						Level:  &debugLevel,
						Format: &textFormat,
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
		require.Equal(t, true, accesslog["disabled"])
	})

	t.Run("enable access logs", func(t *testing.T) {
		infoLevel := "info"
		jsonFormat := "json"
		accessLogEnabled := true
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					Logging: &v1alpha1.Logging{
						Level:            &infoLevel,
						Format:           &jsonFormat,
						AccessLogEnabled: &accessLogEnabled,
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

	t.Run("disable access logs explicitly", func(t *testing.T) {
		infoLevel := "info"
		jsonFormat := "json"
		accessLogEnabled := false
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					Logging: &v1alpha1.Logging{
						Level:            &infoLevel,
						Format:           &jsonFormat,
						AccessLogEnabled: &accessLogEnabled,
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

	t.Run("use defaults when logging spec is set but values are nil", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					Logging: &v1alpha1.Logging{},
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

	t.Run("sanitize console format to text", func(t *testing.T) {
		infoLevel := "info"
		consoleFormat := "console"
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					Logging: &v1alpha1.Logging{
						Level:  &infoLevel,
						Format: &consoleFormat,
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
		require.Equal(t, "text", logConfig["formatter"]) // console should be converted to text
		accesslog := logConfig["accesslog"].(map[string]interface{})
		require.Equal(t, true, accesslog["disabled"])
	})

	t.Run("only set log level without affecting other fields", func(t *testing.T) {
		debugLevel := "debug"
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					Logging: &v1alpha1.Logging{
						Level: &debugLevel,
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
		require.Equal(t, "json", logConfig["formatter"]) // default, not auto-filled from CR
		accesslog := logConfig["accesslog"].(map[string]interface{})
		require.Equal(t, true, accesslog["disabled"]) // default, not auto-filled from CR
	})
}
