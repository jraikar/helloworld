package v1

// EKSOptions input for creating workload cluster on EKS
// swagger:model
type EKSOptions struct {
	// Region of workload cluster
	// required : true
	// example: us-east-1
	Region string `json:"region,omitempty"`
	// Type of ec2 instance for workload cluster
	// required : true
	// example: t3.medium
	InstanceType string `json:"instanceType,omitempty"`
	// SSH key name
	// required : true
	// example: default
	SSHKey string `json:"sshKey,omitempty"`
}

type AKSOptions struct{}

type GKEOptions struct{}

type DockerOptions struct{}
