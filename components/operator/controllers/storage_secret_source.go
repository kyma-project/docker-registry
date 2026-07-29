package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// storageSecretSource watches the Secrets holding external storage credentials.
//
// Those Secrets belong to the user, so they carry no operator labels and cannot be part of the
// manager cache, which is restricted to the credentials Secrets this operator propagates. They get
// their own metadata-only cache instead: the operator needs the events to notice a rotation, and it
// always reads the credentials themselves straight from the API server.
func storageSecretSource(mgr manager.Manager, mapFunc func(context.Context, client.Object) []ctrl.Request) (source.SyncingSource, error) {
	secretCache, err := ctrlcache.New(mgr.GetConfig(), ctrlcache.Options{
		Scheme:           mgr.GetScheme(),
		Mapper:           mgr.GetRESTMapper(),
		DefaultTransform: keepIdentityOnly,
	})
	if err != nil {
		return nil, err
	}

	if err := mgr.Add(secretCache); err != nil {
		return nil, err
	}

	secretMetadata := &metav1.PartialObjectMetadata{}
	secretMetadata.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))

	eventHandler := handler.TypedEnqueueRequestsFromMapFunc(
		func(ctx context.Context, secret *metav1.PartialObjectMetadata) []ctrl.Request {
			return mapFunc(ctx, secret)
		})

	// informers resync on their own, which repeats an update event for every Secret in the cluster
	// even though nothing changed; only a new resource version means new credentials
	return source.Kind(secretCache, secretMetadata, eventHandler, storageSecretChanged), nil
}

// storageSecretChanged passes the events that can carry new credentials, and drops informer resyncs.
var storageSecretChanged = predicate.TypedResourceVersionChangedPredicate[*metav1.PartialObjectMetadata]{}

// keepIdentityOnly drops everything the storage secret watch does not look at, so that the cache costs
// as little as possible per Secret in the cluster.
func keepIdentityOnly(obj interface{}) (interface{}, error) {
	metadata, ok := obj.(*metav1.PartialObjectMetadata)
	if !ok {
		return obj, nil
	}

	metadata.Labels = nil
	metadata.Annotations = nil
	metadata.ManagedFields = nil
	metadata.OwnerReferences = nil
	metadata.Finalizers = nil

	return metadata, nil
}
