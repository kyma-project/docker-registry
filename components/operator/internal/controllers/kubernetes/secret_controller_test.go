package kubernetes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kyma-project/docker-registry/components/operator/internal/resource"
)

const (
	testBaseNamespace   = "docker-registry"
	testBaseSecretName  = "dockerregistry-config"
	testTargetNamespace = "deployer"
)

// The kyma-module-label-protection ValidatingAdmissionPolicy rejects the copy in these
// namespaces on managed runtimes, on every pass, indefinitely. Asserting that both of
// them are reported is what makes these tests independent of the namespace order, which
// the cached client does not guarantee.
const (
	deniedIstio = "istio-system"
	deniedKyma  = "kyma-system"
)

func TestSecretReconcilerPropagatesToRemainingNamespacesWhenOneIsDenied(t *testing.T) {
	//GIVEN
	c := fixClientDenyingSecretWrites(t, deniedIstio, deniedKyma)
	reconciler := fixSecretReconciler(c)

	//WHEN
	_, err := reconciler.Reconcile(context.TODO(), fixBaseSecretRequest())

	//THEN
	require.NoError(t, err)

	var propagated corev1.Secret
	require.NoError(t,
		c.Get(context.TODO(), client.ObjectKey{Namespace: testTargetNamespace, Name: testBaseSecretName}, &propagated),
		"a denial in an unrelated namespace must not stop propagation to %s", testTargetNamespace)
	require.Equal(t, []byte("secret-password"), propagated.Data["password"])
}

func TestSecretReconcilerKeepsRefreshingWhenANamespaceIsPermanentlyDenied(t *testing.T) {
	//GIVEN
	c := fixClientDenyingSecretWrites(t, deniedIstio, deniedKyma)
	reconciler := fixSecretReconciler(c)

	//WHEN
	result, err := reconciler.Reconcile(context.TODO(), fixBaseSecretRequest())

	//THEN
	// Returning the denial would make controller-runtime discard RequeueAfter and back off
	// instead, so the refresh that covers namespaces created before the base Secret would
	// stop running. The denial is permanent, so there is nothing to retry.
	require.NoError(t, err, "a permanently denied namespace must not poison the reconcile result")
	require.Equal(t, time.Minute, result.RequeueAfter, "the periodic refresh must survive a denial")

	for _, namespace := range []string{deniedIstio, deniedKyma} {
		var denied corev1.Secret
		getErr := c.Get(context.TODO(),
			client.ObjectKey{Namespace: namespace, Name: testBaseSecretName}, &denied)
		require.True(t, apierrors.IsNotFound(getErr),
			"the denied write must not have landed in %s, got %v", namespace, getErr)
	}
}

func TestSecretReconcilerNeverWritesIntoProtectedNamespaces(t *testing.T) {
	//GIVEN
	objects := []client.Object{fixNamespace(testBaseNamespace), fixBaseSecret()}
	for _, namespace := range []string{deniedIstio, testTargetNamespace, deniedKyma} {
		objects = append(objects, fixNamespace(namespace))
	}

	// The platform rejects these writes anyway. Attempting them logs an error on every pass
	// and nothing running there pulls from the in-cluster registry, so the write must not be
	// attempted at all rather than tolerated.
	refuseAttempt := func(obj client.Object) error {
		if _, ok := obj.(*corev1.Secret); !ok {
			return nil
		}
		switch obj.GetNamespace() {
		case deniedIstio, deniedKyma:
			return fmt.Errorf("write attempted in protected namespace %s", obj.GetNamespace())
		}
		return nil
	}

	c := fake.NewClientBuilder().
		WithScheme(fixScheme(t)).
		WithObjects(objects...).
		WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				if err := refuseAttempt(obj); err != nil {
					return err
				}
				return cl.Create(ctx, obj, opts...)
			},
			Update: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				if err := refuseAttempt(obj); err != nil {
					return err
				}
				return cl.Update(ctx, obj, opts...)
			},
		}).
		Build()

	config := fixConfig()
	config.ExcludedNamespaces = DefaultExcludedNamespaces()
	reconciler := NewSecret(c, zap.NewNop().Sugar(), config,
		NewSecretService(resource.New(c, c.Scheme()), config))

	//WHEN
	_, err := reconciler.Reconcile(context.TODO(), fixBaseSecretRequest())

	//THEN
	require.NoError(t, err)

	var propagated corev1.Secret
	require.NoError(t,
		c.Get(context.TODO(), client.ObjectKey{Namespace: testTargetNamespace, Name: testBaseSecretName}, &propagated),
		"consumer namespaces must still be served")
}

