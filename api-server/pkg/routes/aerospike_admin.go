package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aerospike/aerostation/api-server/pkg/requests"
	"github.com/aerospike/aerostation/api-server/pkg/responses"
	pb "github.com/aerospike/aerostation/capi-api/messages"

	v1 "github.com/aerospike/aerostation/api/v1"
	"github.com/gorilla/mux"
)

type AerospikeAdminRouter struct {
	Client pb.AerostationAerospikeAPIClient
}

func NewAerospikeAdminRouter(client pb.AerostationAerospikeAPIClient) *AerospikeAdminRouter {
	return &AerospikeAdminRouter{Client: client}
}

//getAerospikeCluster - get a aerospike cluster
// swagger:operation GET /api/v1/admin/aerospike/clusters/{namespace}/{name} admin aerospike
// ---
//  summary: Get all aerospike database clusters
//  description: Get names of all the workload clusters in array of string
//  operationId: adminGetAerospikeCluster
//  produces:
//    - application/json
//  parameters:
//   - name: namespace
//     in: path
//     description: Namespace of the cluster
//     required: true
//     type: string
//   - name: name
//     in: path
//     description: Name of the cluster
//     required: true
//     type: string
//  responses:
//    '200':
//      description: success
//      schema:
//         $ref: '#/definitions/AerospikeClusterResponse'
//    '401':
//      description: Unauthorized
//    '404':
//      description: cluster not found
func (a *AerospikeAdminRouter) GetAerospikeCluster(w http.ResponseWriter, r *http.Request) {
	// TODO: Get User namespaces?

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	clusterName := vars["name"]
	clusterNamespace := vars["namespace"]
	//TODO: Get user namespace/scope/claim
	res, err := a.Client.GetCluster(context.Background(), &pb.GetAerospikeClusterRequest{
		AerospikeNamespace: clusterNamespace,
		AerospikeName:      clusterName,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("unable to get cluster " + clusterName + "| Reason : " + err.Error())
		return
	}

	clusterReader := bytes.NewReader(res.Cluster)
	var cluster v1.AeroDatabaseSpec
	if err := json.NewDecoder(clusterReader).Decode(&cluster); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Internal Server Error")
		return
	}

	statusReader := bytes.NewReader(res.Status)
	var status v1.AeroDatabaseStatus
	if err := json.NewDecoder(statusReader).Decode(&status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Internal Server Error")
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("unable to get cluster " + clusterName + "| Reason : " + err.Error())
		return
	}

	json.NewEncoder(w).Encode(responses.AerospikeClusterResponse{Cluster: cluster, Status: status})
}

//GetAllAerospikeClusters - get list of all the database clusters
// swagger:operation GET /api/v1/admin/aerospike/clusters admin aerospike
// ---
//  summary: Get all aerospike database clusters
//  description: Get names of all the workload clusters in array of string
//  operationId: getAllAerospikeClusters
//  produces:
//    - application/json
//  responses:
//    '200':
//      description: success
//      schema:
//       type: array
//       items:
//	       $ref: '#/definitions/NamespacedName'
//       example: [{"namespace":"cluster-namespace","name": "cluster-name"}, {"namespace":"cluster2-namespace","name": "cluster2-name"}]
//    '401':
//      description: Unauthorized
//    '500':
//      description: Internal server error
func (a *AerospikeAdminRouter) GetAllAerospikeClusters(w http.ResponseWriter, r *http.Request) {
	// TODO: Get User namespaces?

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//TODO: get user's namespace's/claims
	res, err := a.Client.GetAllClusters(context.Background(), &pb.GetAllAerospikeClustersRequest{})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("Internal server error " + "| Reason : " + err.Error())
	}

	json.NewEncoder(w).Encode(json.RawMessage(res.Clusters))
}

