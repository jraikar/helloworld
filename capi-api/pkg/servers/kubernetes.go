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
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesServer struct {
	pb.AerostationKubernetesAPIServer
	Client client.Client
}

func NewKubernetesServer(client client.Client) *KubernetesServer {
	return &KubernetesServer{Client: client}
}

func (k *KubernetesServer) CreateCluster(ctx context.Context, request *pb.CreateKubernetesClusterRequest) (*pb.CreateKubernetesClusterResponse, error) {
	cluster := &v1.AeroClusterManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      request.Name,
			Namespace: request.Namespace,
		},
		Spec: v1.AeroClusterManagerSpec{
			Name:    request.Name,
			Suspend: false,
			ClusterOptions: v1.ClusterOptions{
				Name:          request.Name,
				Provider:      request.Provider,
				Replicas:      request.Replicas,
				KubeVersion:   request.KubeVersion,
				DockerOptions: &v1.DockerOptions{},
				EKSOptions: &v1.EKSOptions{
					Region:       request.Eks.Region,
					InstanceType: request.Eks.InstanceType,
					SSHKey:       request.Eks.SSHKey,
				},
			},
			ClusterID: v1.NamespacedName{
				Namespace: request.Namespace,
				Name:      request.Name,
			},
			Managed: true,
		},
	}

	fmt.Printf("[DEBUG] creating cluster %v \n", request)

	if err := k.Client.Create(context.TODO(), cluster); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to create cluster: %s", err.Error())
	}

	return &pb.CreateKubernetesClusterResponse{}, nil
}

func (k *KubernetesServer) GetCluster(ctx context.Context, request *pb.GetKubernetesClusterRequest) (*pb.GetKubernetesClusterResponse, error) {
	clusters := &v1.AeroClusterManagerList{}
	err := k.Client.List(context.Background(), clusters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get cluster "+request.Name+"| Reason : "+err.Error())
	}

	for _, v := range clusters.Items {
		if request.Name == v.Name {
			cluster, _ := json.Marshal(v.Spec)
			status, _ := json.Marshal(v.Status)
			return &pb.GetKubernetesClusterResponse{Cluster: cluster, Status: status}, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "cluster not found")
}

func (k *KubernetesServer) GetClusters(ctx context.Context, request *pb.GetKubernetesClustersRequest) (*pb.GetKubernetesClustersResponse, error) {
	clusters := &v1.AeroClusterManagerList{}
	err := k.Client.List(context.Background(), clusters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal service error")
	}

	var names []v1.NamespacedName
	for _, v := range clusters.Items {
		// obj, _ := json.Marshal(*v.GetNamespacedName())
		names = append(names, *v.GetNamespacedName())
	}

	res, _ := json.Marshal(names)
	return &pb.GetKubernetesClustersResponse{Clusters: res}, nil
}

func (k *KubernetesServer) UpdateCluster(ctx context.Context, request *pb.UpdateKubernetesClusterRequest) (*pb.UpdateKubernetesClusterResponse, error) {
	manager := &v1.AeroClusterManager{}
	awsInput := false

	// get AWS manager
	if request.Name != "" {
		awsInput = true
		if err := k.Client.Get(context.Background(), client.ObjectKey{
			Name:      request.Name,
			Namespace: metav1.NamespaceDefault,
		}, manager); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Unable to update cluster "+"| Reason : "+err.Error())
		}
	}

	patchHelper, err := patch.NewHelper(manager, k.Client)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Unable to update cluster "+"| Reason : "+err.Error())
	}

	if awsInput {
		manager.Spec.ClusterOptions.EKSOptions = &v1.EKSOptions{
			InstanceType: request.Eks.InstanceType,
			Region:       request.Eks.Region,
			SSHKey:       request.Eks.SSHKey,
		}
		manager.Spec.ClusterOptions.Replicas = request.Replicas
		fmt.Printf("[DEBUG] updating cluster %s \n", request.Name)
	}

	if err = patchHelper.Patch(context.Background(), manager); err != nil {
		return nil, status.Errorf(codes.NotFound, "Unable to update cluster "+"| Reason : "+err.Error())
	}

	return &pb.UpdateKubernetesClusterResponse{}, nil
}

func (k *KubernetesServer) DeleteCluster(ctx context.Context, request *pb.DeleteKubernetesClusterRequest) (*pb.DeleteKubernetesClusterResponse, error) {
	if request.Name == "" || request.Namespace == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to delete cluster ")
	}

	cluster := &v1.AeroClusterManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      request.Name,
			Namespace: request.Namespace,
		},
	}

	fmt.Printf("[DEBUG] deleting cluster %s \n", request.Name)

	if err := k.Client.Delete(context.Background(), cluster); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to delete cluster Test"+request.Name+"| Reason : "+err.Error())
	}

	return &pb.DeleteKubernetesClusterResponse{}, nil
}
