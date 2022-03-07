/*
package main

import (
	"flag"
	"fmt"
	"k8s.io/client-go/rest"
	"log"
	"net/http"

	"github.com/aerospike/aerostation/cmd/aeroctl/cmd"
	"github.com/aerospike/aerostation/pkg/dashboard/routes"
	"github.com/gorilla/mux"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var kubeconfigfile string

func init() {
	flag.StringVar(&kubeconfigfile, "kc", "", "Path to a kubeconfig")

	flag.Parse()
}

//go:generate swagger generate spec -o ./swaggerui/swagger.json -m
func main() {

	routes.SetupGoGuardian()
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	client, err := client.New(config, client.Options{Scheme: cmd.Scheme})
	if err != nil {
		fmt.Println("Error creating kube client!")
	}

	r := Router(client)
	fmt.Println("Starting server on the port 8085...")
	log.Fatal(http.ListenAndServe(":8085", r))
}

// Router is exported and used in main.go
func Router(client client.Client) *mux.Router {
	router := mux.NewRouter()

	kubeRouter := routes.NewKubernetesRouter(client)

	aeroRouter := routes.NewAerospikeRouter(client)

	// Kubernetes Routes
	router.HandleFunc("/api/v1/kubernetes/clusters", routes.AuthMiddleware(http.HandlerFunc(kubeRouter.GetAllKubernetesClusters))).Methods("GET")
	router.HandleFunc("/api/v1/kubernetes/clusters/{name}/spec", routes.AuthMiddleware(http.HandlerFunc(kubeRouter.GetKubernetesClusterSpec))).Methods("GET")
	router.HandleFunc("/api/v1/kubernetes/clusters/{name}/status", routes.AuthMiddleware(http.HandlerFunc(kubeRouter.GetKubernetesClusterStatus))).Methods("GET")
	router.HandleFunc("/api/v1/kubernetes/clusters", routes.AuthMiddleware(http.HandlerFunc(kubeRouter.CreateKubernetesCluster))).Methods("POST")
	router.HandleFunc("/api/v1/kubernetes/clusters/{name}", routes.AuthMiddleware(http.HandlerFunc(kubeRouter.DeleteKubernetesCluster))).Methods("DELETE")
	router.HandleFunc("/api/v1/kubernetes/clusters", routes.AuthMiddleware(http.HandlerFunc(kubeRouter.UpdateKubernetesCluster))).Methods("PATCH")

	// Aerospike Routes
	router.HandleFunc("/api/v1/aerospike/clusters", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.GetAllAerospikeClusters))).Methods("GET")
	router.HandleFunc("/api/v1/aerospike/clusters/{name}/spec", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.GetAerospikeClusterSpec))).Methods("GET")
	router.HandleFunc("/api/v1/aerospike/clusters/{name}/status", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.GetAerospikeClusterStatus))).Methods("GET")
	router.HandleFunc("/api/v1/aerospike/clusters", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.CreateAerospikeCluster))).Methods("POST")
	router.HandleFunc("/api/v1/aerospike/clusters/{name}", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.DeleteAerospikeCluster))).Methods("DELETE")
	router.HandleFunc("/api/v1/aerospike/clusters/{name}", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.UpdateAerospikeCluster))).Methods("PATCH")
	router.HandleFunc("/api/v1/aerospike/{kubernetesClusterId}/clusters", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.GetAllAerospikeClustersForKubernetesCluster))).Methods("GET")
	return router
}
*/
package dashboard