func TestSecretReconcilerReturnsRetryableFailures(t *testing.T) {
	//GIVEN
	c := fixClientFailingSecretWrites(t, func(name string) error {
		return apierrors.NewInternalError(fmt.Errorf("etcd is unavailable"))
	}, deniedIstio)
	reconciler := fixSecretReconciler(c)

	//WHEN
	_, err := reconciler.Reconcile(context.TODO(), fixBaseSecretRequest())

	//THEN
	require.Error(t, err, "a failure that may succeed on retry must be returned")
	require.ErrorContains(t, err, deniedIstio)

	var propagated corev1.Secret
	require.NoError(t,
		c.Get(context.TODO(), client.ObjectKey{Namespace: testTargetNamespace, Name: testBaseSecretName}, &propagated),
		"the other namespaces must still be served")
}

func TestSecretReconcilerPropagatesToEveryNamespaceWhenNothingIsDenied(t *testing.T) {
	//GIVEN
	c := fixClientDenyingSecretWrites(t)
	reconciler := fixSecretReconciler(c)

	//WHEN
	result, err := reconciler.Reconcile(context.TODO(), fixBaseSecretRequest())

	//THEN
	require.NoError(t, err)
	require.Equal(t, time.Minute, result.RequeueAfter)

	for _, namespace := range []string{deniedIstio, testTargetNamespace, deniedKyma} {
		var propagated corev1.Secret
		require.NoError(t,
			c.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: testBaseSecretName}, &propagated),
			"secret should have been propagated to %s", namespace)
	}
}

func TestSecretReconcilerSkipsBaseNamespace(t *testing.T) {
	//GIVEN
	c := fixClientDenyingSecretWrites(t)
	reconciler := fixSecretReconciler(c)

	//WHEN
	_, err := reconciler.Reconcile(context.TODO(), fixBaseSecretRequest())

	//THEN
	require.NoError(t, err)

	var base corev1.Secret
	require.NoError(t,
		c.Get(context.TODO(), client.ObjectKey{Namespace: testBaseNamespace, Name: testBaseSecretName}, &base))
	require.Contains(t, base.GetFinalizers(), cfgSecretFinalizerName)
}

func TestSecretServiceHandleFinalizerDeletesFromRemainingNamespacesWhenOneIsDenied(t *testing.T) {
	//GIVEN
	deletedAt := metav1.Now()
	base := fixBaseSecret()
	base.DeletionTimestamp = &deletedAt
	base.Finalizers = []string{cfgSecretFinalizerName}

	objects := []client.Object{fixNamespace(testBaseNamespace), base}
	for _, namespace := range []string{deniedIstio, testTargetNamespace, deniedKyma} {
		objects = append(objects, fixNamespace(namespace), fixPropagatedSecret(namespace))
	}

	c := fake.NewClientBuilder().
		WithScheme(fixScheme(t)).
		WithObjects(objects...).
		WithInterceptorFuncs(interceptor.Funcs{
			Delete: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				if namespace := obj.GetNamespace(); namespace == deniedIstio || namespace == deniedKyma {
					return fixLabelProtectionDenial(obj.GetName())
				}
				return cl.Delete(ctx, obj, opts...)
			},
		}).
		Build()

	svc := NewSecretService(resource.New(c, fixScheme(t)), fixConfig())

	//WHEN
	err := svc.HandleFinalizer(context.TODO(), zap.NewNop().Sugar(), base,
		[]string{deniedIstio, testTargetNamespace, deniedKyma})

	//THEN
	require.Error(t, err)
	require.ErrorContains(t, err, deniedIstio)
	require.ErrorContains(t, err, deniedKyma,
		"every failing namespace must be reported, not just the first one")

	var target corev1.Secret
	getErr := c.Get(context.TODO(),
		client.ObjectKey{Namespace: testTargetNamespace, Name: testBaseSecretName}, &target)
	require.True(t, apierrors.IsNotFound(getErr),
		"a denial in an unrelated namespace must not stop deletion in %s, got %v", testTargetNamespace, getErr)
}

