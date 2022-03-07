package cmd

import (
	"context"
	"errors"
	"log"

	v1 "github.com/aerospike/aerostation/api/v1"
	"github.com/aerospike/aerostation/pkg/utils"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func init() {
	deleteCluster.Flags().StringVar(&kubeOptions.kubeconfig, "kubeconfig", "~/.kube/config", "path to kubeconfig file to upload")
	deleteCluster.Flags().StringVar(&kubeOptions.kubeconfigContext, "context", "~/.kube/config", "kube context to use")
	deleteCluster.Flags().StringVar(&kubeOptions.kubeconfig, "namespace", "~/.kube/config", "kube namespace to use")

	deleteDatabase.Flags().StringVar(&kubeOptions.kubeconfig, "kubeconfig", "~/.kube/config", "path to kubeconfig file to upload")
	deleteDatabase.Flags().StringVar(&kubeOptions.kubeconfigContext, "context", "~/.kube/config", "kube context to use")
	deleteDatabase.Flags().StringVar(&kubeOptions.kubeconfig, "namespace", "~/.kube/config", "kube namespace to use")

	deleteCmd.AddCommand(deleteCluster)
	deleteCmd.AddCommand(deleteDatabase)
}

var deleteCluster = &cobra.Command{
	Use:   "cluster",
	Short: "delete a cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		return runDeleteCluster(args[0])
	},
}

var deleteDatabase = &cobra.Command{
	Use:   "database [cluster-name]",
	Short: "delete a cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		return runDeleteCluster(args[0])
	},
}

/*
func NewDeleteCmd(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	deleteFlags := cmddel.NewDeleteCommandFlags("containing the resource to delete.")

	cmd := &cobra.Command{
		Use:                   "delete ([CLUSTER|DATABASE] [NAME] |)",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Delete clusters and databases by name"),
		Args:                  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			o, err := deleteFlags.ToOptions(nil, streams)
			cmdutil.CheckErr(err)
			cmdutil.CheckErr(o.Complete(f, args, cmd))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.RunDelete(f))
		},
		SuggestFor: []string{"rm"},
	}

	deleteFlags.AddFlags(cmd)
	cmdutil.AddDryRunFlag(cmd)

	return cmd
}
*/

func runDeleteCluster(name string) error {

	if name == "" {
		return errors.New("[ERROR] invalid input")
	}

	kubeClient, err := utils.GetKubeClient(kubeOptions.kubeconfig, Scheme)
	if err != nil {
		return err
	}
	createOptions.Name = name
	log.Printf("[INFO] deleting cluster - %s \n", name)

	cluster := &v1.AeroClusterManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
	}

	ctx := context.Background()
	err = kubeClient.Get(ctx, types.NamespacedName{Name: name, Namespace: metav1.NamespaceDefault}, cluster)

	if err != nil {
		return err
	}

	return kubeClient.Delete(context.Background(), cluster)
}
