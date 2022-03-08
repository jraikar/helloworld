package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"

	v1 "github.com/aerospike/aerostation/api/v1"
	"github.com/aerospike/aerostation/pkg/utils"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type kubeConfigOptions struct {
	kubeconfig        string
	kubeconfigContext string
	//namespace         string
}

var (
	kubeOptions   = &kubeConfigOptions{}
	createOptions = &clusterOptions{}
	provider      string
	kubeVersion   string
)

func init() {
	newCluster.Flags().StringVar(&kubeOptions.kubeconfig, "kubeconfig", "~/.kube/config", "path to kubeconfig file to upload")
	newCluster.Flags().StringVar(&kubeOptions.kubeconfigContext, "context", "~/.kube/config", "kube context to use")
	newCluster.Flags().StringVar(&kubeOptions.kubeconfig, "namespace", "~/.kube/config", "kube namespace to use")

	newdb.Flags().StringVar(&kubeOptions.kubeconfig, "kubeconfig", "~/.kube/config", "path to kubeconfig file to upload")
	newdb.Flags().StringVar(&kubeOptions.kubeconfigContext, "context", "~/.kube/config", "kube context to use")
	newdb.Flags().StringVar(&kubeOptions.kubeconfig, "namespace", "~/.kube/config", "kube namespace to use")

	newCluster.Flags().StringVar(&provider, "provider", "eks", "docker,eks,aks,gke")
	newCluster.Flags().StringVar(&kubeVersion, "kubeversion", "v1.22.0", "k8s version")
	newCluster.Flags().IntVar(&createOptions.Replicas, "replicas", 3, "number of workers in the kubernetes cluster")
	newCluster.Flags().StringVar(&createOptions.InstanceType, "instance-type", "t3.large", "instance to use")
	newCluster.Flags().StringVar(&createOptions.Region, "region", "us-east-1", "the region")
	newCluster.Flags().StringVar(&createOptions.SSHKey, "ssh-key", "aerospike", "the aws ssh-key-file")

	createCmd.AddCommand(newCluster)
	createCmd.AddCommand(newdb)
}

var newCluster = &cobra.Command{
	Use:   "cluster [cluster-name]",
	Short: "create a cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		return runNewCluster(args[0])
	},
}

var newdb = &cobra.Command{
	Use:   "database [db-name]",
	Short: "create a new db in a kubernetes cluster",
	Args:  cobra.ExactArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		return runNewDatabase(args[0], args[1])
	},
}

func runNewCluster(name string) error {
	// todo: validate other fields like ssh-key, instance, region provided as input
	if name == "" {
		return errors.New("[ERROR] invalid input")
	}

	kubeClient, err := utils.GetKubeClient(kubeOptions.kubeconfig, Scheme)
	if err != nil {
		return err
	}
	createOptions.Name = name
	log.Printf("[INFO] aerostation creating workload cluster - %s - %s \n", name, provider)

	clusterOptions := v1.ClusterOptions{
		Name:        name,
		Provider:    provider,
		Replicas:    int32(createOptions.Replicas),
		KubeVersion: kubeVersion,
	}

	switch provider {
	case "docker":
		clusterOptions.DockerOptions = &v1.DockerOptions{}
	case "eks":
		clusterOptions.EKSOptions = &v1.EKSOptions{
			Region:       createOptions.Region,
			InstanceType: createOptions.InstanceType,
			SSHKey:       createOptions.SSHKey,
		}
	default:
		return fmt.Errorf("unhandled provider type")
	}

	cluster := &v1.AeroClusterManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1.AeroClusterManagerSpec{
			Name:           name,
			Suspend:        false,
			ClusterOptions: clusterOptions,
			ClusterID: v1.NamespacedName{
				Namespace: metav1.NamespaceDefault,
				Name:      name,
			},
			Managed: true,
		},
	}

	if err = kubeClient.Create(context.TODO(), cluster); err != nil {
		return err
	}

	log.Println("[INFO] workload cluster creating")

	return nil
}

func runNewDatabase(name, cluster string) error {
	c, err := utils.GetKubeClient(kubeOptions.kubeconfig, Scheme)
	if err != nil {
		return err
	}

	fmt.Println("got kubernetes client")

	db := &v1.AeroDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1.AeroDatabaseSpec{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
			Cluster: v1.ClusterKey{
				Name:      cluster,
				Namespace: metav1.NamespaceDefault, // TODO: pass in namespace
			},
			DatabaseType: "memory",
			Options: v1.DatabaseOptions{
				Replicas: 2,
			},
		},
	}

	return c.Create(context.Background(), db)
}

func GetClusterClientset(kubeOptions kubeConfigOptions) (*kubernetes.Clientset, error) {
	// uses the current context in kubeconfig
	// path-to-kubeconfig -- for example, /root/.kube/config
	config, err := clientcmd.BuildConfigFromFlags("", kubeOptions.kubeconfig)
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, _ := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	// access the API to list pods
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	_ = len(pods.Items)

	return clientset, nil
}
