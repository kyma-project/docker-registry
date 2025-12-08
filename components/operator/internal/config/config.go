package config

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/vrischmann/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ChartPath string `envconfig:"default=/module-chart" yaml:"chartPath"`
	LogLevel  string `envconfig:"default=info" yaml:"logLevel"`
	LogFormat string `envconfig:"default=json" yaml:"logFormat"`
}

func GetConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

func readConfigFile(cfgPath string) (Config, error) {
	cfg := Config{}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}

// RunOnConfigChange watches for config changes and executes the callback function
// When Kubernetes updates ConfigMaps, it atomically updates symlinks, so we watch the parent directory
func RunOnConfigChange(ctx context.Context, log interface{ Info(...interface{}) }, cfgPath string, onChangeFunc func(Config)) {
	if cfgPath == "" {
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Info("unable to create config watcher", "error", err)
		return
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Info("error closing watcher", "error", err)
		}
	}()

	// Watch the directory containing the config file to catch Kubernetes ConfigMap updates
	// which are done via atomic symlink changes
	configDir := filepath.Dir(cfgPath)

	if err := watcher.Add(configDir); err != nil {
		log.Info("unable to watch config directory", "error", err)
		return
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-watcher.Events:
			// Kubernetes ConfigMap updates trigger Create events on the ..data symlink
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
				cfg, err := readConfigFile(cfgPath)
				if err != nil {
					log.Info("unable to read config", "error", err)
					continue
				}
				onChangeFunc(cfg)
			}
		case err := <-watcher.Errors:
			log.Info("config watcher error", "error", err)
		case <-ticker.C:
			// Periodic check to ensure watcher is still active
		}
	}
}
