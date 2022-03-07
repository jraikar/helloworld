package utils

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

	v1 "github.com/aerospike/aerostation/api/v1"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/rancher/wrangler/pkg/kubeconfig"
	"sigs.k8s.io/cluster-api/util/patch"

	"github.com/aerospike/aerostation/pkg/remote"
	"github.com/aerospike/aerostation/pkg/utils/ako"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rntyaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

func Patch(ctx context.Context, objToPatch, empty client.Object, client client.Client) error {
	err := client.Get(ctx, types.NamespacedName{Name: objToPatch.GetName(), Namespace: objToPatch.GetNamespace()}, empty)
	// Do we need to create this Object?
	if err != nil {

		if apierrors.IsNotFound(err) {
			log.Println("Did not find, creating")
			return client.Create(ctx, objToPatch)
		}
		return err
	}
	clusterPatchHelper, err := patch.NewHelper(empty, client)
	if err != nil {
		return err
	}

	if err := clusterPatchHelper.Patch(ctx, objToPatch, patch.WithForceOverwriteConditions{}, patch.WithStatusObservedGeneration{}); err != nil {
		return err
	}
	return nil
}

func GetKubeClient(kconfig string, scheme *runtime.Scheme) (client.Client, error) {
	cfg, err := getRestClient(kconfig)
	if err != nil {
		return nil, err
	}
	return client.New(cfg, client.Options{Scheme: scheme})
}

func getRestClient(kconfig string) (*rest.Config, error) {
	return kubeconfig.GetNonInteractiveClientConfig(kconfig).ClientConfig()
}

// Helper functions to check and remove string from a slice of strings.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// DecodeYAML unmarshals a YAML document or multidoc YAML as unstructured
// objects, placing each decoded object into a channel.
func DecodeYAML(data []byte) (<-chan *unstructured.Unstructured, <-chan error) {
	var (
		chanErr        = make(chan error)
		chanObj        = make(chan *unstructured.Unstructured)
		multidocReader = utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))
	)

	go func() {
		defer close(chanErr)
		defer close(chanObj)

		// Iterate over the data until Read returns io.EOF. Every successful
		// read returns a complete YAML document.
		for {
			buf, err := multidocReader.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				chanErr <- errors.Wrap(err, "failed to read yaml data")
				return
			}

			// Do not use this YAML doc if it is unkind.
			var typeMeta runtime.TypeMeta
			if err := yaml.Unmarshal(buf, &typeMeta); err != nil {
				continue
			}
			if typeMeta.Kind == "" {
				continue
			}

			// Define the unstructured object into which the YAML document will be
			// unmarshaled.
			obj := &unstructured.Unstructured{
				Object: map[string]interface{}{},
			}

			dec := rntyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			_, _, err = dec.Decode(buf, nil, obj)

			if err != nil {
				chanErr <- errors.Wrap(err, "failed to unmarshal yaml data")
				return
			}
			/*
				// Unmarshal the YAML document into the unstructured object.
				if err := yaml.Unmarshal(buf, &obj.Object); err != nil {
					chanErr <- errors.Wrap(err, "failed to unmarshal yaml data")
					return
				}
			*/
			// Place the unstructured object into the channel.
			chanObj <- obj
		}
	}()

	return chanObj, chanErr
}

// ForEachObjectInYAMLActionFunc is a function that is executed against each
// object found in a YAML document.
// When a non-empty namespace is provided then the object is assigned the
// namespace prior to any other actions being performed with or to the object.
type ForEachObjectInYAMLActionFunc func(ctx context.Context, unstruct *unstructured.Unstructured, cfg *rest.Config) error

// ForEachObjectInYAML excutes actionFn for each object in the provided YAML.
// If an error is returned then no further objects are processed.
// The data may be a single YAML document or multidoc YAML.
func ForEachObjectInYAML(ctx context.Context, restConfig *rest.Config, data []byte, namespace string, actionFn ForEachObjectInYAMLActionFunc) error {
	chanObj, chanErr := DecodeYAML(data)
	for {
		select {
		case obj := <-chanObj:
			if obj == nil {
				return nil
			}
			if namespace != "" {
				// not sure about this could be requires or not.
				obj.SetNamespace(namespace)
			}

			log.Println("Name of Object: ", obj.GetName())
			if err := actionFn(ctx, obj, restConfig); err != nil {
				return err
			}
		case err := <-chanErr:
			if err == nil {
				return nil
			}
			return errors.Wrap(err, "received error while decoding yaml")
		}
	}
}

func Apply(ctx context.Context, obj client.Object, c client.Client, cluster client.ObjectKey) error {
	restcfg, err := remote.RESTConfig(ctx, c, cluster)
	if err != nil {
		return err
	}
	return ApplyObject(ctx, obj, restcfg)
}

func ApplyObject(ctx context.Context, obj client.Object, restcfg *rest.Config) error {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}

	return ApplyResource(ctx, &unstructured.Unstructured{Object: unstructuredObj}, restcfg)
}

// ApplyResource should probably patch using a dynamic client.
func ApplyResource(ctx context.Context, unstruct *unstructured.Unstructured, cfg *rest.Config) error {
	log.Println("===Applying===")

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	mapping, err := mapper.RESTMapping(unstruct.GroupVersionKind().GroupKind(), unstruct.GroupVersionKind().Version)
	if err != nil {
		return err
	}

	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(unstruct.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	data, err := json.Marshal(unstruct)
	if err != nil {
		return err
	}

	_, err = dr.Patch(ctx, unstruct.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		Force:        func() *bool { b := true; return &b }(),
		FieldManager: "aerostation-controller",
	})

	return err
}

