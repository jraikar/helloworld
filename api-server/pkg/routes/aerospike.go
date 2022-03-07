package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aerospike/aerostation/api-server/pkg/requests"
	"github.com/aerospike/aerostation/api-server/pkg/validation"
	v1 "github.com/aerospike/aerostation/api/v1"
	pb "github.com/aerospike/aerostation/capi-api/messages"
	"github.com/gorilla/mux"
)

type AerospikeRouter struct {
	Client pb.AerostationAerospikeAPIClient
}

func NewAerospikeRouter(client pb.AerostationAerospikeAPIClient) *AerospikeRouter {
	return &AerospikeRouter{Client: client}
}

//GetAerospikeClusters - get list of all the database clusters
// swagger:operation GET /api/v1/aerospike/clusters aerospike aerospike
// ---
//  summary: Get all aerospike database clusters
//  description: Get names of all the workload clusters in array of string
//  operationId: getAerospikeClusters
//  produces:
//    - application/json
//  responses:
//    '200':
//      description: success
//      schema:
//       type: array
//       items:
//         type: string
//       example: ["aerospike-cluster-1", "aerospike-cluster-2"]
//    '401':
//      description: Unauthorized
//    '500':
//      description: Internal server error
func (a *AerospikeRouter) GetAerospikeClusters(w http.ResponseWriter, r *http.Request) {
	// TODO: Get User namespaces?

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//TODO: get user's namespace's/claims
	res, err := a.Client.GetClusters(context.Background(), &pb.GetAerospikeClustersByNamespaceRequest{
		AerospikeNamespace: "default",
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("Internal server error " + "| Reason : " + err.Error())
	}

	json.NewEncoder(w).Encode(json.RawMessage(res.Clusters))
}

//GetAerospikeCluster - Get aerospike cluster by name
// swagger:operation GET /api/v1/aerospike/clusters/{name} aerospike aerospike
// ---
//  summary: Get database status by name
//  description: Returns cluster status by name
//  operationId: getAerospikeCluster
//  produces:
//    - application/json
//  parameters:
//   - name: name
//     in: path
//     description: Name of the aerospike cluster
//     required: true
//     type: string
//  responses:
//    '200':
//      description: success
//      schema:
//         $ref: '#/definitions/AeroDatabaseSpec'
//    '401':
//      description: Unauthorized
//    '404':
//      description: cluster not found
func (a *AerospikeRouter) GetAerospikeCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	dbName := vars["name"]

	//TODO: get user namespace/claims
	res, err := a.Client.GetCluster(context.Background(), &pb.GetAerospikeClusterRequest{
		AerospikeName:      dbName,
		AerospikeNamespace: "default",
	})

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("cluster not found ")
	}

	json.NewEncoder(w).Encode(json.RawMessage(res.Cluster))
}

//CreateAerospikeCluster - Create Aerospike Database Cluster
// swagger:operation POST /api/v1/aerospike/clusters aerospike aerospike
// ---
//  summary: Create a aerospike database cluster
//  description: Create a aerospike database cluster
//  operationId: createAerospikeCluster
//  produces:
//    - application/json
//  parameters:
//   - name: body
//     in: body
//     description: Request body for creating workload cluster
//     required: true
//     schema:
//	     $ref: '#/definitions/CreateAerospikeClusterRequest'
//  responses:
//    '200':
//      description: Aerospike database created successfully
//    '400':
//      description: Bad request
//    '401':
//      description: Unauthorized
//    '500':
//      description: Unable to create cluster
func (a *AerospikeRouter) CreateAerospikeCluster(w http.ResponseWriter, r *http.Request) {
	// TODO: get user namespace/claim
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var input requests.CreateAerospikeClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("[ERROR] : ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode("Bad request" + " | Reason : " + err.Error())
		return
	}

	if err := validation.ValidateCreateAerospikeClusterRequest(input); err != nil {
		log.Println("[ERROR] : ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode("Unable to create database" + " | Reason : " + err.Error())
		return
	}

	fmt.Printf("[DEBUG] creating cluster %s \n", input.Name)

	//TODO: get Kubernetes Cluster from Region and get all user info
	_, err := a.Client.CreateCluster(context.Background(), &pb.CreateAerospikeClusterRequest{
		KubernetesClusterName:      "default-kubernetes-cluster",
		KubernetesClusterNamespace: "default",
		Name:                       input.Name,
		Namespace:                  "default",
		DatabaseType:               "memory",
		Options:                    &pb.DatabaseOptions{Replicas: input.Options.Replicas},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode("Unable to create cluster")
		return
	}

	_ = json.NewEncoder(w).Encode("Aerospike database created successfully")
}

//DeleteAerospikeCluster - delete aerospike database
// swagger:operation DELETE /api/v1/aerospike/clusters/{name} aerospike aerospike
// ---
//  summary: Delete a aerospike database cluster
//  description: Delete a aerospike database cluster
//  operationId: deleteAerospikeClusters
//  parameters:
//   - name: name
//     in: path
//     description: Name of the aerospike cluster to delete
//     required: true
//     type: string
//  produces:
//    - application/json
//  responses:
//    '200':
//      description: Aerospike cluster deleted successfully
//    '401':
//      description: Unauthorized
//    '500':
//      description: Internal Server Error
func (a *AerospikeRouter) DeleteAerospikeCluster(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	clusterName := vars["name"]

	fmt.Printf("[DEBUG] deleting aerospike cluster %s \n", clusterName)

	ctx := context.Background()

	_, err := a.Client.DeleteCluster(ctx, &pb.DeleteAerospikeClusterRequest{
		AerospikeName:      clusterName,
		AerospikeNamespace: "default",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Internal Server Error | Reason : " + err.Error())
		return
	}

	json.NewEncoder(w).Encode("Aerospike cluster deleted successfully")
	return
}

//UpdateAerospikeCluster - update aerospike cluster
// swagger:operation PATCH /api/v1/aerospike/clusters/{name} aerospike aerospike
// ---
//  summary: Update a aerospike cluster
//  description: Update a aerospike cluster
//  operationId: updateAerospikeClusters
//  produces:
//    - application/json
//  parameters:
//   - name: name
//     in: path
//     description: Name of the aerospike cluster to update
//     required: true
//     type: string
//   - name: body
//     in: body
//     description: Request body for updating aerospike cluster
//     required: true
//     schema:
//	     $ref: '#/definitions/DatabaseOptions'
//  responses:
//    '200':
//      description: Cluster updated successfully
//    '400':
//      description: Unable to update aerospike cluster
//    '401':
//      description: Unauthorized
//    '500':
//      description: Internal Server Error
func (a *AerospikeRouter) UpdateAerospikeCluster(w http.ResponseWriter, r *http.Request) {
	//TODO: get user claim/namespace

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	dbName := vars["name"]

	var input v1.DatabaseOptions
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("[ERROR] : ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode("Unable to update aerospike cluster")
		return
	}

	_, err := a.Client.UpdateCluster(context.Background(), &pb.UpdateAerospikeClusterRequest{
		ClusterName:      dbName,
		ClusterNamespace: "default",
		Options:          &pb.DatabaseOptions{Replicas: input.Replicas},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Internal Server Error | Reason : " + err.Error())
		return
	}

	json.NewEncoder(w).Encode("Cluster updated successfully")
	return
}
