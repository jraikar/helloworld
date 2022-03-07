package responses

import v1 "github.com/aerospike/aerostation/api/v1"

// swagger:model
type KubernetesClusterResponse struct {
	Cluster v1.AeroClusterManagerSpec   `json:"cluster,omitempty"`
	Status  v1.AeroClusterManagerStatus `json:"status,omitempty"`
}
