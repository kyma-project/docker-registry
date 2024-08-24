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

// DockerRegistrySpec defines the desired state of DockerRegistry
type DockerRegistrySpec struct {
	// Storage defines the storage configuration ( filesystem / s3 / azure / gcs / btpObjectStore ).
	Storage *Storage `json:"storage,omitempty"`

	// ExternalAccess defines the external access configuration.
	ExternalAccess *ExternalAccess `json:"externalAccess,omitempty"`
}

type ExternalAccess struct {
	// Enable indicates whether the external access is enabled.
	// default: false
	Enabled *bool `json:"enabled,omitempty"`

	// Gateway defines gateway name (in format: <namespace>/<name>)
	// default: kyma-system/kyma-gateway
	Gateway *string `json:"gateway,omitempty"`

	// Host defines address under which registry will be exposed
	// should fit to at least one server defined in the gateway
	Host *string `json:"host,omitempty"`
}

type Storage struct {
	Azure          *StorageAzure          `json:"azure,omitempty"`
	S3             *StorageS3             `json:"s3,omitempty"`
	GCS            *StorageGCS            `json:"gcs,omitempty"`
	BTPObjectStore *StorageBTPObjectStore `json:"btpObjectStore,omitempty"`
	PVC            *StoragePVC            `json:"pvc,omitempty"`
}

type StorageAzure struct {
	SecretName string `json:"secretName"`
}

type StorageAzureSecrets struct {
	AccountName string
	AccountKey  string
	Container   string
}

type StorageGCSSecrets struct {
	AccountKey string `json:"accountkey"`
}

type StorageGCS struct {
	Bucket        string `json:"bucket"`
	SecretName    string `json:"secretName,omitempty"`
	Rootdirectory string `json:"rootdirectory,omitempty"`
	Chunksize     int    `json:"chunksize,omitempty"`
}

type StorageS3 struct {
	Bucket         string `json:"bucket"`
	Region         string `json:"region"`
	RegionEndpoint string `json:"regionEndpoint,omitempty"`
	Encrypt        bool   `json:"encrypt,omitempty"`
	Secure         bool   `json:"secure,omitempty"`
	SecretName     string `json:"secretName,omitempty"`
}

type StorageS3Secrets struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type StorageBTPObjectStore struct {
	SecretName string `json:"secretName,omitempty"`
}

type StoragePVC struct {
	Name string `json:"name"`
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

type NetworkAccess struct {
	// Enabled indicates whether the network access is enabled.
	Enabled string `json:"enabled,omitempty"`

	// SecretName is the name of the Secret containing the addresses and auth methods.
	SecretName string `json:"secretName,omitempty"`

	// PushAddress contains an address that can be used to push images to the registry from inside the cluster.
	PushAddress string `json:"pushAddress,omitempty"`

	// PullAddress contains address kubernetes can use to pull images from the registry.
	PullAddress string `json:"pullAddress,omitempty"`
}

type DockerRegistryStatus struct {
	// InternalAccess contains the in-cluster access configuration of the DockerRegistry.
	InternalAccess NetworkAccess `json:"internalAccess,omitempty"`

	// ExternalAccess contains the external access configuration of the DockerRegistry.
	ExternalAccess NetworkAccess `json:"externalAccess,omitempty"`

	// Storage signifies the storage type of DockerRegistry.
	Storage string `json:"storage,omitempty"`

	PVC string `json:"pvc,omitempty"`

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
