package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type conditionMatcher struct {
	expectedState           v1alpha1.State
	expectedConditionStatus metav1.ConditionStatus
}

func ConditionTrueMatcher() gomegatypes.GomegaMatcher {
	return &conditionMatcher{
		expectedState:           v1alpha1.StateReady,
		expectedConditionStatus: metav1.ConditionTrue,
	}
}

func (matcher *conditionMatcher) Match(actual interface{}) (success bool, err error) {
	status, ok := actual.(v1alpha1.DockerRegistryStatus)
	if !ok {
		return false, fmt.Errorf("ConditionMatcher matcher expects an v1alpha1.DockerRegistryStatus")
	}

	if status.State != matcher.expectedState {
		return false, nil
	}

	for _, condition := range status.Conditions {
		if condition.Status != matcher.expectedConditionStatus {
			return false, nil
		}
	}

	return true, nil
}

func (matcher *conditionMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto be in %s state with all %s conditions",
		actual, matcher.expectedState, matcher.expectedConditionStatus)
}

func (matcher *conditionMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto be in %s state with all %s conditions",
		actual, matcher.expectedState, matcher.expectedConditionStatus)
}

type testHelper struct {
	ctx           context.Context
	namespaceName string
}

func (h *testHelper) updateDeploymentStatus(deploymentName string) {
	By(fmt.Sprintf("Updating deployment status: %s", deploymentName))
	var deployment appsv1.Deployment
	Eventually(h.getKubernetesObjectFunc(deploymentName, &deployment)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 30).
		Should(BeTrue())

	deployment.Status.Conditions = append(deployment.Status.Conditions, appsv1.DeploymentCondition{
		Type:    appsv1.DeploymentAvailable,
		Status:  corev1.ConditionTrue,
		Reason:  "test-reason",
		Message: "test-message",
	})
	deployment.Status.Replicas = 1
	Expect(k8sClient.Status().Update(h.ctx, &deployment)).To(Succeed())

	replicaSetName := h.createReplicaSetForDeployment(deployment)

	var replicaSet appsv1.ReplicaSet
	Eventually(h.getKubernetesObjectFunc(replicaSetName, &replicaSet)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 30).
		Should(BeTrue())

	replicaSet.Status.ReadyReplicas = 1
	replicaSet.Status.Replicas = 1
	Expect(k8sClient.Status().Update(h.ctx, &replicaSet)).To(Succeed())

	By(fmt.Sprintf("Deployment status updated: %s", deploymentName))
}

func (h *testHelper) createReplicaSetForDeployment(deployment appsv1.Deployment) string {
	replicaSetName := fmt.Sprintf("%s-replica-set", deployment.Name)
	By(fmt.Sprintf("Creating replica set (for deployment): %s", replicaSetName))
	var (
		trueValue = true
		one       = int32(1)
	)
	replicaSet := appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      replicaSetName,
			Namespace: h.namespaceName,
			Labels:    deployment.Spec.Selector.MatchLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       deployment.Name,
					UID:        deployment.GetUID(),
					Controller: &trueValue,
				},
			},
		},
		// dummy values
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &one,
			Selector: deployment.Spec.Selector,
			Template: deployment.Spec.Template,
		},
	}
	Expect(k8sClient.Create(h.ctx, &replicaSet)).To(Succeed())
	By(fmt.Sprintf("Replica set (for deployment) created: %s", replicaSetName))
	return replicaSetName
}

func (h *testHelper) createDockerRegistry(crName string, spec v1alpha1.DockerRegistrySpec) {
	By(fmt.Sprintf("Creating cr: %s", crName))
	dockerRegistry := v1alpha1.DockerRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crName,
			Namespace: h.namespaceName,
			Labels: map[string]string{
				"operator.kyma-project.io/kyma-name": "test",
			},
		},
		Spec: spec,
	}
	Expect(k8sClient.Create(h.ctx, &dockerRegistry)).To(Succeed())
	By(fmt.Sprintf("Crd created: %s", crName))
}

