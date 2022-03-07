package manifests

import (
	"github.com/aerospike/aerospike-kubernetes-operator/api/v1beta1"
	v1 "github.com/aerospike/aerostation/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func StrToDBSpec(database string) *v1beta1.AerospikeCluster {
	obj := &v1beta1.AerospikeCluster{}
	sch := runtime.NewScheme()
	_ = v1.AddToScheme(sch)
	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode
	_, _, _ = decode([]byte(database), nil, obj)

	obj.Status.Pods = make(map[string]v1beta1.AerospikePodStatus)
	obj.Spec.AerospikeConfig = &v1beta1.AerospikeConfigSpec{Value: map[string]interface{}{}}

	return obj
}

func GetDBNostorageStorageStruct() *v1beta1.AerospikeCluster {
	obj := &v1beta1.AerospikeCluster{}
	sch := runtime.NewScheme()
	_ = v1.AddToScheme(sch)
	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode
	_, _, _ = decode([]byte(Nostorage), nil, obj)

	obj.Status.Pods = make(map[string]v1beta1.AerospikePodStatus)
	obj.Spec.AerospikeConfig = &v1beta1.AerospikeConfigSpec{Value: map[string]interface{}{}}

	return obj
}

var Nostorage = `
apiVersion: asdb.aerospike.com/v1beta1
kind: AerospikeCluster
metadata:
  name: aerocluster
  namespace: aerospike
spec:
  size: 2
  image: aerospike/aerospike-server-enterprise:5.6.0.7
  podSpec:
    multiPodPerHost: true
  validationPolicy:
    skipWorkDirValidate: true
    skipXdrDlogFileValidate: true
  storage:
    volumes:
      - name: aerospike-config-secret
        source:
          secret:
            secretName: aerospike-secret
        aerospike:
          path: /etc/aerospike/secret
  aerospikeConfig:
    service:
      feature-key-file: /etc/aerospike/secret/features.conf
    security:
      enable-security: false
    namespaces:
      - name: test
        memory-size: 3000000000
        replication-factor: 2
        storage-engine:
          type: memory
`
