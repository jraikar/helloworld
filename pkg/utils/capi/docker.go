package capi

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path"

	aeroremote "github.com/aerospike/aerostation/pkg/remote"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/aerospike/aerostation/api/v1"
	"github.com/aerospike/aerostation/pkg/utils"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	infrav1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
)

func ApplyDocker(kubeClient client.Client, manager *v1.AeroClusterManagerSpec, config *rest.Config) error {
	cluster := getCapiClusterDocker(&manager.ClusterOptions)

	// debug(cluster)

	log.Println("[INFO] creating Cluster")

	if err := utils.ApplyObject(context.Background(), cluster, config); err != nil {
		log.Print("Failed to create cluster: ", err)
		return err
	}

	dockerCluster := getDockerCluster(manager)
	// debug(dockerCluster)

	log.Println("[INFO] creating DockerCluster")

	if err := utils.ApplyObject(context.Background(), dockerCluster, config); err != nil {
		log.Print("Failed to create DockerCluster: ", err)
		return err
	}

	cpdockerMachineTemplate := getDockerMachineTemplate(manager, "-control-plane")
	// debug(cpdockerMachineTemplate)

	log.Printf("[INFO] creating DockerMachineTemplate %s\n", cpdockerMachineTemplate.ObjectMeta.Name)

	if err := utils.ApplyObject(context.Background(), cpdockerMachineTemplate, config); err != nil {
		log.Print("Failed to create DockerMachineTemplate: ", err)
		return err
	}

	kubeadmControlPlane := getKubeadmControlPlane(manager)
	// debug(kubeadmControlPlane)

	log.Println("[INFO] creating KubeadmControlPlane")

	if err := utils.ApplyObject(context.Background(), kubeadmControlPlane, config); err != nil {
		log.Print("Failed to create KubeadmControlPlane: ", err)
		return err
	}

	mddockerMachineTemplate := getDockerMachineTemplate(manager, "-md-0")
	// debug(mddockerMachineTemplate)

	log.Printf("[INFO] creating DockerMachineTemplate %s\n", mddockerMachineTemplate.ObjectMeta.Name)

	if err := utils.ApplyObject(context.Background(), mddockerMachineTemplate, config); err != nil {
		log.Print("Failed to create DockerMachineTemplate md: ", err)
		return err
	}

	kubeadmConfigTemplate := getKubeadmConfigTemplate(manager)
	// debug(kubeadmConfigTemplate)

	log.Println("[INFO] creating KubeadmConfigTemplate")

	if err := utils.ApplyObject(context.Background(), kubeadmConfigTemplate, config); err != nil {
		log.Print("Failed to create KubeadmConfigTemplate: ", err)
		return err
	}

	machineDeployment := getMachineDeploymentDocker(manager)
	// debug(machineDeployment)

	log.Println("[INFO] creating MachineDeployment")

	if err := utils.ApplyObject(context.Background(), machineDeployment, config); err != nil {
		log.Print("Failed to create MachineDeployment: ", err)
		return err
	}

	return nil
}

func getDockerCluster(manager *v1.AeroClusterManagerSpec) *infrav1.DockerCluster {
	return &infrav1.DockerCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DockerCluster",
			APIVersion: infrav1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      manager.ClusterOptions.Name,
			Namespace: manager.ClusterID.Namespace,
		},
	}
}

func getDockerMachineTemplate(manager *v1.AeroClusterManagerSpec, suffix string) (t *infrav1.DockerMachineTemplate) {
	t = &infrav1.DockerMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DockerMachineTemplate",
			APIVersion: infrav1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      manager.ClusterOptions.Name + suffix,
			Namespace: manager.ClusterID.Namespace,
		},
		Spec: infrav1.DockerMachineTemplateSpec{
			Template: infrav1.DockerMachineTemplateResource{
				Spec: infrav1.DockerMachineSpec{
					ExtraMounts: []infrav1.Mount{
						{
							ContainerPath: "/var/run/docker.sock",
							HostPath:      "/var/run/docker.sock",
						},
					},
				},
			},
		},
	}
	if suffix == "-md-0" {
		t.Spec = infrav1.DockerMachineTemplateSpec{}
	}
	return t
}

func getKubeadmControlPlane(manager *v1.AeroClusterManagerSpec) *controlplanev1.KubeadmControlPlane {
	extraArgs := make(map[string]string)
	extraArgs["enable-hostpath-provisioner"] = "true"
	kubeletExtraArgs := make(map[string]string)
	kubeletExtraArgs["cgroup-driver"] = "cgroupfs"
	kubeletExtraArgs["eviction-hard"] = "nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%"
	return &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: controlplanev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getControlplaneName(&manager.ClusterOptions),
			Namespace: manager.ClusterID.Namespace,
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			Version:  manager.ClusterOptions.KubeVersion,
			Replicas: &manager.ClusterOptions.Replicas,
			KubeadmConfigSpec: bootstrapv1.KubeadmConfigSpec{
				ClusterConfiguration: &bootstrapv1.ClusterConfiguration{
					APIServer: bootstrapv1.APIServer{
						CertSANs: []string{
							"localhost",
							"127.0.0.1",
						},
					},
					ControllerManager: bootstrapv1.ControlPlaneComponent{
						ExtraArgs: extraArgs,
					},
				},
				InitConfiguration: &bootstrapv1.InitConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						CRISocket:        "/var/run/containerd/containerd.sock",
						KubeletExtraArgs: kubeletExtraArgs,
					},
				},
				JoinConfiguration: &bootstrapv1.JoinConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						CRISocket:        "/var/run/containerd/containerd.sock",
						KubeletExtraArgs: kubeletExtraArgs,
					},
				},
			},
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: corev1.ObjectReference{
					APIVersion: infrav1.GroupVersion.String(),
					Kind:       "DockerMachineTemplate",
					Name:       getControlplaneName(&manager.ClusterOptions),
					Namespace:  manager.ClusterID.Namespace,
				},
			},
		},
	}
}

