package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/patch"

	v1 "github.com/aerospike/aerostation/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesRouter struct {
	Client client.Client
}

func NewKubernetesRouter(client client.Client) *KubernetesRouter {
	return &KubernetesRouter{Client: client}
}

func (c *KubernetesRouter) GetAllKubernetesClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	clusters := &v1.AeroClusterManagerList{}
	err := c.Client.List(context.Background(), clusters)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("unable to get all kubernetes clusters " + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
	}

	var response []string
	for _, v := range clusters.Items {
		response = append(response, v.GetName())
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
}

func (c *KubernetesRouter) GetKubernetesClusterSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	clusterName := vars["name"]
	clusters := &v1.AeroClusterManagerList{}
	err := c.Client.List(context.Background(), clusters)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("unable to get cluster " + clusterName + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
	}

	for _, v := range clusters.Items {
		if clusterName == v.Name {
			err := json.NewEncoder(w).Encode(v.Spec)
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

func (c *KubernetesRouter) GetKubernetesClusterStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	clusterName := vars["name"]
	clusters := &v1.AeroClusterManagerList{}
	err := c.Client.List(context.Background(), clusters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("unable to get cluster " + clusterName + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
	}

	for _, v := range clusters.Items {
		if clusterName == v.Name {
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

func (c *KubernetesRouter) CreateKubernetesCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var input v1.ClusterOptions
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("Unable to create cluster" + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	cluster := &v1.AeroClusterManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1.AeroClusterManagerSpec{
			Name:    input.Name,
			Suspend: false,
			ClusterOptions: v1.ClusterOptions{
				Name:     input.Name,
				Replicas: input.Replicas,
				EKSOptions: &v1.EKSOptions{
					Region:       input.EKSOptions.Region,
					InstanceType: input.EKSOptions.InstanceType,
					SSHKey:       input.EKSOptions.SSHKey,
				},
			},
			ClusterID: v1.NamespacedName{
				Namespace: metav1.NamespaceDefault,
				Name:      input.Name,
			},
			Managed: true,
		},
	}

	fmt.Printf("[DEBUG] creating cluster %s \n", input.Name)

	if err := c.Client.Create(context.TODO(), cluster); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(fmt.Sprintf("Unable to create cluster: %s", err.Error()))
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

func (c *KubernetesRouter) DeleteKubernetesCluster(w http.ResponseWriter, r *http.Request) {
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

	cluster := &v1.AeroClusterManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: metav1.NamespaceDefault,
		},
	}

	fmt.Printf("[DEBUG] deleting cluster %s \n", clusterName)

	ctx := context.Background()
	if err := c.Client.Get(ctx, types.NamespacedName{Name: clusterName, Namespace: metav1.NamespaceDefault}, cluster); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("unable to delete cluster " + clusterName + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	if err := c.Client.Delete(context.Background(), cluster); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("Unable to delete cluster " + clusterName + "| Reason : " + err.Error())
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

func (c *KubernetesRouter) UpdateKubernetesCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var input v1.ClusterOptions
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("[ERROR] :  ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		err := json.NewEncoder(w).Encode("Unable to update cluster " + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	manager := &v1.AeroClusterManager{}
	awsInput := false

	// get AWS manager
	if input.Name != "" {
		awsInput = true
		if err := c.Client.Get(context.Background(), client.ObjectKey{
			Name:      input.Name,
			Namespace: metav1.NamespaceDefault,
		}, manager); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			err := json.NewEncoder(w).Encode("Unable to update cluster " + "| Reason : " + err.Error())
			if err != nil {
				fmt.Printf("error %s\n", err.Error())
			}
			return
		}
	}

	patchHelper, err := patch.NewHelper(manager, c.Client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		err := json.NewEncoder(w).Encode("Unable to update cluster " + "| Reason : " + err.Error())
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
		return
	}

	if awsInput {
		manager.Spec.ClusterOptions.EKSOptions = &v1.EKSOptions{
			InstanceType: input.EKSOptions.InstanceType,
			Region:       input.EKSOptions.Region,
			SSHKey:       input.EKSOptions.SSHKey,
		}
		manager.Spec.ClusterOptions.Replicas = input.Replicas
		fmt.Printf("[DEBUG] updating cluster %s \n", input.Name)
	}

	if err = patchHelper.Patch(context.Background(), manager); err != nil {
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
