package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aerospike/aerostation/api-server/pkg/responses"
	pb "github.com/aerospike/aerostation/capi-api/messages"

	v1 "github.com/aerospike/aerostation/api/v1"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubernetesRouter struct {
	Client pb.AerostationKubernetesAPIClient
}

func NewKubernetesRouter(client pb.AerostationKubernetesAPIClient) *KubernetesRouter {
	return &KubernetesRouter{Client: client}
}

// GetAllKubernetesClusters - get list of all the workload clusters
// swagger:operation GET /api/v1/admin/kubernetes/clusters admin kubernetes
// ---
//  summary: Get all clusters
//  description: Get names of all the workload clusters in array of string
//  operationId: getAllClusters
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
func (k *KubernetesRouter) GetAllKubernetesClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// TODO: get namespace of user
	res, err := k.Client.GetClusters(context.Background(), &pb.GetKubernetesClustersRequest{
		Namespace: "default",
	}) //.List(context.Background(), clusters)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("unable to get all kubernetes clusters " + "| Reason : " + err.Error())
		return
	}

	json.NewEncoder(w).Encode(json.RawMessage(res.Clusters))
}

// GetKubernetesCluster - Get cluster spec by name
// swagger:operation GET /api/v1/admin/kubernetes/clusters/{namespace}/{name} admin kubernetes
// ---
//  summary: Get cluster spec by name
//  description: Returns cluster spec by name
//  operationId: getClusterSpec
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
//         $ref: '#/definitions/KubernetesClusterResponse'
//    '401':
//      description: Unauthorized
//    '404':
//      description: cluster not found
func (k *KubernetesRouter) GetKubernetesCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	clusterName := vars["name"]
	clusterNamespace := vars["namespace"]
	// TODO: Get user namespace/scope/claim
	res, err := k.Client.GetCluster(context.Background(), &pb.GetKubernetesClusterRequest{
		Namespace: clusterNamespace,
		Name:      clusterName,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("unable to get cluster " + clusterName + "| Reason : " + err.Error())
		return
	}

	clusterReader := bytes.NewReader(res.Cluster)
	var cluster v1.AeroClusterManagerSpec
	if err := json.NewDecoder(clusterReader).Decode(&cluster); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Internal Server Error")
		return
	}

	statusReader := bytes.NewReader(res.Status)
	var status v1.AeroClusterManagerStatus
	if err := json.NewDecoder(statusReader).Decode(&status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Internal Server Error")
		return
	}

	json.NewEncoder(w).Encode(responses.KubernetesClusterResponse{Cluster: cluster, Status: status})
}

// CreateKubernetesCluster - Create cluster
// swagger:operation POST /api/v1/admin/kubernetes/clusters admin kubernetes
// ---
//  summary: Create a workload cluster
//  description: Create a workload cluster
//  operationId: createClusters
//  produces:
//    - application/json
//  parameters:
//   - name: body
//     in: body
//     description: Request body for creating workload cluster
//     required: true
//     schema:
//	     $ref: '#/definitions/ClusterOptions'
//  responses:
//    '200':
//      description: Cluster created successfully
//    '400':
//      description: Bad request
//    '401':
//      description: Unauthorized
//    '500':
//      description: Unable to create cluster
func (k *KubernetesRouter) CreateKubernetesCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var input v1.ClusterOptions
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode("Bad request" + "| Reason : " + err.Error())
		return
	}
	fmt.Printf("ClusterOptions are %v\n", input)

	// TODO: Get user namespace/claim
	fmt.Printf("[DEBUG] creating cluster %s on %s\n", input.Name, input.Provider)
	eksOptions := &pb.CreateEKSKubernetesClusterRequest{}
	dockerOptions := &pb.CreateDockerKubernetesClusterRequest{}
	switch input.Provider {
	case "docker":
		dockerOptions = &pb.CreateDockerKubernetesClusterRequest{}
	case "eks":
		eksOptions = &pb.CreateEKSKubernetesClusterRequest{
			Region:       input.EKSOptions.Region,
			InstanceType: input.EKSOptions.InstanceType,
			SSHKey:       input.EKSOptions.SSHKey,
		}
	default:
		dockerOptions = &pb.CreateDockerKubernetesClusterRequest{}
		eksOptions = &pb.CreateEKSKubernetesClusterRequest{
			Region:       input.EKSOptions.Region,
			InstanceType: input.EKSOptions.InstanceType,
			SSHKey:       input.EKSOptions.SSHKey,
		}
	}

	_, err := k.Client.CreateCluster(context.Background(), &pb.CreateKubernetesClusterRequest{
		Name:        input.Name,
		Provider:    input.Provider,
		Replicas:    input.Replicas,
		KubeVersion: input.KubeVersion,
		Namespace:   metav1.NamespaceDefault,
		Eks:         eksOptions,
		Gke:         &pb.CreateGCPKubernetesClusterRequest{},
		Docker:      dockerOptions,
		Aks:         &pb.CreateAKSKubernetesClusterRequest{},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("Unable to create cluster: %s", err.Error()))
		return
	}

	json.NewEncoder(w).Encode("Cluster created successfully")
}

