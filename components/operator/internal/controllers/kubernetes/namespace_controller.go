package kubernetes

import (
	"context"
	goerrors "errors"
	"fmt"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type NamespaceReconciler struct {
	Log       *zap.SugaredLogger
	client    client.Client
	config    Config
	secretSvc SecretService
}

func NewNamespace(client client.Client, log *zap.SugaredLogger, config Config,
	secretSvc SecretService) *NamespaceReconciler {
	return &NamespaceReconciler{
		client:    client,
		Log:       log,
		config:    config,
		secretSvc: secretSvc,
	}
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("namespace-controller").
		For(&corev1.Namespace{}).
		WithEventFilter(r.predicate()).
		Complete(r)
}

func (r *NamespaceReconciler) predicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			namespace, ok := e.Object.(*corev1.Namespace)
			if !ok {
				return false
			}
			return !isExcludedNamespace(namespace.Name, r.config.BaseNamespace, r.config.ExcludedNamespaces)
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

// Reconcile reads that state of the cluster for a Namespace object and updates other resources based on it
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete

func (r *NamespaceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	instance := &corev1.Namespace{}
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger := r.Log.With("name", instance.GetName())

	logger.Debug(fmt.Sprintf("Updating Secret in namespace '%s'", instance.GetName()))
	var errs []error
	secrets, err := r.secretSvc.GetBase(ctx)
	if err != nil {
		errs = append(errs, err)
	}
	for _, secret := range secrets {
		err = r.secretSvc.UpdateNamespace(ctx, logger, instance.GetName(), &secret)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return ctrl.Result{}, goerrors.Join(errs...)
}
