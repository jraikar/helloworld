package ako

import (
	"bytes"
	"fmt"
	"text/template"

	v1 "github.com/aerospike/aerostation/api/v1"
)

func GetMemoryDB(options v1.AeroDatabaseSpec) []byte {
	type customization struct {
		Name      string
		Namespace string
		Replicas  int32
	}

	c := customization{
		Name:      options.Name,
		Namespace: options.Namespace,
		Replicas:  options.Options.Replicas,
	}

	t := template.Must(template.New("db").Parse(Nostorage))

	var b bytes.Buffer
	err := t.Execute(&b, c)
	if err != nil {
		fmt.Printf("error in templating %s\n", err.Error())
	}

	fmt.Printf("AerospikeCluster is going to be:  %s\n", b.String())

	return b.Bytes()
}

var Nostorage = `
apiVersion: asdb.aerospike.com/v1beta1
kind: AerospikeCluster
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  size: {{ .Replicas }}
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
