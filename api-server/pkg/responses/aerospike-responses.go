package responses

import v1 "github.com/aerospike/aerostation/api/v1"

// swagger:model
type AerospikeClusterResponse struct {
	Cluster v1.AeroDatabaseSpec   `json:"cluster,omitempty"`
	Status  v1.AeroDatabaseStatus `json:"status,omitempty"`
}