//AdminCreateAerospikeCluster - Create Aerospike Database Cluster
// swagger:operation POST /api/v1/admin/aerospike/clusters admin aerospike
// ---
//  summary: Create a aerospike database cluster
//  description: Create a aerospike database cluster
//  operationId: adminCreateAerospikeCluster
//  produces:
//    - application/json
//  parameters:
//   - name: body
//     in: body
//     description: Request body for creating workload cluster
//     required: true
//     schema:
//	     $ref: '#/definitions/CreateAerospikeClusterAdminRequest'
//  responses:
//    '200':
//      description: Aerospike database created successfully
//    '400':
//      description: Unable to create cluster
//    '401':
//      description: Unauthorized
//    '404':
//      description: Unable to create cluster
func (c *AerospikeAdminRouter) AdminCreateAerospikeCluster(w http.ResponseWriter, r *http.Request) {
	// TODO: get user namespace/claim
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var input requests.CreateAerospikeClusterAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("[ERROR] : ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode("Unable to create database" + " | Reason : " + err.Error())
		return
	}

	fmt.Printf("[DEBUG] creating cluster %s \n", input.Name)

	//TODO: get Kubernetes Cluster from Region.
	_, err := c.Client.CreateCluster(context.Background(), &pb.CreateAerospikeClusterRequest{
		KubernetesClusterName:      input.KubernetesClusterName,
		KubernetesClusterNamespace: input.KubernetesClusterNamespace,
		TargetNamespace:            input.TargetNamespace,
		DeployClient:               input.DeployClient,
		Name:                       input.Name,
		Namespace:                  input.Namespace,
		DatabaseType:               "memory",
		Options:                    &pb.DatabaseOptions{Replicas: input.Options.Replicas},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		json.NewEncoder(w).Encode("Unable to create cluster")
		return
	}

	json.NewEncoder(w).Encode("Cluster created successfully")
}

//AdminDeleteAerospikeCluster - delete aerospike database
// swagger:operation DELETE /api/v1/admin/aerospike/clusters/{namespace}/{name} admin aerospike
// ---
//  summary: Delete a aerospike database cluster
//  description: Delete a aerospike database cluster
//  operationId: adminDeleteAerospikeCluster
//  parameters:
//   - name: name
//     in: path
//     description: Name of the aerospike cluster to delete
//     required: true
//     type: string
//   - name: namespace
//     in: path
//     description: Namespace of the aerospike cluster to delete
//     required: true
//     type: string
//  produces:
//    - application/json
//  responses:
//    '200':
//      description: Aerospike cluster deleted successfully
//    '400':
//      description: Unable to delete cluster
//    '401':
//      description: Unauthorized
func (c *AerospikeAdminRouter) AdminDeleteAerospikeCluster(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	clusterName := vars["name"]
	clusterNamespaceName := vars["namespace"]

	if clusterName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid input")
		return
	}

	fmt.Printf("[DEBUG] deleting aerospike cluster %s \n", clusterName)

	ctx := context.Background()

	_, err := c.Client.DeleteCluster(ctx, &pb.DeleteAerospikeClusterRequest{
		AerospikeName:      clusterName,
		AerospikeNamespace: clusterNamespaceName,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode("unable to delete aerospike cluster bob" + clusterName + " | Reason : " + err.Error())
		return
	}

	json.NewEncoder(w).Encode("Cluster deleted successfully")
}

//AdminUpdateAerospikeCluster - update aerospike cluster
// swagger:operation PATCH /api/v1/admin/aerospike/clusters/{namespace}/{name} admin aerospike
// ---
//  summary: Update a aerospike cluster
//  description: Update a aerospike cluster
//  operationId: AdminUpdateAerospikeCluster
//  produces:
//    - application/json
//  parameters:
//   - name: name
//     in: path
//     description: Name of the aerospike cluster to delete
//     required: true
//     type: string
//   - name: namespace
//     in: path
//     description: Namespace of the aerospike cluster to delete
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
//      description: Bad request
//    '401':
//      description: Unauthorized
//    '500':
//      description: Unable to find cluster
func (a *AerospikeAdminRouter) AdminUpdateAerospikeCluster(w http.ResponseWriter, r *http.Request) {
	//TODO: get user claim/namespace
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	dbName := vars["name"]

	dbNamespace := vars["namespace"]

	var input v1.DatabaseOptions
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("[ERROR] : ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode("Bad request")
		return
	}

	_, err := a.Client.UpdateCluster(context.Background(), &pb.UpdateAerospikeClusterRequest{
		ClusterName:      dbName,
		ClusterNamespace: dbNamespace,
		Options:          &pb.DatabaseOptions{Replicas: input.Replicas},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Unable to update cluster " + "| Reason : " + err.Error())
		return
	}

	json.NewEncoder(w).Encode("Cluster updated successfully")
	return
}
