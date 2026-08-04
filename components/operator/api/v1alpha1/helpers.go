package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *DockerRegistry) IsInState(state State) bool {
	return s.Status.State == state
}

func (s *DockerRegistry) IsCondition(conditionType ConditionType) bool {
	return meta.FindStatusCondition(
		s.Status.Conditions, string(conditionType),
	) != nil
}

func (s *DockerRegistry) IsConditionTrue(conditionType ConditionType) bool {
	condition := meta.FindStatusCondition(s.Status.Conditions, string(conditionType))
	return condition != nil && condition.Status == metav1.ConditionTrue
}

// StorageSecretName returns the name of the Secret holding the external storage credentials
// referenced by the spec, or an empty string if no external storage is configured.
func (s *DockerRegistry) StorageSecretName() string {
	storage := s.Spec.Storage
	if storage == nil {
		return ""
	}

	switch {
	case storage.Azure != nil:
		return storage.Azure.SecretName
	case storage.S3 != nil:
		return storage.S3.SecretName
	case storage.GCS != nil:
		return storage.GCS.SecretName
	case storage.BTPObjectStore != nil:
		return storage.BTPObjectStore.SecretName
	default:
		return ""
	}
}

const (
	DefaultEnableInternal = false
	EndpointDisabled      = ""
)
