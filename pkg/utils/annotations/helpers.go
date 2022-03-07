package annotations

import (
	v1 "github.com/aerospike/aerostation/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsPaused returns true if the Cluster is paused or the object has the `paused` annotation.
func IsSuspended(cluster *v1.AeroClusterManager, o metav1.Object) bool {
	if cluster.Spec.Suspend {
		return true
	}
	return hasAnnotation(o, v1.PausedAnnotation)
}

// hasAnnotation returns true if the object has the specified annotation.
func hasAnnotation(o metav1.Object, annotation string) bool {
	annotations := o.GetAnnotations()
	if annotations == nil {
		return false
	}
	_, ok := annotations[annotation]
	return ok
}
