package ako

import (
	"bytes"
	"fmt"
	"text/template"
)

func GetAKORESTClient(namespace, hostname string) []byte {
	type customization struct {
		Namespace string
		Hostname  string
	}

	c := customization{
		Namespace: namespace,
		Hostname:  hostname,
	}

	t := template.Must(template.New("akorestclient").Parse(AKORESTClient))

	var b bytes.Buffer
	err := t.Execute(&b, c)
	if err != nil {
		fmt.Printf("error in templating %s\n", err.Error())
	}

	fmt.Printf("AKO REST Client is going to be:  %s\n", b.String())

	return b.Bytes()
}

var AKORESTClient = `
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/instance: rest-client
    app.kubernetes.io/name: aerospike-rest-client
    app.kubernetes.io/version: 1.6.0
  name: rest-client-aerospike-rest-client
  namespace: {{ .Namespace }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/instance: rest-client
    app.kubernetes.io/name: aerospike-rest-client
    app.kubernetes.io/version: 1.6.0
  name: rest-client-aerospike-rest-client
  namespace: {{ .Namespace }}
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/instance: rest-client
      app.kubernetes.io/name: aerospike-rest-client
  template:
    metadata:
      labels:
        app.kubernetes.io/instance: rest-client
        app.kubernetes.io/name: aerospike-rest-client
    spec:
      containers:
      - env:
        - name: AEROSPIKE_RESTCLIENT_HOSTNAME
          value: {{ .Hostname }}
        - name: AEROSPIKE_RESTCLIENT_PORT
          value: "3000"
        - name: AEROSPIKE_RESTCLIENT_CLIENTPOLICY_USER
        - name: AEROSPIKE_RESTCLIENT_CLIENTPOLICY_PASSWORD
        - name: AEROSPIKE_RESTCLIENT_CLIENTPOLICY_CLUSTERNAME
        - name: AEROSPIKE_RESTCLIENT_REQUIREAUTHENTICATION
          value: "false"
        - name: AEROSPIKE_RESTCLIENT_POOL_SIZE
          value: "16"
        image: aerospike/aerospike-client-rest:latest
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /
            port: http
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        name: aerospike-rest-client
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        readinessProbe:
          failureThreshold: 3
          httpGet:
            path: /
            port: http
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
      serviceAccount: rest-client-aerospike-rest-client
      serviceAccountName: rest-client-aerospike-rest-client
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/instance: rest-client
    app.kubernetes.io/name: aerospike-rest-client
    app.kubernetes.io/version: 1.6.0
    release: rest-client
  name: rest-client-aerospike-rest-client
  namespace: {{ .Namespace }}
spec:
  ports:
  - name: http
    port: 8080
    protocol: TCP
    targetPort: http
  selector:
    app.kubernetes.io/instance: rest-client
    app.kubernetes.io/name: aerospike-rest-client
  sessionAffinity: None
  type: ClusterIP
`
