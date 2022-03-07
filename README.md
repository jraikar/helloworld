# aerostation
a cloud management utility for aerospike databases

Getting started:

1) set up a cluster-api management cluster with support for AWS or Docker-> https://cluster-api.sigs.k8s.io/user/quick-start.html
1) make aeroctl -> builds cli
1) make install -> install crds
1) make run -> run aerostation
1) make swagger-generate -> generate swagger spec in json format
1) make swagger-serve -> serve swagger spec locally


Creating your first kubernetes cluster:
1) ./bin/aeroctl create cluster my-cluster
at this point the cluster should be created and the operator installed. This can take 10 to 20 minutes depending.

To create a cluster using the CAPI Docker provider, you specify the `--provider=docker` command line flag.

Creating your first DB:
1) ./bin/aeroctl create database my-db --ako my-deb-config-file


aeroctl supported commands

```bash
aeroctl create [database|cluster] [name] [-flags]
aeroctl status [database|cluster] [name] [-flags]
aeroctl delete [database|cluster] [name] [-flags]
aeroctl update [database|cluster] [name] [-flags]
```


## Development

A tilt file has been supplied to make local development quick an easy.

1) First you need to install a local kubernetes solution, because we use tilt for development it is best to use their install documentation. https://docs.tilt.dev/install.html
2) With tilt installed and a kubernetes cluster running, its time to install capi, we currently only support EKS and Docker clusters so use the `aws` or `docker` provider
3) Run `make install` in the base aerostation directory.
4) Run `tilt up` in the base aerostation directory.  Use the `Tiltfile.docker` file when using the CAPI Docker provider.
5) The api server will be exposed on `localhost:9000`

## Docker Provider Notes

When using the CAPI Docker provider, you will be required to set up the kind cluster to work with your dev host's Docker socket.

See https://cluster-api.sigs.k8s.io/user/quick-start.html for details on how 'extra mounts' is configured.

Also, to aide your local development, you can run a local Docker registry that is accessible to your kind cluster.  

Settup up a local registry that works with kind is explained in the following links:
https://kind.sigs.k8s.io/docs/user/local-registry/
https://github.com/tilt-dev/kind-local

### Docker Create Example

The following command shows a request to the REST API for creating a cluster using
the Docker provider:
```bash
curl -u aerospike:Aerospike123! -X POST http://<yourdevhost>:9000/api/v1/admin/kubernetes/clusters \
-H 'Content-Type: application/json' \
-d ' { "name": "mycluster", "dockerOptions": { "replicas": 3 }, "provider": "docker" }'
```