func (h *testHelper) createNamespace() {
	By(fmt.Sprintf("Creating namespace: %s", h.namespaceName))
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: h.namespaceName,
		},
	}
	Expect(k8sClient.Create(h.ctx, &namespace)).To(Succeed())
	By(fmt.Sprintf("Namespace created: %s", h.namespaceName))
}

func (h *testHelper) getKubernetesObjectFunc(objectName string, obj client.Object) func() (bool, error) {
	return func() (bool, error) {
		return h.getKubernetesObject(objectName, obj)
	}
}

func (h *testHelper) getKubernetesObject(objectName string, obj client.Object) (bool, error) {
	key := types.NamespacedName{
		Name:      objectName,
		Namespace: h.namespaceName,
	}

	err := k8sClient.Get(h.ctx, key, obj)
	if err != nil {
		return false, err
	}
	return true, err
}

func (h *testHelper) listKubernetesObjectFunc(list client.ObjectList) func() (bool, error) {
	return func() (bool, error) {
		return h.listKubernetesObject(list)
	}
}

func (h *testHelper) listKubernetesObject(list client.ObjectList) (bool, error) {
	opts := client.ListOptions{
		Namespace: h.namespaceName,
	}

	err := k8sClient.List(h.ctx, list, &opts)
	if err != nil {
		return false, err
	}
	return true, err
}

func (h *testHelper) getDockerRegistryStatusFunc(name string) func() (v1alpha1.DockerRegistryStatus, error) {
	return func() (v1alpha1.DockerRegistryStatus, error) {
		return h.getDockerRegistryStatus(name)
	}
}

func (h *testHelper) getDockerRegistryStatus(name string) (v1alpha1.DockerRegistryStatus, error) {
	var dockerRegistry v1alpha1.DockerRegistry
	key := types.NamespacedName{
		Name:      name,
		Namespace: h.namespaceName,
	}
	err := k8sClient.Get(h.ctx, key, &dockerRegistry)
	if err != nil {
		return v1alpha1.DockerRegistryStatus{}, err
	}
	return dockerRegistry.Status, nil
}

type dockerRegistryData struct {
	EventPublisherProxyURL *string
	TraceCollectorURL      *string
	EnableInternal         *bool
	registrySecretData
}

type registrySecretData struct {
	Username        *string
	Password        *string
	ServerAddress   *string
	RegistryAddress *string
}

func (d *registrySecretData) toMap() map[string]string {
	result := map[string]string{}
	if d.Username != nil {
		result["username"] = *d.Username
	}
	if d.Password != nil {
		result["password"] = *d.Password
	}
	if d.ServerAddress != nil {
		result["serverAddress"] = *d.ServerAddress
	}
	if d.RegistryAddress != nil {
		result["registryAddress"] = *d.RegistryAddress
	}
	return result
}

func (h *testHelper) createCheckRegistrySecretFunc(registrySecret string, expected registrySecretData) func() (bool, error) {
	return func() (bool, error) {
		var configurationSecret corev1.Secret

		if ok, err := h.getKubernetesObject(
			registrySecret, &configurationSecret); !ok || err != nil {
			return ok, err
		}
		if err := secretContainsSameValues(
			expected.toMap(), configurationSecret); err != nil {
			return false, err
		}
		if err := secretContainsRequired(configurationSecret); err != nil {
			return false, err
		}
		return true, nil
	}
}

func secretContainsRequired(configurationSecret corev1.Secret) error {
	for _, k := range []string{"username", "password", "pullRegAddr", "pushRegAddr", ".dockerconfigjson"} {
		_, ok := configurationSecret.Data[k]
		if !ok {
			return fmt.Errorf("values not propagated (%s is required)", k)
		}
	}
	return nil
}

func secretContainsSameValues(expected map[string]string, configurationSecret corev1.Secret) error {
	for k, expectedV := range expected {
		v, okV := configurationSecret.Data[k]
		if okV == false {
			return fmt.Errorf("values not propagated (%s: nil != %s )", k, expectedV)
		}
		if expectedV != string(v) {
			return fmt.Errorf("values not propagated (%s: %s != %s )", k, string(v), expectedV)
		}
	}
	return nil
}
