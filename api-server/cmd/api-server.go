package cmd

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/aerospike/aerostation/pkg/auth"

	"github.com/gorilla/handlers"

	routes "github.com/aerospike/aerostation/api-server/pkg/routes"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/examples/data"

	pb "github.com/aerospike/aerostation/capi-api/messages"
)

var (
	tls                string
	caFile             string
	serverAddr         string
	serverHostOverride string
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cobra",
		Short: "A generator for Cobra based Applications",
		Long: `Cobra is a CLI library for Go that empowers applications.
		This application is a tool to generate the needed files
		to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting the aerostation api server")
			routes.SetupGoGuardian()

			fmt.Println("Parsing Flags")
			flag.Parse()
			var opts []grpc.DialOption
			if tls != "" {
				fmt.Println("tls?")
				if caFile == "" {
					caFile = data.Path("x509/ca_cert.pem")
				}

				creds, err := credentials.NewClientTLSFromFile(caFile, serverHostOverride)
				if err != nil {
					log.Fatalf("Failed to create TLS credentials %v", err)
				}
				opts = append(opts, grpc.WithTransportCredentials(creds))
			} else {
				fmt.Println("Warning: Running in insecure mode")
				opts = append(opts, grpc.WithInsecure())
			}

			fmt.Println("server addr stuff")
			fmt.Println(serverAddr)
			if serverAddr == "" {
				fmt.Println("it's empty?")
			}
			fmt.Printf("Dialing capi server %s", serverAddr)

			conn, err := grpc.Dial(serverAddr, opts...)
			if err != nil {
				log.Fatalf("fail to dial: %v", err)
			}

			defer conn.Close()

			r := Router(conn)
			fmt.Println("Starting server on the port 8085...")

			credentials := handlers.AllowCredentials()
			methods := handlers.AllowedMethods([]string{"POST", "GET", "HEAD", "OPTIONS"})
			ttl := handlers.MaxAge(3600)
			origins := handlers.AllowedOrigins([]string{"*"})

			log.Fatal(http.ListenAndServe(":8085", handlers.CORS(credentials, methods, ttl, origins)(r)))
		},
	}
	cmd.PersistentFlags().StringVar(&tls, "tls", "", "tls for secure communications")
	cmd.PersistentFlags().StringVar(&caFile, "ca_file", "", "ca file for secure communications")
	cmd.PersistentFlags().StringVar(&serverAddr, "server_addr", "aerostation-capi-service:10000", "Server address")
	cmd.PersistentFlags().StringVar(&serverHostOverride, "server_host_override", "", "server host override")
	return cmd
}

// Execute executes the root command.
func Execute() error {
	return rootCmd().Execute()
}

// Router is exported and used in main.go
func Router(conn grpc.ClientConnInterface) *mux.Router {
	router := mux.NewRouter()

	kubernetesClient := pb.NewAerostationKubernetesAPIClient(conn)
	kubeRouter := routes.NewKubernetesRouter(kubernetesClient)

	aerospikeClient := pb.NewAerostationAerospikeAPIClient(conn)
	aeroAdminRouter := routes.NewAerospikeAdminRouter(aerospikeClient)

	awsRouter := routes.NewAWSValues()

	cloudRouter := routes.GetCloudRouter()

	aeroRouter := routes.NewAerospikeRouter(aerospikeClient)

	// Admin Kubernetes Routes
	//router.HandleFunc("/api/v1/admin/kubernetes/clusters", auth.AuthMiddleware(http.HandlerFunc(kubeRouter.GetAllKubernetesClusters))).Methods("GET") //todo: use auth middleware in future
	router.HandleFunc("/api/v1/admin/kubernetes/clusters", routes.AuthMiddleware(http.HandlerFunc(kubeRouter.GetAllKubernetesClusters))).Methods("GET")
	router.HandleFunc("/api/v1/admin/kubernetes/clusters/{namespace}/{name}", routes.AuthMiddleware(http.HandlerFunc(kubeRouter.GetKubernetesCluster))).Methods("GET")
	router.HandleFunc("/api/v1/admin/kubernetes/clusters", routes.AuthMiddleware(http.HandlerFunc(kubeRouter.CreateKubernetesCluster))).Methods("POST")
	router.HandleFunc("/api/v1/admin/kubernetes/clusters/{namespace}/{name}", routes.AuthMiddleware(http.HandlerFunc(kubeRouter.DeleteKubernetesCluster))).Methods("DELETE")
	router.HandleFunc("/api/v1/admin/kubernetes/clusters", routes.AuthMiddleware(http.HandlerFunc(kubeRouter.UpdateKubernetesCluster))).Methods("PATCH")

	// Admin Aerospike Routes
	router.HandleFunc("/api/v1/admin/aerospike/clusters", routes.AuthMiddleware(http.HandlerFunc(aeroAdminRouter.GetAllAerospikeClusters))).Methods("GET")
	router.HandleFunc("/api/v1/admin/aerospike/clusters/{namespace}/{name}", routes.AuthMiddleware(http.HandlerFunc(aeroAdminRouter.GetAerospikeCluster))).Methods("GET")
	router.HandleFunc("/api/v1/admin/aerospike/clusters", routes.AuthMiddleware(http.HandlerFunc(aeroAdminRouter.AdminCreateAerospikeCluster))).Methods("POST")
	router.HandleFunc("/api/v1/admin/aerospike/clusters/{namespace}/{name}", routes.AuthMiddleware(http.HandlerFunc(aeroAdminRouter.AdminDeleteAerospikeCluster))).Methods("DELETE")
	router.HandleFunc("/api/v1/admin/aerospike/clusters/{namespace}/{name}", routes.AuthMiddleware(http.HandlerFunc(aeroAdminRouter.AdminUpdateAerospikeCluster))).Methods("PATCH")

	// User Aerospike Routes
	router.HandleFunc("/api/v1/aerospike/clusters/", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.GetAerospikeClusters))).Methods("GET")
	router.HandleFunc("/api/v1/aerospike/clusters/{name}", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.GetAerospikeCluster))).Methods("GET")
	router.HandleFunc("/api/v1/aerospike/clusters", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.CreateAerospikeCluster))).Methods("POST")
	router.HandleFunc("/api/v1/aerospike/clusters/{name}", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.DeleteAerospikeCluster))).Methods("DELETE")
	router.HandleFunc("/api/v1/aerospike/clusters/{name}", routes.AuthMiddleware(http.HandlerFunc(aeroRouter.UpdateAerospikeCluster))).Methods("PATCH")

	// Cloud Values
	router.HandleFunc("/api/v1/clouds", routes.AuthMiddleware(http.HandlerFunc(cloudRouter.GetSupportedClouds))).Methods("GET")
	router.HandleFunc("/api/v1/clouds/aws/regions", routes.AuthMiddleware(http.HandlerFunc(awsRouter.GetAWSRegions))).Methods("GET")
	router.HandleFunc("/api/v1/clouds/aws/{region}/instances", routes.AuthMiddleware(http.HandlerFunc(awsRouter.GetRegionInstanceTypes))).Methods("GET")

	router.Use(auth.LoggingMiddleware)
	return router
}
