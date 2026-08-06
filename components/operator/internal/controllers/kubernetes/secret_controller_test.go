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

// deniedBefore and deniedAfter sort on either side of testTargetNamespace so the
// test holds no matter which order the namespace list comes back in.
const (
	deniedBefore = "aaa-protected"
	deniedAfter  = "zzz-protected"
)

func TestSecretReconcilerPropagatesToRemainingNamespacesWhenOneIsDenied(t *testing.T) {
	//GIVEN
	c := fixClientDenyingSecretWrites(t, deniedBefore, deniedAfter)
	reconciler := fixSecretReconciler(c)

	//WHEN
	_, err := reconciler.Reconcile(context.TODO(), fixBaseSecretRequest())

	//THEN
	require.Error(t, err, "a denied namespace must still surface as an error")
	require.ErrorContains(t, err, deniedBefore)
	require.ErrorContains(t, err, deniedAfter,
		"every failing namespace must be reported, not just the first one")

	var propagated corev1.Secret
	require.NoError(t,
		c.Get(context.TODO(), client.ObjectKey{Namespace: testTargetNamespace, Name: testBaseSecretName}, &propagated),
		"a denial in an unrelated namespace must not stop propagation to %s", testTargetNamespace)
	require.Equal(t, []byte("secret-password"), propagated.Data["password"])
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

	for _, namespace := range []string{deniedBefore, testTargetNamespace, deniedAfter} {
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
	for _, namespace := range []string{deniedBefore, testTargetNamespace, deniedAfter} {
		objects = append(objects, fixNamespace(namespace), fixPropagatedSecret(namespace))
	}

	c := fake.NewClientBuilder().
		WithScheme(fixScheme(t)).
		WithObjects(objects...).
		WithInterceptorFuncs(interceptor.Funcs{
			Delete: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				if namespace := obj.GetNamespace(); namespace == deniedBefore || namespace == deniedAfter {
					return fixVAPDenial(obj.GetName())
				}
				return cl.Delete(ctx, obj, opts...)
			},
		}).
		Build()

	svc := NewSecretService(resource.New(c, fixScheme(t)), fixConfig())

	//WHEN
	err := svc.HandleFinalizer(context.TODO(), zap.NewNop().Sugar(), base,
		[]string{deniedBefore, testTargetNamespace, deniedAfter})

	//THEN
	require.Error(t, err)
	require.ErrorContains(t, err, deniedBefore)
	require.ErrorContains(t, err, deniedAfter,
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

	denied := map[string]struct{}{}
	for _, namespace := range deniedNamespaces {
		denied[namespace] = struct{}{}
	}

	objects := []client.Object{fixNamespace(testBaseNamespace), fixBaseSecret()}
	for _, namespace := range []string{deniedBefore, testTargetNamespace, deniedAfter} {
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
					return fixVAPDenial(obj.GetName())
				}
				return cl.Create(ctx, obj, opts...)
			},
			Update: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				if isDenied(obj) {
					return fixVAPDenial(obj.GetName())
				}
				return cl.Update(ctx, obj, opts...)
			},
		}).
		Build()
}

// fixVAPDenial mimics the kyma-module-label-protection ValidatingAdmissionPolicy
// rejecting a write that carries a kyma-project.io/ label.
func fixVAPDenial(name string) error {
	return apierrors.NewForbidden(
		schema.GroupResource{Resource: "secrets"}, name,
		fmt.Errorf("ValidatingAdmissionPolicy 'kyma-module-label-protection' denied request"),
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
