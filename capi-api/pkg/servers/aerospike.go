package servers

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "github.com/aerospike/aerostation/api/v1"
	pb "github.com/aerospike/aerostation/capi-api/messages"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AerospikeServer struct {
	pb.AerostationAerospikeAPIServer
	Client client.Client
}

func NewAerospikeServer(client client.Client) *AerospikeServer {

	return &AerospikeServer{Client: client}
}

func (a *AerospikeServer) ValidateAerospikeCluster(ctx context.Context, db *v1.AeroDatabase) error {
	cluster := &v1.AeroClusterManager{}
	err := a.Client.Get(ctx, db.Spec.Cluster.ToObjectKey(), cluster)

	if err != nil {
		return err
	}

	return nil
}

func (a *AerospikeServer) CreateCluster(ctx context.Context, request *pb.CreateAerospikeClusterRequest) (*pb.CreateAerospikeClusterResponse, error) {
	database := &v1.AeroDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      request.Name,
			Namespace: request.Namespace,
		},
		Spec: v1.AeroDatabaseSpec{
			Name:            request.Name,
			Namespace:       request.Namespace,
			TargetNamespace: request.TargetNamespace,
			DeployClient:    request.DeployClient,
			Cluster: v1.ClusterKey{
				Name:      request.KubernetesClusterName,
				Namespace: request.KubernetesClusterNamespace, //TODO: pass in namespace
			},
			DatabaseType: "memory",
			Options: v1.DatabaseOptions{
				Replicas: request.Options.Replicas,
			},
		},
	}

	fmt.Println("Creating Aerospike Cluster")
	fmt.Println("Size of cluster: ", database.Spec.Options.Replicas)

	err := a.ValidateAerospikeCluster(ctx, database)

	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG] creating cluster %s \n", request.Name)

	if err := a.Client.Create(context.TODO(), database); err != nil {
		return nil, status.Errorf(codes.Internal,
			"Internal Service Error: %s", err)
	}

	return &pb.CreateAerospikeClusterResponse{}, nil
}

func (a *AerospikeServer) GetCluster(ctx context.Context, request *pb.GetAerospikeClusterRequest) (*pb.GetAerospikeClusterResponse, error) {
	db := &v1.AeroDatabase{}

	if err := a.Client.Get(context.Background(), client.ObjectKey{
		Name:      request.AerospikeName,
		Namespace: request.AerospikeNamespace,
	}, db); err != nil {
		return nil, status.Errorf(codes.InvalidArgument,
			"Unable to update cluster "+"| Reason : "+err.Error())
	}
	cluster, err := json.Marshal(db.Spec)
	if err != nil {
		return nil, err
	}
	stat, err := json.Marshal(db.Status)
	if err != nil {
		return nil, err
	}
	return &pb.GetAerospikeClusterResponse{Cluster: cluster, Status: stat}, nil
}

func (a *AerospikeServer) GetClustersByNamespace(ctx context.Context, request *pb.GetAerospikeClustersByNamespaceRequest) (*pb.GetAerospikeClustersResponse, error) {
	db := &v1.AeroDatabaseList{}

	if err := a.Client.List(ctx, db, &client.ListOptions{Namespace: request.AerospikeNamespace}); err != nil {
		return nil, status.Errorf(codes.Internal,
			"Unable list clusters "+"| Reason : "+err.Error())
	}
	result, err := json.Marshal(db)
	if err != nil {
		return nil, err
	}
	return &pb.GetAerospikeClustersResponse{Clusters: result}, nil
}

func (a *AerospikeServer) GetAllClusters(ctx context.Context, request *pb.GetAllAerospikeClustersRequest) (*pb.GetAerospikeClustersResponse, error) {
	db := &v1.AeroDatabaseList{}

	if err := a.Client.List(ctx, db); err != nil {
		return nil, status.Errorf(codes.Internal,
			"Unable list clusters "+"| Reason : "+err.Error())
	}

	var names []v1.NamespacedName
	for _, v := range db.Items {
		//obj, _ := json.Marshal(*v.GetNamespacedName())
		names = append(names, v1.NamespacedName{Name: v.Name, Namespace: v.Namespace})
	}

	result, err := json.Marshal(names)
	if err != nil {
		return nil, err
	}

	return &pb.GetAerospikeClustersResponse{Clusters: result}, nil
}

func (a *AerospikeServer) UpdateCluster(ctx context.Context, request *pb.UpdateAerospikeClusterRequest) (*pb.UpdateAerospikeClusterResponse, error) {

	db := &v1.AeroDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      request.ClusterName,
			Namespace: request.ClusterNamespace,
		},
	}

	if err := a.Client.Get(context.Background(), client.ObjectKey{
		Name:      request.ClusterName,
		Namespace: metav1.NamespaceDefault,
	}, db); err != nil {
		return nil, status.Errorf(codes.InvalidArgument,
			"Unable to update cluster "+"| Reason : "+err.Error())
	}

	patchHelper, err := patch.NewHelper(db, a.Client)
	if err != nil {
		return nil, status.Errorf(codes.NotFound,
			"Unable to update cluster "+"| Reason : "+err.Error())
	}

	db.Spec.Options.Replicas = request.Options.Replicas
	fmt.Printf("[DEBUG] updating cluster %s \n", request.ClusterName)

	if err = patchHelper.Patch(context.Background(), db); err != nil {
		return nil, status.Errorf(codes.NotFound,
			"Unable to update cluster "+"|  Reason : "+err.Error())
	}

	return &pb.UpdateAerospikeClusterResponse{}, nil
}

func (k *AerospikeServer) DeleteCluster(ctx context.Context, request *pb.DeleteAerospikeClusterRequest) (*pb.DeleteAerospikeClusterResponse, error) {

	if request.AerospikeName == "" || request.AerospikeNamespace == "" {
		return nil, status.Errorf(codes.InvalidArgument,
			"Invalid input")
	}

	cluster := &v1.AeroDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      request.AerospikeName,
			Namespace: request.AerospikeNamespace,
		},
	}

	fmt.Printf("[DEBUG] deleting aerospike cluster %s \n", request.AerospikeName)

	if err := k.Client.Get(ctx, types.NamespacedName{Name: request.AerospikeName, Namespace: request.AerospikeNamespace}, cluster); err != nil {
		return nil, status.Errorf(codes.InvalidArgument,
			"unable to delete aerospike cluster "+request.AerospikeName+"| Reason : "+err.Error())
	}

	if err := k.Client.Delete(context.Background(), cluster); err != nil {
		return nil, status.Errorf(codes.Internal,
			"unable to delete aerospike cluster "+request.AerospikeName+"| Reason : "+err.Error())
	}

	return &pb.DeleteAerospikeClusterResponse{}, nil
}
