package cmd

import (
	"context"
	"fmt"

	"github.com/aerospike/aerostation/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	v1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

func init() {

	clusterstatusCmd.Flags().StringVar(&kubeOptions.kubeconfig, "kubeconfig", "~/.kube/config", "path to kubeconfig file to upload")
	clusterstatusCmd.Flags().StringVar(&kubeOptions.kubeconfigContext, "context", "~/.kube/config", "kube context to use")
	clusterstatusCmd.Flags().StringVar(&kubeOptions.kubeconfig, "namespace", "~/.kube/config", "kube namespace to use")

	statusCmd.AddCommand(clusterstatusCmd)
	statusCmd.AddCommand(dbstatusCmd)
	statusCmd.AddCommand(clusterstatusCmd)
}

var clusterstatusCmd = &cobra.Command{
	Use:   "cluster [cluster-name]",
	Short: "get the status of type cluster",
	Example: Examples(`
		# Get the status of the cluster named test-1.
		aeroctl status cluster test-1
		`),
	Args: cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		return runStatusCluster(args[0])
	},
}

var dbstatusCmd = &cobra.Command{
	Use:   "database [database-name]",
	Short: "get the status of type cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		return runStatusDB(args[0])
	},
}

func runStatusDB(name string) error {
	return nil
}

func runStatusCluster(name string) error {
	c, err := utils.GetKubeClient(kubeOptions.kubeconfig, Scheme)

	if err != nil {
		return err
	}
	var cluster v1beta1.Cluster
	key := types.NamespacedName{
		Namespace: "default",
		Name:      name,
	}
	err = c.Get(context.Background(), key, &cluster)
	if err != nil {
		return err
	}

	fmt.Print(cluster)

	return nil
}
