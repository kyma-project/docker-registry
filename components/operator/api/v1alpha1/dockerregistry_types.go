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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Endpoint struct {
	Endpoint string `json:"endpoint"`
}

// DockerRegistrySpec defines the desired state of DockerRegistry
type DockerRegistrySpec struct {
	// Sets the timeout for the Function health check. The default value in seconds is `10`
	HealthzLivenessTimeout string   `json:"healthzLivenessTimeout,omitempty"` //TODO: probably it was only used by serverless so it could be removed
	Storage                *Storage `json:"storage,omitempty"`
}

type Storage struct {
	Azure *StorageAzure `json:"azure,omitempty"`
}

type StorageAzure struct {
	SecretName string `json:"secretName"`
}

type StorageAzureSecrets struct {
	AccountName string
	AccountKey  string
	Container   string
}

type State string

type Served string

type ConditionReason string

type ConditionType string

const (
	StateReady      State = "Ready"
	StateProcessing State = "Processing"
	StateWarning    State = "Warning"
	StateError      State = "Error"
	StateDeleting   State = "Deleting"

	ServedTrue  Served = "True"
	ServedFalse Served = "False"

	// installation and deletion details
	ConditionTypeInstalled = ConditionType("Installed")

	// prerequisites and soft dependencies
	ConditionTypeConfigured = ConditionType("Configured")

	// deletion
	ConditionTypeDeleted = ConditionType("Deleted")

	ConditionReasonConfiguration    = ConditionReason("Configuration")
	ConditionReasonConfigurationErr = ConditionReason("ConfigurationErr")
	ConditionReasonConfigured       = ConditionReason("Configured")
	ConditionReasonInstallation     = ConditionReason("Installation")
	ConditionReasonInstallationErr  = ConditionReason("InstallationErr")
	ConditionReasonInstalled        = ConditionReason("Installed")
	ConditionReasonDuplicated       = ConditionReason("Duplicated")
	ConditionReasonDeletion         = ConditionReason("Deletion")
	ConditionReasonDeletionErr      = ConditionReason("DeletionErr")
	ConditionReasonDeleted          = ConditionReason("Deleted")

	Finalizer = "dockerregistry-operator.kyma-project.io/deletion-hook"
)

type DockerRegistryStatus struct {
	SecretName string `json:"secretName,omitempty"`

	Storage string `json:"storage,omitempty"`

	HealthzLivenessTimeout string `json:"healthzLivenessTimeout,omitempty"`

	// State signifies current state of DockerRegistry.
	// Value can be one of ("Ready", "Processing", "Error", "Deleting").
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Processing;Deleting;Ready;Error;Warning
	State State `json:"state,omitempty"`

	// Served signifies that current DockerRegistry is managed.
	// Value can be one of ("True", "False").
	// +kubebuilder:validation:Enum=True;False
	Served Served `json:"served"`

	// Conditions associated with CustomStatus.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen=true

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Configured",type="string",JSONPath=".status.conditions[?(@.type=='Configured')].status"
//+kubebuilder:printcolumn:name="Installed",type="string",JSONPath=".status.conditions[?(@.type=='Installed')].status"
//+kubebuilder:printcolumn:name="generation",type="integer",JSONPath=".metadata.generation"
//+kubebuilder:printcolumn:name="age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="state",type="string",JSONPath=".status.state"

// DockerRegistry is the Schema for the dockerregistry API
type DockerRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DockerRegistrySpec   `json:"spec,omitempty"`
	Status DockerRegistryStatus `json:"status,omitempty"`
}

func (s *DockerRegistry) UpdateConditionFalse(c ConditionType, r ConditionReason, err error) {
	condition := metav1.Condition{
		Type:               string(c),
		Status:             "False",
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            err.Error(),
	}
	meta.SetStatusCondition(&s.Status.Conditions, condition)
}

func (s *DockerRegistry) UpdateConditionUnknown(c ConditionType, r ConditionReason, msg string) {
	condition := metav1.Condition{
		Type:               string(c),
		Status:             "Unknown",
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            msg,
	}
	meta.SetStatusCondition(&s.Status.Conditions, condition)
}

func (s *DockerRegistry) UpdateConditionTrue(c ConditionType, r ConditionReason, msg string) {
	condition := metav1.Condition{
		Type:               string(c),
		Status:             "True",
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            msg,
	}
	meta.SetStatusCondition(&s.Status.Conditions, condition)
}

func (s *DockerRegistry) IsServedEmpty() bool {
	return s.Status.Served == ""
}

//+kubebuilder:object:root=true

// DockerRegistryList contains a list of DockerRegistry
type DockerRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DockerRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DockerRegistry{}, &DockerRegistryList{})
}
