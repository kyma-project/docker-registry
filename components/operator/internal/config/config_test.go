package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetConfig_EnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("CHART_PATH", "/env/chart/path")
	defer func() {
		os.Unsetenv("CHART_PATH")
	}()

	// Get config from environment
	cfg, err := GetConfig("")
	require.NoError(t, err)

	// Verify the value from environment
	require.Equal(t, "/env/chart/path", cfg.ChartPath)
}

func TestGetConfig_Defaults(t *testing.T) {
	// Ensure environment variable is not set
	os.Unsetenv("CHART_PATH")

	// Get config with defaults
	cfg, err := GetConfig("")
	require.NoError(t, err)

	// Verify the default value
	require.Equal(t, "/module-chart", cfg.ChartPath)
}
