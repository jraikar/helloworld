package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/aerospike/aerostation/api-server/pkg/responses"
)

type AwsValues struct {
}

func NewAWSValues() *AwsValues {
	return &AwsValues{}
}

// GetAWSRegions - get list of all the workload clusters
// swagger:operation GET /api/v1/clouds/aws/regions cloud cloud
// ---
//  summary: Get all supported regions in AWS
//  description: Get all of the supported regions in AWS
//  operationId: GetAWSRegions
//  produces:
//    - application/json
//  responses:
//    '200':
//      description: success
//      schema:
//       type: array
//       items:
//	       $ref: '#/definitions/AwsRegion'
//       example: [{"name":"N. Virginia","region":"us-east-1","continent":"North America"},{"name":"N. California","region":"us-west-1","continent":"North America"}]
//    '401':
//      description: Unauthorized
//    '500':
//      description: Internal server error
func (k *AwsValues) GetAWSRegions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	regions := []responses.AwsRegion{
		{
			Name:      "N. Virginia",
			Region:    "us-east-1",
			Continent: "North America",
		},
		{
			Name:      "N. California",
			Region:    "us-west-1",
			Continent: "North America",
		},
	}

	json.NewEncoder(w).Encode(regions)
}

// GetRegionInstanceTypes - get list of all the instances for a region in AWS
// swagger:operation GET /api/v1/clouds/aws/{region}/instances cloud cloud
// ---
//  summary: Get list of all the instances for a region in AWS
//  description: Get list of all the instances for a region in AWS
//  operationId: GetRegionInstanceTypes
//  produces:
//    - application/json
//  parameters:
//   - name: region
//     in: path
//     description: AWS region
//     required: true
//     type: string
//  responses:
//    '200':
//      description: success
//      schema:
//       type: array
//       items:
//	       $ref: '#/definitions/AwsInstanceType'
//       example: [{"name":"A3000","cost":1.3,"flavors":["Memory"]},{"name":"A4000","cost":2.1,"flavors":["Memory","HMA"]}]
//    '401':
//      description: Unauthorized
//    '500':
//      description: Internal server error
func (k *AwsValues) GetRegionInstanceTypes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	region := vars["region"]

	err := k.validateAWSRegion(region)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Region")
		return
	}

	instances := []responses.AwsInstanceType{
		{
			Name:    "A3000",
			Cost:    1.3,
			Flavors: []string{"Memory"},
		},
		{
			Name:    "A4000",
			Cost:    2.1,
			Flavors: []string{"Memory", "HMA"},
		},
	}

	json.NewEncoder(w).Encode(instances)
}

func (k *AwsValues) validateAWSRegion(region string) error {
	s := []string{"us-west-1", "us-east-1"}
	if !contains(s, region) {
		return fmt.Errorf("%s is an invalid region", region)
	}

	return nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