// ApplyOperatorator takes a client to the management cluster!
func ApplyOperatorator(ctx context.Context, c client.Client, cluster client.ObjectKey) error {
	restcfg, err := remote.RESTConfig(ctx, c, cluster)
	if err != nil {
		return err
	}
	return ForEachObjectInYAML(ctx, restcfg, []byte(ako.AkoDeployFile), "", ApplyResource)
}

func ApplyDatabase(ctx context.Context, c client.Client, db *v1.AeroDatabase, customResourceBytes []byte) error {

	restcfg, err := remote.RESTConfig(ctx, c, db.Spec.Cluster.ToObjectKey())
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(restcfg)
	if err != nil {
		return err
	}

	//check for target namespace, create if necessary
	err = createNamespace(ctx, clientset, db.Spec.TargetNamespace)
	if err != nil {
		return err
	}

	//check for serviceaccount, create if necessary
	err = createServiceAccount(ctx, clientset, db.Spec.TargetNamespace)
	if err != nil {
		return err
	}

	//update the ako clusterrolebindings to include the targetNamespace
	err = updateClusterRoleBindings(ctx, clientset, db.Spec.TargetNamespace)
	if err != nil {
		return err
	}

	// create the aerospike feature secret, required for the database
	wd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return err
	}

	var secretBytes []byte
	secretBytes, err = ioutil.ReadFile(path.Join(wd, "/manifests/feature-secret.yaml"))
	if err != nil {
		log.Println(err)
		return err
	}

	err = ForEachObjectInYAML(ctx, restcfg, secretBytes, db.Spec.TargetNamespace, ApplyResource)
	if err != nil {
		log.Println(err)
		return err
	}

	// create the ako CR in the targetNamespace
	err = ForEachObjectInYAML(ctx, restcfg, customResourceBytes, db.Spec.TargetNamespace, ApplyResource)
	if err != nil {
		log.Println(err)
		return err
	}

	// apply the aerospike database rest client in the targetNamespace
	if db.Spec.DeployClient {
		restClientBytes := ako.GetAKORESTClient(db.Spec.TargetNamespace, db.Spec.Name)
		err = ForEachObjectInYAML(ctx, restcfg, restClientBytes, db.Spec.TargetNamespace, ApplyResource)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

// ApplyCertManager takes a client to the management cluster!
func ApplyCertManager(ctx context.Context, c client.Client, cluster client.ObjectKey) error {
	wd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	dat, err := ioutil.ReadFile(path.Join(wd, "/manifests/cert-manager.yaml"))
	if err != nil {
		log.Println("Failed to read cert-manager file!!!!")
		return err
	}

	restcfg, err := remote.RESTConfig(ctx, c, cluster)
	if err != nil {
		return err
	}

	return ForEachObjectInYAML(ctx, restcfg, dat, "", ApplyResource)
}

func createNamespace(ctx context.Context, clientset *kubernetes.Clientset, targetNamespace string) error {
	//check for target namespace, create if necessary
	labels := map[string]string{
		"istio-injection": "enabled",
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   targetNamespace,
			Labels: labels,
		},
	}
	_, err := clientset.CoreV1().Namespaces().Get(ctx, targetNamespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			_, err := clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func createServiceAccount(ctx context.Context, clientset *kubernetes.Clientset, targetNamespace string) error {

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "aerospike-operator-controller-manager",
		},
	}

	_, err := clientset.CoreV1().ServiceAccounts(targetNamespace).Get(ctx, sa.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			_, err := clientset.CoreV1().ServiceAccounts(targetNamespace).Create(ctx, sa, metav1.CreateOptions{})
			if err != nil {
				log.Printf("error when creating serviceaccount %s\n", err.Error())
				return err
			}
		} else {
			log.Printf("error when getting serviceaccount %s\n", err.Error())
			return err
		}
	}
	return nil
}

func updateClusterRoleBindings(ctx context.Context, clientset *kubernetes.Clientset, targetNamespace string) error {
	bindings := []string{"aerospike-operator-aerospikecluster-editor-rolebinding",
		"aerospike-operator-aerospikecluster-viewer-rolebinding",
		"aerospike-operator-manager-rolebinding",
		"aerospike-operator-proxy-rolebinding",
	}
	subject := rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      "aerospike-operator-controller-manager",
		Namespace: targetNamespace,
	}
	for i := 0; i < len(bindings); i++ {
		b, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, bindings[i], metav1.GetOptions{})
		if err != nil {
			log.Printf("error when getting clusterrolebinding %s\n", err.Error())
			return err
		}

		needsSubject := true
		for i := 0; i < len(b.Subjects); i++ {
			s := b.Subjects[i]
			if s.Kind == subject.Kind &&
				s.Name == subject.Name &&
				s.Namespace == subject.Namespace {
				needsSubject = false
			}
		}

		if needsSubject {
			b.Subjects = append(b.Subjects, subject)
			_, err = clientset.RbacV1().ClusterRoleBindings().Update(ctx, b, metav1.UpdateOptions{})
			if err != nil {
				log.Printf("error when updating clusterrolebinding %s\n", err.Error())
				return err
			}
		}

	}
	return nil

}
