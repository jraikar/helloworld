package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	aerostationv1 "github.com/aerospike/aerostation/api/v1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	cliflag "k8s.io/component-base/cli/flag"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	capiawsv1beta1 "sigs.k8s.io/cluster-api-provider-aws/api/v1beta1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	kubeadmcontrolplanev1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

const Indentation = `  `

var (
	verbosity *int
)

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(deleteCmd)
}

var rootCmd = &cobra.Command{
	Use:   "aeroctl",
	Short: "cli used to interact with aerostation",
	Run: func(c *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if verbosity != nil && *verbosity >= 5 {
			if err, ok := err.(stackTracer); ok {
				for _, f := range err.StackTrace() {
					fmt.Fprintf(os.Stderr, "%+s:%d\n", f, f)
				}
			}
		}
		// TODO: print cmd help if validation error
		os.Exit(1)
	}
}

func NewAeroctlCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	flags := rootCmd.PersistentFlags()
	flags.SetNormalizeFunc(cliflag.WarnWordSepNormalizeFunc) // Warn for "_" flags

	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	//flags.BoolVar(&warningsAsErrors, "warnings-as-errors", warningsAsErrors, "Treat warnings received from the server as errors and exit with a non-zero exit code")

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	kubeConfigFlags.AddFlags(flags)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(rootCmd.PersistentFlags())
	// Updates hooks to add kubectl command headers: SIG CLI KEP 859.
	// TODO: I don't think we need this. investigate.
	// addCmdHeaderHooks(rootCmd, kubeConfigFlags)

	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// Sending in 'nil' for the getLanguageFn() results in using
	// the LANG environment variable.
	//
	// TODO: Consider adding a flag or file preference for setting
	// the language, instead of just loading from the LANG env. variable.
	i18n.LoadTranslations("kubectl", nil)

	// From this point and forward we get warnings on flags that contain "_" separators
	rootCmd.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

	//f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	//ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: err}

	//rootCmd.AddCommand(NewDeleteCmd(f, ioStreams))
	return rootCmd
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "get the status of type [cluster|db]",
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update the status of type [cluster|db]",
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create different database resources [cluster|db]",
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete resources [cluster|db]",
}

// Examples normalizes a command's examples to follow the conventions.
func Examples(s string) string {
	if len(s) == 0 {
		return s
	}
	return normalizer{s}.trim().indent().string
}

type normalizer struct {
	string
}

func (s normalizer) heredoc() normalizer {
	s.string = heredoc.Doc(s.string)
	return s
}

func (s normalizer) trim() normalizer {
	s.string = strings.TrimSpace(s.string)
	return s
}

func (s normalizer) indent() normalizer {
	splitLines := strings.Split(s.string, "\n")
	indentedLines := make([]string, 0, len(splitLines))
	for _, line := range splitLines {
		trimmed := strings.TrimSpace(line)
		indented := Indentation + trimmed
		indentedLines = append(indentedLines, indented)
	}
	s.string = strings.Join(indentedLines, "\n")
	return s
}

var (
	// Scheme contains a set of API resources used by clusterctl.
	Scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
	utilruntime.Must(aerostationv1.AddToScheme(Scheme))
	utilruntime.Must(capiv1beta1.AddToScheme(Scheme))
	utilruntime.Must(capiawsv1beta1.AddToScheme(Scheme))
	utilruntime.Must(kubeadmv1beta1.AddToScheme(Scheme))
	utilruntime.Must(kubeadmcontrolplanev1beta1.AddToScheme(Scheme))
	utilruntime.Must(rbac.AddToScheme(Scheme))
	//+kubebuilder:scaffold:scheme
}
