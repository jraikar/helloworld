package kube

import (
	"context"

	"github.com/aerospike/aerostation/pkg/secrets"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func KubeApply(ctx context.Context, obj *unstructured.Unstructured, client *client.Client) {

}

type KubeHelper struct {
	Client *client.Client
}

func Patch(ctx context.Context, namespace, name, resource string, client rest.RESTClient, pt types.PatchType, isScope bool, data []byte) (runtime.Object, error) {
	return client.Patch(pt).
		NamespaceIfScoped(namespace, isScope).
		//(<resource>/[ns/<namespace>/]<name>)
		Resource(resource).
		Name(name).
		Body(data).
		Do(ctx).
		Get()
}

// FromSecret fetches the Kubeconfig for a Cluster.
func FromSecret(ctx context.Context, c client.Reader, cluster client.ObjectKey) ([]byte, error) {
	out, err := secrets.Get(ctx, c, cluster, secrets.Kubeconfig)
	if err != nil {
		return nil, err
	}
	return toKubeconfigBytes(out)
}

func toKubeconfigBytes(out *corev1.Secret) ([]byte, error) {
	data, ok := out.Data[secrets.KubeconfigDataName]
	if !ok {
		return nil, errors.Errorf("missing key %q in secret data", secrets.KubeconfigDataName)
	}
	return data, nil
}
