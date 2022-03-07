package remote

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	gruntime "runtime"

	"github.com/aerospike/aerostation/pkg/kube"
	"github.com/pkg/errors"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/version"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultClientTimeout = 10 * time.Second
	unknowString         = "unknown"
)

// RESTConfig returns a configuration instance to be used with a Kubernetes client.
func RESTConfig(ctx context.Context, c client.Reader, cluster client.ObjectKey) (*restclient.Config, error) {
	kubeConfig, err := kube.FromSecret(ctx, c, cluster)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve kubeconfig secret for Cluster %s/%s", cluster.Namespace, cluster.Name)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create REST configuration for Cluster %s/%s", cluster.Namespace, cluster.Name)
	}

	restConfig.UserAgent = DefaultUserAgent("cluster-cache-tracker")
	restConfig.Timeout = defaultClientTimeout

	return restConfig, nil
}

func buildUserAgent(command, version, sourceName, os, arch, commit string) string {
	return fmt.Sprintf(
		"%s/%s %s (%s/%s) cluster.x-k8s.io/%s", command, version, sourceName, os, arch, commit)
}

// DefaultClusterAPIUserAgent returns a User-Agent string built from static global vars.
func DefaultUserAgent(sourceName string) string {
	return buildUserAgent(
		adjustCommand(os.Args[0]),
		adjustVersion(version.Get().GitVersion),
		adjustSourceName(sourceName),
		gruntime.GOOS,
		gruntime.GOARCH,
		adjustCommit(version.Get().GitCommit))
}

// adjustSourceName returns the name of the source calling the client.
func adjustSourceName(c string) string {
	if len(c) == 0 {
		return unknowString
	}
	return c
}

// adjustCommit returns sufficient significant figures of the commit's git hash.
func adjustCommit(c string) string {
	if len(c) == 0 {
		return unknowString
	}
	if len(c) > 7 {
		return c[:7]
	}
	return c
}

// adjustVersion strips "alpha", "beta", etc. from version in form
// major.minor.patch-[alpha|beta|etc].
func adjustVersion(v string) string {
	if len(v) == 0 {
		return unknowString
	}
	seg := strings.SplitN(v, "-", 2)
	return seg[0]
}

// adjustCommand returns the last component of the
// OS-specific command path for use in User-Agent.
func adjustCommand(p string) string {
	// Unlikely, but better than returning "".
	if len(p) == 0 {
		return unknowString
	}
	return filepath.Base(p)
}