// Delete - delete cluster
// swagger:operation DELETE /api/v1/admin/kubernetes/clusters/{namespace}/{name} admin kubernetes
// ---
//  summary: Delete a workload cluster
//  description: Delete a workload cluster
//  operationId: deleteClusters
//  parameters:
//   - name: namespace
//     in: path
//     description: Name of the cluster to delete
//     required: true
//     type: string
//   - name: name
//     in: path
//     description: Name of the cluster to delete
//     required: true
//     type: string
//  produces:
//    - application/json
//  responses:
//    '200':
//      description: Cluster deleted successfully
//    '401':
//      description: Unauthorized
//    '500':
//      description: Unable to delete cluster
func (k *KubernetesRouter) DeleteKubernetesCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	clusterName := vars["name"]

	if clusterName == "" || namespace == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid input, either namespace or cluster name is not provided")
		return
	}

	fmt.Printf("[DEBUG] deleting cluster %s \n", clusterName)
	ctx := context.Background()
	if _, err := k.Client.DeleteCluster(ctx, &pb.DeleteKubernetesClusterRequest{Name: clusterName, Namespace: namespace}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Unable to delete cluster " + clusterName + " | Reason : " + err.Error())
		return
	}

	json.NewEncoder(w).Encode("Cluster deleted successfully")
}

// UpdateKubernetesCluster - update cluster
// swagger:operation PATCH /api/v1/admin/kubernetes/clusters admin kubernetes
// ---
//  summary: Update a workload cluster
//  description: Update a workload cluster
//  operationId: updateClusters
//  produces:
//    - application/json
//  parameters:
//   - name: body
//     in: body
//     description: Request body for updating workload cluster
//     required: true
//     schema:
//	     $ref: '#/definitions/ClusterOptions'
//  responses:
//    '200':
//      description: Cluster updated successfully
//    '400':
//      description: Unable to update cluster
//    '401':
//      description: Unauthorized
//    '404':
//      description: Unable to update cluster
func (k *KubernetesRouter) UpdateKubernetesCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var input v1.ClusterOptions
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("[ERROR] :  ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode("Unable to update cluster " + "| Reason : " + err.Error())
		return
	}
	// TODO: get user namespace
	// get AWS manager

	_, err := k.Client.UpdateCluster(context.Background(), &pb.UpdateKubernetesClusterRequest{
		Name:      input.Name,
		Namespace: "default",
		Replicas:  input.Replicas,
		Eks: &pb.CreateEKSKubernetesClusterRequest{
			Region:       input.EKSOptions.Region,
			InstanceType: input.EKSOptions.InstanceType,
			SSHKey:       input.EKSOptions.InstanceType,
		},
		Gke: &pb.CreateGCPKubernetesClusterRequest{},
		Aks: &pb.CreateAKSKubernetesClusterRequest{},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode("Unable to update cluster " + "| Reason : " + err.Error())
		return
	}

	json.NewEncoder(w).Encode("Cluster updated successfully")
}