func getKubeadmConfigTemplate(manager *v1.AeroClusterManagerSpec) *bootstrapv1.KubeadmConfigTemplate {
	kubeletExtraArgs := make(map[string]string)
	kubeletExtraArgs["cgroup-driver"] = "cgroupfs"
	kubeletExtraArgs["eviction-hard"] = "nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%"
	return &bootstrapv1.KubeadmConfigTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmConfigTemplate",
			APIVersion: bootstrapv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      manager.ClusterOptions.Name + "-md-0",
			Namespace: manager.ClusterID.Namespace,
		},
		Spec: bootstrapv1.KubeadmConfigTemplateSpec{
			Template: bootstrapv1.KubeadmConfigTemplateResource{
				Spec: bootstrapv1.KubeadmConfigSpec{
					JoinConfiguration: &bootstrapv1.JoinConfiguration{
						NodeRegistration: bootstrapv1.NodeRegistrationOptions{
							KubeletExtraArgs: kubeletExtraArgs,
						},
					},
				},
			},
		},
	}
}

func getMachineDeploymentDocker(manager *v1.AeroClusterManagerSpec) *capiv1beta1.MachineDeployment {
	return &capiv1beta1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       machineDeploymentKind,
			APIVersion: capiv1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        manager.ClusterOptions.Name + "-md-0",
			Namespace:   manager.ClusterID.Namespace,
			ClusterName: getDefaultName(&manager.ClusterOptions),
		},
		Spec: capiv1beta1.MachineDeploymentSpec{
			ClusterName: getDefaultName(&manager.ClusterOptions),
			Replicas:    pointer.Int32Ptr(manager.ClusterOptions.Replicas),
			Template: capiv1beta1.MachineTemplateSpec{
				Spec: capiv1beta1.MachineSpec{
					Bootstrap: capiv1beta1.Bootstrap{
						ConfigRef: &corev1.ObjectReference{
							Kind:       "KubeadmConfigTemplate",
							Name:       getMachineDeploymentName(&manager.ClusterOptions),
							Namespace:  manager.ClusterID.Namespace,
							APIVersion: bootstrapv1.GroupVersion.String(),
						},
					},
					ClusterName: getDefaultName(&manager.ClusterOptions),
					InfrastructureRef: corev1.ObjectReference{
						Kind:       "DockerMachineTemplate",
						Name:       getMachineDeploymentName(&manager.ClusterOptions),
						Namespace:  manager.ClusterID.Namespace,
						APIVersion: infrav1.GroupVersion.String(),
					},
					Version: pointer.StringPtr(manager.ClusterOptions.KubeVersion),
				},
			},
		},
	}
}

// getCapiCluster return a CAPI Cluster object
func getCapiClusterDocker(createOpt *v1.ClusterOptions) *capiv1beta1.Cluster {
	return &capiv1beta1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       clusterKind,
			APIVersion: "cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getDefaultName(createOpt), // createOpt.Name
			Namespace: metav1.NamespaceDefault,
		},
		Spec: capiv1beta1.ClusterSpec{
			ClusterNetwork: &capiv1beta1.ClusterNetwork{
				ServiceDomain: "cluster.local",
				Services: &capiv1beta1.NetworkRanges{
					CIDRBlocks: []string{"10.128.0.0/12"},
				},
				Pods: &capiv1beta1.NetworkRanges{
					CIDRBlocks: []string{"192.168.0.0/16"},
				},
			},
			ControlPlaneRef: &corev1.ObjectReference{
				Kind:       "KubeadmControlPlane",
				Namespace:  metav1.NamespaceDefault,
				Name:       getControlplaneName(createOpt),
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
			},
			InfrastructureRef: &corev1.ObjectReference{
				Kind:       "DockerCluster",
				Namespace:  metav1.NamespaceDefault,
				Name:       createOpt.Name,
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			},
		},
	}
}

func debug(o interface{}) {
	y, err := yaml.Marshal(o)
	if err != nil {
		panic(err)
	}
	log.Printf("%s\n", string(y))
}

func ApplyCNI(ctx context.Context, c client.Client, tracker *remote.ClusterCacheTracker, clusterkey client.ObjectKey) error {
	// see if a calico resource already exists on the aerostation deployed cluster, if so, assume we have already applied CNI
	cniSA := corev1.ServiceAccount{}
	thing := types.NamespacedName{
		Namespace: "kube-system",
		Name:      "calico-kube-controllers",
	}

	cli, err := tracker.GetClient(ctx, clusterkey)
	if err != nil {
		return err
	}

	err = cli.Get(ctx, thing, &cniSA)
	if err == nil {
		log.Println("calico SA found, so assuming CNI is applied")
		return nil
	}

	if apierrors.IsNotFound(err) {
		// apply CNI logic follows
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		data, err := ioutil.ReadFile(path.Join(wd, "/manifests/calico.yaml"))
		if err != nil {
			return err
		}

		// use the client for the capi cluster, not the aerostation deployed cluster
		restcfg, err := aeroremote.RESTConfig(ctx, c, clusterkey)
		if err != nil {
			return err
		}

		return utils.ForEachObjectInYAML(ctx, restcfg, data, "", utils.ApplyResource)
	}

	if err != nil {
		log.Printf("error getting calico SA %s\n", err.Error())
	}
	return nil
}
