package validation

import (
	"errors"

	"github.com/aerospike/aerostation/api-server/pkg/requests"
)

// ValidateCreateAerospikeClusterRequest validates a
func ValidateCreateAerospikeClusterRequest(request requests.CreateAerospikeClusterRequest) error {
	//TODO: check if this is a real region
	if request.Region == "" {
		return errors.New("region cannot be empty")
	}

	if request.Name == "" {
		return errors.New("rame cannot be empty")
	}

	return nil
}
