package kubernetes

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type SecretReconciler struct {
	Log    *zap.SugaredLogger
	client client.Client
	config Config
	svc    SecretService
}

func NewSecret(client client.Client, log *zap.SugaredLogger, config Config, secretSvc SecretService) *SecretReconciler {
	return &SecretReconciler{
		client: client,
		Log:    log,
		config: config,
		svc:    secretSvc,
	}
}

func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("secret-controller").
		For(&corev1.Secret{}).
		WithEventFilter(r.predicate()).
		Complete(r)
}

func (r *SecretReconciler) predicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			runtime, ok := e.Object.(*corev1.Secret)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			runtime, ok := e.ObjectNew.(*corev1.Secret)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			runtime, ok := e.Object.(*corev1.Secret)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			runtime, ok := e.Object.(*corev1.Secret)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
	}
}

// Reconcile reads that state of the cluster for a Secret object and makes changes based
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

func (r *SecretReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	instance := &corev1.Secret{}
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger := r.Log.With("namespace", instance.GetNamespace(), "name", instance.GetName())

	namespaces, err := getNamespaces(ctx, r.client, r.config.BaseNamespace, r.config.ExcludedNamespaces)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.svc.HandleFinalizer(ctx, logger, instance, namespaces); err != nil {
		return ctrl.Result{}, err
	}
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	var errs []error
	for _, namespace := range namespaces {
		err = r.svc.UpdateNamespace(ctx, logger, namespace, instance)
		switch {
		case err == nil:
		case apierrors.IsForbidden(err):
			// An admission policy or a quota refuses this namespace and will refuse it on every
			// pass. Returning the error would make controller-runtime discard RequeueAfter in
			// favour of backoff, stopping the refresh the other namespaces rely on.
			logger.Warnf("Skipping namespace %s, the write is refused: %s", namespace, err)
		default:
			errs = append(errs, fmt.Errorf("namespace %s: %w", namespace, err))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: r.config.SecretRequeueDuration}, nil
}