func TestSecretServiceHandleFinalizerRemovesFinalizerWhenEveryDeleteSucceeds(t *testing.T) {
	//GIVEN
	deletedAt := metav1.Now()
	base := fixBaseSecret()
	base.DeletionTimestamp = &deletedAt
	base.Finalizers = []string{cfgSecretFinalizerName}

	c := fake.NewClientBuilder().
		WithScheme(fixScheme(t)).
		WithObjects(fixNamespace(testBaseNamespace), base,
			fixNamespace(testTargetNamespace), fixPropagatedSecret(testTargetNamespace)).
		Build()

	svc := NewSecretService(resource.New(c, fixScheme(t)), fixConfig())

	//WHEN
	err := svc.HandleFinalizer(context.TODO(), zap.NewNop().Sugar(), base, []string{testTargetNamespace})

	//THEN
	require.NoError(t, err)
	require.NotContains(t, base.GetFinalizers(), cfgSecretFinalizerName)
}

func fixSecretReconciler(c client.Client) *SecretReconciler {
	config := fixConfig()
	return NewSecret(c, zap.NewNop().Sugar(), config, NewSecretService(resource.New(c, c.Scheme()), config))
}

func fixClientDenyingSecretWrites(t *testing.T, deniedNamespaces ...string) client.WithWatch {
	t.Helper()
	return fixClientFailingSecretWrites(t, fixLabelProtectionDenial, deniedNamespaces...)
}

func fixClientFailingSecretWrites(t *testing.T, failure func(name string) error, failingNamespaces ...string) client.WithWatch {
	t.Helper()

	denied := map[string]struct{}{}
	for _, namespace := range failingNamespaces {
		denied[namespace] = struct{}{}
	}

	objects := []client.Object{fixNamespace(testBaseNamespace), fixBaseSecret()}
	for _, namespace := range []string{deniedIstio, testTargetNamespace, deniedKyma} {
		objects = append(objects, fixNamespace(namespace))
	}

	isDenied := func(obj client.Object) bool {
		if _, ok := obj.(*corev1.Secret); !ok {
			return false
		}
		_, ok := denied[obj.GetNamespace()]
		return ok
	}

	return fake.NewClientBuilder().
		WithScheme(fixScheme(t)).
		WithObjects(objects...).
		WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				if isDenied(obj) {
					return failure(obj.GetName())
				}
				return cl.Create(ctx, obj, opts...)
			},
			Update: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				if isDenied(obj) {
					return failure(obj.GetName())
				}
				return cl.Update(ctx, obj, opts...)
			},
		}).
		Build()
}

// fixLabelProtectionDenial reproduces the denial the operator gets when it copies the
// base Secret, whose labels carry the kyma-project.io/ prefix, into a protected namespace.
func fixLabelProtectionDenial(name string) error {
	return apierrors.NewForbidden(
		schema.GroupResource{Resource: "secrets"}, name,
		fmt.Errorf("ValidatingAdmissionPolicy 'kyma-module-label-protection' denied request: "+
			"Setting labels with the 'kyma-project.io/' prefix on resources in protected "+
			"namespaces is not allowed"),
	)
}

func fixScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(s))
	return s
}

func fixConfig() Config {
	return Config{
		BaseNamespace:          testBaseNamespace,
		BaseInternalSecretName: testBaseSecretName,
		BaseExternalSecretName: "dockerregistry-config-external",
		ExcludedNamespaces:     []string{testBaseNamespace},
		SecretRequeueDuration:  time.Minute,
	}
}

func fixBaseSecretRequest() reconcile.Request {
	return reconcile.Request{
		NamespacedName: client.ObjectKey{Namespace: testBaseNamespace, Name: testBaseSecretName},
	}
}

func fixBaseSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testBaseSecretName,
			Namespace: testBaseNamespace,
			Labels:    map[string]string{ConfigLabel: CredentialsLabelValue},
		},
		Data: map[string][]byte{"password": []byte("secret-password")},
	}
}

func fixPropagatedSecret(namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testBaseSecretName,
			Namespace: namespace,
			Labels:    map[string]string{ConfigLabel: CredentialsLabelValue},
		},
		Data: map[string][]byte{"password": []byte("secret-password")},
	}
}

func fixNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
}
