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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AeroDatabaseSpec defines the desired state of AeroDatabase
// swagger:model
type AeroDatabaseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ClusterKey of database kubernetes cluster
	// required : true
	Cluster ClusterKey `json:"clusterKey,omitempty"`
	// Name of aerospike cluster
	// required : true
	// example: name
	Name string `json:"name,omitempty"`
	// Namespace of aerospike cluster
	// required : true
	// example : default
	Namespace string `json:"namespace,omitempty"`
	// Namespace on remote cluster where aerospike database will be created
	// required : true
	// example : default
	TargetNamespace string `json:"targetNamespace,omitempty"`
	// flag if set will cause the REST client to be deployed
	// required : false
	// example : true
	DeployClient bool `json:"deployClient,omitempty"`
	// DatabaseType of the aerospike cluster (memory/ssd/performance)
	// required : true
	// example : memory
	DatabaseType string `json:"databaseType,omitempty"`
	// Options of the aerospike cluster, the only thing that can be changed after creation
	// required : true
	Options DatabaseOptions `json:"options,omitempty"`
}

// AeroDatabaseStatus defines the observed state of AeroDatabase
// swagger:model
type AeroDatabaseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase     string `json:"phase,omitempty"`
	LastError string `json:"lastError,omitempty"`
}

const (
	DBPhasePending  = ClusterPhase("Pending")
	DBPhaseDeployed = ClusterPhase("Deployed")
	DBPhaseRunning  = ClusterPhase("Running")
)

type DBPhase string

func (c *AeroDatabaseStatus) SetTypedPhase(p DBPhase) {
	c.Phase = string(p)
}

// ClusterKey is the key to the cluster
// swagger:model
type ClusterKey struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

func (k ClusterKey) ToObjectKey() client.ObjectKey {
	return client.ObjectKey{
		Name:      k.Name,
		Namespace: k.Namespace,
	}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AeroDatabase is the Schema for the aerodatabases API
// swagger:model
type AeroDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AeroDatabaseSpec   `json:"spec,omitempty"`
	Status AeroDatabaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AeroDatabaseList contains a list of AeroDatabase
type AeroDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AeroDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AeroDatabase{}, &AeroDatabaseList{})
}
