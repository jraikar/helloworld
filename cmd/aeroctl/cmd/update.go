package cmd

import (
	"context"

	v1 "github.com/aerospike/aerostation/api/v1"
	"github.com/aerospike/aerostation/pkg/utils"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clusterOptions struct {
	// Provider string
	Name         string
	Replicas     int
	InstanceType string
	Region       string
	SSHKey       string

	// ClusterId string - unique id assigned to workload cluster (todo?)
	// Owner string 	//workload cluster owner (todo?)
	// AvailabilityZones    []string // todo ?
}

var updateOptions = &clusterOptions{}

func init() {
	updateCluster.Flags().StringVar(&kubeOptions.kubeconfig, "kubeconfig", "~/.kube/config", "path to kubeconfig file to upload")
	updateCluster.Flags().StringVar(&kubeOptions.kubeconfigContext, "context", "", "kube context to use")
	updateCluster.Flags().StringVar(&kubeOptions.kubeconfig, "namespace", "default", "kube namespace to use")

	// updateCluster.Flags().StringVar(&updateOptions.Provider, "provider", "aws", "Installation infrastructure provider") todo:
	updateCluster.Flags().IntVar(&updateOptions.Replicas, "replicas", 3, "number of workers in the kubernetes cluster")
	updateCluster.Flags().StringVar(&updateOptions.InstanceType, "instance-type", "t3.large", "instance to use")
	updateCluster.Flags().StringVar(&updateOptions.Region, "region", "us-east-1", "the region")
	updateCluster.Flags().StringVar(&updateOptions.SSHKey, "ssh-key", "aerospike.pem", "the aws ssh-key-file")

	updateCmd.AddCommand(updateCluster)
}

var updateCluster = &cobra.Command{
	Use:   "cluster [cluster-name]",
	Short: "update a cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		return runUpdateCluster(args[0])
	},
}

func runUpdateCluster(name string) error {
	kubeClient, err := utils.GetKubeClient(kubeOptions.kubeconfig, Scheme)
	if err != nil {
		return err
	}

	manager := &v1.AeroClusterManager{}
	if err = kubeClient.Get(context.Background(), client.ObjectKey{
		Name:      name,
		Namespace: metav1.NamespaceDefault,
	}, manager); err != nil {
		return err
	}

	patchHelper, err := patch.NewHelper(manager, kubeClient)
	if err != nil {
		return err
	}

	manager.Spec.ClusterOptions.EKSOptions = &v1.EKSOptions{
		InstanceType: updateOptions.InstanceType,
		Region:       updateOptions.Region,
		SSHKey:       updateOptions.SSHKey,
	}
	manager.Spec.ClusterOptions.Replicas = int32(updateOptions.Replicas)

	return patchHelper.Patch(context.Background(), manager)
}
