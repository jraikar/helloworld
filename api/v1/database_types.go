package v1

// DatabaseOptions the mutable options of an aerospike cluster
// swagger:model
type DatabaseOptions struct {
	// Replicas of the aerospike cluster
	// required : true
	// example : 2
	Replicas int32 `json:"replicas,omitempty"`
}
