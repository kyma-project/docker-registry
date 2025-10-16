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

package controllers

import (
	"context"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/components/operator/internal/chart"
	"github.com/kyma-project/docker-registry/components/operator/internal/predicate"
	"github.com/kyma-project/docker-registry/components/operator/internal/state"
	"github.com/kyma-project/docker-registry/components/operator/internal/tracing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// dockerRegistryReconciler reconciles a DockerRegistry object
type dockerRegistryReconciler struct {
	initStateMachine func(*zap.SugaredLogger) state.StateReconciler
	client           client.Client
	log              *zap.SugaredLogger
}

func NewDockerRegistryReconciler(client client.Client, config *rest.Config, recorder record.EventRecorder, log *zap.SugaredLogger, chartPath string) *dockerRegistryReconciler {
	cache := chart.NewSecretManifestCache(client)

	return &dockerRegistryReconciler{
		initStateMachine: func(log *zap.SugaredLogger) state.StateReconciler {
			return state.NewMachine(client, config, recorder, log, cache, chartPath)
		},
		client: client,
		log:    log,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (sr *dockerRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DockerRegistry{}, builder.WithPredicates(predicate.NoStatusChangePredicate{})).
		Watches(&v1alpha1.DockerRegistry{}, &handler.Funcs{
			// retrigger all DockerRegistry CRs reconciliations when one is deleted
			// this should ensure at least one DockerRegistry CR is served
			DeleteFunc: sr.retriggerAllDockerRegistryCRs,
		}).
		Watches(&corev1.Service{}, tracing.ServiceCollectorWatcher()).
		Complete(sr)
}

func (sr *dockerRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := sr.log.With("request", req)
	log.Info("reconciliation started")

	instance, err := state.GetDockerRegistryOrServed(ctx, req, sr.client)
	if err != nil {
		log.Warnf("while getting dockerregistry, got error: %s", err.Error())
		return ctrl.Result{}, errors.Wrap(err, "while fetching dockerregistry instance")
	}
	if instance == nil {
		log.Info("Couldn't find proper instance of dockerregistry")
		return ctrl.Result{}, nil
	}

	r := sr.initStateMachine(log)
	return r.Reconcile(ctx, *instance)
}

func (sr *dockerRegistryReconciler) retriggerAllDockerRegistryCRs(ctx context.Context, e event.DeleteEvent, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	log := sr.log.With("deletion_watcher")

	list := &v1alpha1.DockerRegistryList{}
	err := sr.client.List(ctx, list, &client.ListOptions{})
	if err != nil {
		log.Errorf("error listing dockerregistry objects: %s", err.Error())
		return
	}

	for _, s := range list.Items {
		log.Debugf("retriggering reconciliation for DockerRegistry %s/%s", s.GetNamespace(), s.GetName())
		q.Add(ctrl.Request{NamespacedName: client.ObjectKey{
			Namespace: s.GetNamespace(),
			Name:      s.GetName(),
		}})
	}
}
