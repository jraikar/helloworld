// Aerostation API service
//
// A space station except way more Aero-dynamic
//
//
//     Schemes: http
//     BasePath: /v1
//     Version: 1.0.0
//     Contact: sbhagat@aerospike.com
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     securityDefinitions:
//       BasicAuth:
//         type : basic
//
//     security:
//       BasicAuth:
// swagger:meta
package main

import (
	"fmt"

	"github.com/aerospike/aerostation/api-server/cmd"
)

//go:generate swagger generate spec -o ./swaggerui/swagger.json -m
func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
}
