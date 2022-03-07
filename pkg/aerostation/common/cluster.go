package common

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	// ClusterLabelName is the label set on machines linked to a cluster and
	// external objects(bootstrap and infrastructure providers).
	ClusterLabelName = "aerospike.com/cluster-name"

	ClusterSecretType corev1.SecretType = "aerospike.com/secret"
)
