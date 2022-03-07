package requests

import v1 "github.com/aerospike/aerostation/api/v1"

// swagger:model
type GetAerospikeClustersByNamespaceRequest struct {
	// Namespace of user
	// required : true
	// example : {UUID}
	Namespace string `json:"namespace"`
}

// CreateAerospikeClusterAdminRequest A request to create an aerospike cluster
// swagger:model
type CreateAerospikeClusterAdminRequest struct {
	// KubernetesClusterName of workload cluster
	// required : true
	// example : cluster-1
	KubernetesClusterName string `json:"kubernetesClusterName,omitempty"`
	// KubernetesClusterNamespace of workload cluster
	// required : true
	// example : default
	KubernetesClusterNamespace string `json:"kubernetesClusterNamespace,omitempty"`
	// Name of aerospike cluster
	// required : true
	// example: name
	Name string `json:"name,omitempty"`
	// Namespace on remote cluster where database will be created
	// required : true
	// example: customer1
	TargetNamespace string `json:"targetNamespace,omitempty"`
	// Flag if set to true will cause the REST Client to be deployed
	// required : false
	// example: true
	DeployClient bool `json:"deployClient,omitempty"`
	// capi Provider type
	// required : true
	// example: docker
	Provider string `json:"provider,omitempty"`
	// kube version
	// required : true
	// example: v1.22.0
	KubeVersion string `json:"kubeversion,omitempty"`
	// replicas
	// required : true
	// example: 1
	Replicas int32 `json:"replicas,omitempty"`
	// Namespace of aerospike cluster
	// required : true
	// example : default
	Namespace string `json:"namespace,omitempty"`
	// DatabaseType of the aerospike cluster (memory/ssd/performance)
	// required : true
	// example : memory
	DatabaseType string `json:"databaseType,omitempty"`
	// Options of the aerospike cluster, the only thing that can be changed after creation
	// required : true
	Options v1.DatabaseOptions `json:"options,omitempty"`
}

// CreateAerospikeClusterRequest A request to create an aerospike cluster
// swagger:model
type CreateAerospikeClusterRequest struct {
	// KubernetesClusterName of workload cluster
	// required : true
	// example : cluster-1
	Region string `json:"region,omitempty"`
	// Name of aerospike cluster
	// required : true
	// example: name
	Name string `json:"name,omitempty"`
	// capi provider type
	// required : true
	// example: docker
	TargetName string `json:"targetNamespace,omitempty"`
	// target database namespace
	// required : true
	// example: default
	Provider string `json:"provider,omitempty"`
	// DatabaseType of the aerospike cluster (memory/ssd/performance)
	// required : true
	// example : memory
	DatabaseType string `json:"databaseType,omitempty"`
	// Options of the aerospike cluster, the only thing that can be changed after creation
	// required : true
	Options v1.DatabaseOptions `json:"options,omitempty"`
}
