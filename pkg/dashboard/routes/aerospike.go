package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aerospike/aerostation/api-server/pkg/requests"
	v1 "github.com/aerospike/aerostation/api/v1"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AerospikeRouter struct {
	Client client.Client
}

func NewAerospikeRouter(client client.Client) *AerospikeRouter {
	return &AerospikeRouter{Client: client}
}

func (c *AerospikeRouter) GetAllAerospikeClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	databases := &v1.AeroDatabaseList{}
	err := c.Client.List(context.Background(), databases)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(w).Encode("unable to get aerospike clusters " + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
	}
	var response []string
	for _, v := range databases.Items {
		response = append(response, v.GetName())
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
}

func (c *AerospikeRouter) GetAerospikeClusterSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	dbName := vars["name"]
	databases := &v1.AeroDatabaseList{}
	err := c.Client.List(context.Background(), databases)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(w).Encode("unable to get database " + dbName + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
	}

	for _, v := range databases.Items {
		if dbName == v.Name {
			err := json.NewEncoder(w).Encode(v.Spec)
			if err != nil {
				fmt.Printf("error %s\n", err.Error())
			}
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	err = json.NewEncoder(w).Encode("database not found ")
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
}

func (c *AerospikeRouter) GetAerospikeClusterStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	dbName := vars["name"]
	databases := &v1.AeroDatabaseList{}
	err := c.Client.List(context.Background(), databases)
	if err != nil {
		err := json.NewEncoder(w).Encode("unable to get database " + dbName + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
	}

	for _, v := range databases.Items {
		if dbName == v.Name {
			err := json.NewEncoder(w).Encode(v.Status)
			if err != nil {
				fmt.Printf("error %s\n", err.Error())
			}
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	err = json.NewEncoder(w).Encode("cluster not found ")
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
}

func (c *AerospikeRouter) CreateAerospikeCluster(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var input requests.CreateAerospikeClusterAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("[ERROR] : ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("Unable to create database" + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	database := &v1.AeroDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1.AeroDatabaseSpec{
			Name:      input.Name,
			Namespace: metav1.NamespaceDefault,
			Cluster: v1.ClusterKey{
				Name:      input.KubernetesClusterName,
				Namespace: input.KubernetesClusterNamespace, //TODO: pass in namespace
			},
			DatabaseType: "memory",
			Options: v1.DatabaseOptions{
				Replicas: input.Options.Replicas,
			},
		},
	}

	fmt.Printf("[DEBUG] creating cluster %s \n", input.Name)

	if err := c.Client.Create(context.TODO(), database); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		err := json.NewEncoder(w).Encode("Unable to create cluster")
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	err := json.NewEncoder(w).Encode("Cluster created successfully")
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
}

func (c *AerospikeRouter) DeleteAerospikeCluster(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	clusterName := vars["name"]

	if clusterName == "" {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("Invalid input")
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	cluster := &v1.AeroDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: metav1.NamespaceDefault,
		},
	}

	fmt.Printf("[DEBUG] deleting aerospike cluster %s \n", clusterName)

	ctx := context.Background()
	if err := c.Client.Get(ctx, types.NamespacedName{Name: clusterName, Namespace: metav1.NamespaceDefault}, cluster); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("unable to delete aerospike cluster " + clusterName + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}

		return
	}

	if err := c.Client.Delete(context.Background(), cluster); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("Unable to delete aerospike cluster " + clusterName + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	err := json.NewEncoder(w).Encode("Cluster deleted successfully")
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
}

func (c *AerospikeRouter) GetAllAerospikeClustersForKubernetesCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	kubeID := vars["kubernetesClusterId"]
	databases := &v1.AeroDatabaseList{}
	err := c.Client.List(context.Background(), databases)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(w).Encode("unable to get databases for kubernetes cluster " + kubeID + ". Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	var response []string
	for _, v := range databases.Items {
		if kubeID == v.ClusterName {
			response = append(response, v.GetName())
		}
	}

	if len(response) < 1 {
		w.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(w).Encode("cluster not found ")
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
}

func (c *AerospikeRouter) UpdateAerospikeCluster(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	dbName := vars["name"]

	var input v1.DatabaseOptions
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("[ERROR] : ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("Unable to create database")
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	db := &v1.AeroDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbName,
			Namespace: metav1.NamespaceDefault,
		},
	}

	if err := c.Client.Get(context.Background(), client.ObjectKey{
		Name:      dbName,
		Namespace: metav1.NamespaceDefault,
	}, db); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("Unable to update cluster " + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	patchHelper, err := patch.NewHelper(db, c.Client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		err := json.NewEncoder(w).Encode("Unable to update cluster " + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	db.Spec.Options = input
	fmt.Printf("[DEBUG] updating cluster %s \n", dbName)

	if err = patchHelper.Patch(context.Background(), db); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		err := json.NewEncoder(w).Encode("Unable to update cluster " + "|  Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	err = json.NewEncoder(w).Encode("Cluster updated successfully")
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
}
