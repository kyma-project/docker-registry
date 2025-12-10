/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/pkg/errors"
	uberzap "go.uber.org/zap"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	operatorv1alpha1 "github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/controllers"
	"github.com/kyma-project/docker-registry/components/operator/internal/config"
	k8s "github.com/kyma-project/docker-registry/components/operator/internal/controllers/kubernetes"
	"github.com/kyma-project/docker-registry/components/operator/internal/gitrepository"
	"github.com/kyma-project/docker-registry/components/operator/internal/logging"
	"github.com/kyma-project/docker-registry/components/operator/internal/registry"
	internalresource "github.com/kyma-project/docker-registry/components/operator/internal/resource"
	//+kubebuilder:scaffold:imports
)

var (
	scheme         = runtime.NewScheme()
	cleanupTimeout = time.Second * 10
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))

	utilruntime.Must(apiextensionsscheme.AddToScheme(scheme))

	utilruntime.Must(istionetworking.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var probeAddr string
	var configPath string
	var syncPeriod time.Duration

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&configPath, "config-path", "", "Path to config file for dynamic reconfiguration.")
	flag.DurationVar(&syncPeriod, "sync-period", 30*time.Minute, "Sync period for controller cache.")
	flag.Parse()

	cfg, err := config.GetConfig("")
	if err != nil {
		os.Exit(1)
	}

	// Setup logging with atomic level for dynamic reconfiguration
	atomicLevel := uberzap.NewAtomicLevel()
	log, err := logging.ConfigureLogger(cfg.LogLevel, cfg.LogFormat, atomicLevel)
	if err != nil {
		os.Exit(1)
	}

	zapLog := log.WithContext()

	// Setup signal handler once - used for both manager and dynamic config
	signalCtx := ctrl.SetupSignalHandler()

	// Start dynamic reconfiguration in background if config path is provided
	if configPath != "" {
		go logging.ReconfigureOnConfigChange(signalCtx, zapLog, atomicLevel, configPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	zapLog.Info("cleaning orphan deprecated resources")
	err = cleanupOrphanDeprecatedResources(ctx)
	if err != nil {
		zapLog.Error("while removing orphan resources", "error", err)
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: ctrlmetrics.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		Cache: ctrlcache.Options{
			SyncPeriod: &syncPeriod,
		},
		Client: ctrlclient.Options{
			Cache: &ctrlclient.CacheOptions{
				DisableFor: []ctrlclient.Object{
					&corev1.Secret{},
					&corev1.ConfigMap{},
				},
			},
		},
	})
	if err != nil {
		zapLog.Error("unable to start manager", "error", err)
		os.Exit(1)
	}

	reconciler := controllers.NewDockerRegistryReconciler(
		mgr.GetClient(), mgr.GetConfig(),
		mgr.GetEventRecorderFor("dockerregistry-operator"),
		zapLog,
		cfg.ChartPath,
	)

	configKubernetes := k8s.Config{
		BaseNamespace:                 "kyma-system",
		BaseInternalSecretName:        registry.InternalAccessSecretName,
		BaseExternalSecretName:        registry.ExternalAccessSecretName,
		ExcludedNamespaces:            []string{"kyma-system"},
		ConfigMapRequeueDuration:      time.Minute,
		SecretRequeueDuration:         time.Minute,
		ServiceAccountRequeueDuration: time.Minute,
	}

	resourceClient := internalresource.New(mgr.GetClient(), scheme)
	secretSvc := k8s.NewSecretService(resourceClient, configKubernetes)

	if err = reconciler.SetupWithManager(mgr); err != nil {
		zapLog.Error("unable to create controller", "controller", "DockerRegistry", "error", err)
		os.Exit(1)
	}

	if err := k8s.NewNamespace(mgr.GetClient(), zapLog, configKubernetes, secretSvc).
		SetupWithManager(mgr); err != nil {
		zapLog.Error("unable to create Namespace controller", "error", err)
		os.Exit(1)
	}

	if err := k8s.NewSecret(mgr.GetClient(), zapLog, configKubernetes, secretSvc).
		SetupWithManager(mgr); err != nil {
		zapLog.Error("unable to create Secret controller", "error", err)
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		zapLog.Error("unable to set up health check", "error", err)
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		zapLog.Error("unable to set up ready check", "error", err)
		os.Exit(1)
	}

	zapLog.Info("starting manager")
	if err := mgr.Start(signalCtx); err != nil {
		zapLog.Error("problem running manager", "error", err)
		os.Exit(1)
	}
}

func cleanupOrphanDeprecatedResources(ctx context.Context) error {
	// We are going to talk to the API server _before_ we start the manager.
	// Since the default manager client reads from cache, we will get an error.
	// So, we create a "serverClient" that would read from the API directly.
	// We only use it here, this only runs at start up, so it shouldn't be to much for the API
	serverClient, err := ctrlclient.New(ctrl.GetConfigOrDie(), ctrlclient.Options{
		Scheme: scheme,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create a server client")
	}

	return gitrepository.Cleanup(ctx, serverClient)
}
