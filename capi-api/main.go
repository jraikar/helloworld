package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/aerospike/aerostation/capi-api/messages"
	"github.com/aerospike/aerostation/capi-api/pkg/servers"
	"github.com/aerospike/aerostation/cmd/aeroctl/cmd"
	"google.golang.org/grpc"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	//tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	//certFile   = flag.String("cert_file", "", "The TLS cert file")
	//keyFile    = flag.String("key_file", "", "The TLS key file")
	//jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	port = flag.Int("port", 10000, "The server port")
)

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("parsing flags")
	client, err := client.New(config, client.Options{Scheme: cmd.Scheme})
	if err != nil {
		fmt.Println("Error creating kube client!")
	}

	fmt.Println("parsing flags")
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	//var opts []grpc.ServerOption
	/*
		if *tls {
			if *certFile == "" {
				*certFile = data.Path("x509/server_cert.pem")
			}
			if *keyFile == "" {
				*keyFile = data.Path("x509/server_key.pem")
			}
			creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
			if err != nil {
				log.Fatalf("Failed to generate credentials %v", err)
			}
			opts = []grpc.ServerOption{grpc.Creds(creds)}
		}
	*/
	//opts = []grpc.ServerOption{grpc.EmptyServerOption{}}
	fmt.Println("creating grpc server")
	grpcServer := grpc.NewServer() //(ops...)
	pb.RegisterAerostationKubernetesAPIServer(grpcServer, servers.NewKubernetesServer(client))
	pb.RegisterAerostationAerospikeAPIServer(grpcServer, servers.NewAerospikeServer(client))
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
