package routes

import (
	"encoding/json"
	"net/http"

	"github.com/aerospike/aerostation/api-server/pkg/responses"
)

type CloudsRouter struct {
}

func GetCloudRouter() *CloudsRouter {
	return &CloudsRouter{}
}

// GetSupportedClouds - get list of all the workload clusters
// swagger:operation GET /api/v1/clouds clouds cloud
// ---
//  summary: Get all supported clouds
//  description: get all supported clouds
//  operationId: GetSupportedClouds
//  produces:
//    - application/json
//  responses:
//    '200':
//      description: success
//      schema:
//       type: array
//       items:
//	       $ref: '#/definitions/Cloud'
//       example: [{"name":"AWS","image":"https://some-s3-bucket"}]
//    '401':
//      description: Unauthorized
//    '500':
//      description: Internal server error
func (c *CloudsRouter) GetSupportedClouds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(
		[]responses.Cloud{
			{
				Name:  "AWS",
				Image: "https://some-s3-bucket",
			},
		})
}
