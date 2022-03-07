/*
Copyright 2021.

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

package v1

import (
	"fmt"
	"net"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AeroClusterManagerSpec defines the desired state of AeroClusterManager
// swagger:model
type AeroClusterManagerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name of Cluster
	// example: cluster-id
	Name string `json:"name,omitempty"`

	// Used to pause reconciliation of object for debugging
	Suspend bool `json:"suspend,omitempty"`

	ClusterOptions ClusterOptions `json:"clusterOptions,omitempty"`

	// ClusterID            client.ObjectKey `json:"clusterId,omitempty"`
	ClusterID            NamespacedName `json:"clusterId,omitempty"`
	ControlPlaneEndpoint APIEndpoint    `json:"controlPlaneEndpoint"`
	Managed              bool           `json:"managed"`
}

// swagger:model
type ClusterOptions struct {
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
	// k8s version
	// required : true
	// example: v1.20.0
	KubeVersion string `json:"kubeversion,omitempty"`
	// Number of replicas of instances
	// required : true
	// min: 1
	// example: 1
	Replicas      int32          `json:"replicas,omitempty"`
	DockerOptions *DockerOptions `json:"dockerOptions,omitempty"`
	EKSOptions    *EKSOptions    `json:"eksOptions,omitempty"`
	AKSOptions    *AKSOptions    `json:"aksOptions,omitempty"`
	GKEOptions    *GKEOptions    `json:"gkeOptions,omitempty"`
}

// swagger:model
type NamespacedName struct {
	// example: default
	Namespace string `json:"namespace,omitempty"`
	// example: name
	Name string `json:"name,omitempty"`
}

func (n NamespacedName) ToObjectKey() client.ObjectKey {
	return client.ObjectKey{
		Namespace: n.Namespace,
		Name:      n.Name,
	}
}

// AeroClusterManagerStatus defines the observed state of AeroClusterManager
type AeroClusterManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Should hold status on all apps installed or not installed
	Phase              string            `json:"phase,omitempty"`
	AerospikeOperator  ApplicationStatus `json:"aerospikeOperator,omitempty"`
	PrometheusExporter ApplicationStatus `json:"prometheusExporter,omitempty"`
}

type ClusterPhase string

const (
	ManagerPhasePending            = ClusterPhase("Pending")
	ManagerPhaseClusterCreating    = ClusterPhase("ClusterCreating")
	ManagerPhaseOperatorInstalling = ClusterPhase("OperatorInstalling")
	ManagerPhaseProvisioned        = ClusterPhase("Provisioned")
	ManagerPhaseUnknown            = ClusterPhase("Unknown")
	ManagerPhaseDeleting           = ClusterPhase("Deleting")
)

func (c *AeroClusterManagerStatus) SetTypedPhase(p ClusterPhase) {
	c.Phase = string(p)
}

type ApplicationStatus struct {
	Running bool `json:"running,omitempty"`
}

// APIEndpoint represents a reachable Kubernetes API endpoint.
type APIEndpoint struct {
	// The hostname on which the API server is serving.
	Host string `json:"host"`

	// The port on which the API server is serving.
	Port int32 `json:"port"`
}

// IsZero returns true if both host and port are zero values.
func (v APIEndpoint) IsZero() bool {
	return v.Host == "" && v.Port == 0
}

// IsValid returns true if both host and port are non-zero values.
func (v APIEndpoint) IsValid() bool {
	return v.Host != "" && v.Port != 0
}

// String returns a formatted version HOST:PORT of this APIEndpoint.
func (v APIEndpoint) String() string {
	return net.JoinHostPort(v.Host, fmt.Sprintf("%d", v.Port))
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AeroClusterManager is the Schema for the aeroclustermanagers API
type AeroClusterManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AeroClusterManagerSpec   `json:"spec,omitempty"`
	Status AeroClusterManagerStatus `json:"status,omitempty"`
}

func (a *AeroClusterManager) GetNamespacedName() *NamespacedName {
	return &NamespacedName{Namespace: a.Namespace, Name: a.Name}
}

//+kubebuilder:object:root=true

// AeroClusterManagerList contains a list of AeroClusterManager
//
// swagger:model
type AeroClusterManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// the schema for the aeroclustermanagers API
	//
	// required: true
	Items []AeroClusterManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AeroClusterManager{}, &AeroClusterManagerList{})
}
