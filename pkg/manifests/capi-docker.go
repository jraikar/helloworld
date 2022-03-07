package manifests

var DockerManifests = `apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: capi-docker
  namespace: default
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - 192.168.0.0/16
    serviceDomain: cluster.local
    services:
      cidrBlocks:
      - 10.128.0.0/12
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1beta1
    kind: KubeadmControlPlane
    name: capi-docker-control-plane
    namespace: default
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
    kind: DockerCluster
    name: capi-docker
    namespace: default
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: DockerCluster
metadata:
  name: capi-docker
  namespace: default
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: DockerMachineTemplate
metadata:
  name: capi-docker-control-plane
  namespace: default
spec:
  template:
    spec:
      extraMounts:
      - containerPath: /var/run/docker.sock
        hostPath: /var/run/docker.sock
---
apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: capi-docker-control-plane
  namespace: default
spec:
  infrastructureTemplate:
    apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
    kind: DockerMachineTemplate
    name: capi-docker-control-plane
    namespace: default
  kubeadmConfigSpec:
    clusterConfiguration:
      apiServer:
        certSANs:
        - localhost
        - 127.0.0.1
      controllerManager:
        extraArgs:
          enable-hostpath-provisioner: "true"
    initConfiguration:
      nodeRegistration:
        criSocket: /var/run/containerd/containerd.sock
        kubeletExtraArgs:
          eviction-hard: nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%
    joinConfiguration:
      nodeRegistration:
        criSocket: /var/run/containerd/containerd.sock
        kubeletExtraArgs:
          eviction-hard: nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%
  replicas: 3
  version: v1.18.16
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: DockerMachineTemplate
metadata:
  name: capi-docker-md-0
  namespace: default
spec:
  template:
    spec: {}
---
apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
kind: KubeadmConfigTemplate
metadata:
  name: capi-docker-md-0
  namespace: default
spec:
  template:
    spec:
      joinConfiguration:
        nodeRegistration:
          kubeletExtraArgs:
            eviction-hard: nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  name: capi-docker-md-0
  namespace: default
spec:
  clusterName: capi-docker
  replicas: 3
  selector:
    matchLabels: null
  template:
    spec:
      bootstrap:
        configRef:
          apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
          kind: KubeadmConfigTemplate
          name: capi-docker-md-0
          namespace: default
      clusterName: capi-docker
      infrastructureRef:
        apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
        kind: DockerMachineTemplate
        name: capi-docker-md-0
        namespace: default
      version: v1.18.16
`
