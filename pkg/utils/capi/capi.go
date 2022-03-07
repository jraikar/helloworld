package capi

import (
	"context"
	"fmt"
	"log"

	v1 "github.com/aerospike/aerostation/api/v1"
	"github.com/aerospike/aerostation/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	capiawsv1beta1 "sigs.k8s.io/cluster-api-provider-aws/api/v1beta1"
	awsbootstrap "sigs.k8s.io/cluster-api-provider-aws/bootstrap/eks/api/v1beta1"
	ekscontrolplanev1beta1 "sigs.k8s.io/cluster-api-provider-aws/controlplane/eks/api/v1beta1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	capiQuickstartControlPlaneDefaultName = "-control-plane"
	capiQuickstartInfraDefaultName        = "-md-0"

	eksControlPlaneKind   = "AWSManagedControlPlane"
	clusterKind           = "Cluster"
	machineDeploymentKind = "MachineDeployment"

	awsMachineTemplateKind = "AWSMachineTemplate"
	eksConfigTemplateKind  = "EKSConfigTemplate"
)

func getControlplaneName(eks *v1.ClusterOptions) string {
	return fmt.Sprintf("%s%s", eks.Name, capiQuickstartControlPlaneDefaultName)
}

func getDefaultName(eks *v1.ClusterOptions) string {
	return eks.Name
}

func getMachineDeploymentName(eks *v1.ClusterOptions) string {
	return fmt.Sprintf("%s%s", eks.Name, capiQuickstartInfraDefaultName)
}

func ApplyEks(kubeClient client.Client, eksOptions *v1.ClusterOptions, config *rest.Config) error {
	cluster := getCapiCluster(eksOptions)
	log.Println("[INFO] creating Cluster")
	if err := utils.ApplyObject(context.Background(), cluster, config); err != nil {
		log.Print("Failed to create cluster: ", err)
		return err
	}

	awsManagedControlPlane := getAWSManagedControlPlane(eksOptions)
	log.Println("[INFO] creating AWSManagedControlPlane")
	if err := utils.ApplyObject(context.Background(), awsManagedControlPlane, config); err != nil {
		log.Print("Failed to create AWSManagedControlPlane: ", err)
		return err
	}

	machineDeployment := getMachineDeployment(eksOptions)
	log.Println("[INFO] creating MachineDeployment")
	if err := utils.ApplyObject(context.Background(), machineDeployment, config); err != nil {
		log.Print("Failed to create MachinePool: ", err)
		return err
	}

	machineTemplate := getMachineTemplate(eksOptions)
	log.Println("[INFO] creating MachineTemplate")
	if err := utils.ApplyObject(context.Background(), machineTemplate, config); err != nil {
		log.Print("Failed to create MachinePool: ", err)
		return err
	}

	eksConfigTemplate := getEksConfigTemplate(eksOptions)
	log.Println("[INFO] creating EksConfigTemplate")
	if err := utils.ApplyObject(context.Background(), eksConfigTemplate, config); err != nil {
		log.Print("Failed to create MachinePool: ", err)
		return err
	}
	return nil
}

// getCapiCluster return a CAPI Cluster object
func getCapiCluster(createOpt *v1.ClusterOptions) *capiv1beta1.Cluster {
	return &capiv1beta1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       clusterKind,
			APIVersion: capiv1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getDefaultName(createOpt), // createOpt.Name
			Namespace: metav1.NamespaceDefault,
		},
		Spec: capiv1beta1.ClusterSpec{
			ClusterNetwork: &capiv1beta1.ClusterNetwork{
				Pods: &capiv1beta1.NetworkRanges{
					CIDRBlocks: []string{"192.168.0.0/16"},
				},
			},
			ControlPlaneRef: &corev1.ObjectReference{
				Kind:       eksControlPlaneKind,
				Namespace:  metav1.NamespaceDefault,
				Name:       getControlplaneName(createOpt),
				APIVersion: ekscontrolplanev1beta1.GroupVersion.String(),
			},
			InfrastructureRef: &corev1.ObjectReference{
				Kind:       eksControlPlaneKind,
				Namespace:  metav1.NamespaceDefault,
				Name:       getControlplaneName(createOpt),
				APIVersion: ekscontrolplanev1beta1.GroupVersion.String(),
			},
		},
	}
}

// getAWSManagedControlPlane returns CAPI AWSManagedControlPlane template object
func getAWSManagedControlPlane(createOpts *v1.ClusterOptions) *ekscontrolplanev1beta1.AWSManagedControlPlane {
	return &ekscontrolplanev1beta1.AWSManagedControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       eksControlPlaneKind,
			APIVersion: ekscontrolplanev1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getControlplaneName(createOpts),
			Namespace: metav1.NamespaceDefault,
		},
		Spec: ekscontrolplanev1beta1.AWSManagedControlPlaneSpec{
			// EKSClusterName: getDefaultName(createOpts),
			Region:     createOpts.EKSOptions.Region,
			SSHKeyName: &createOpts.EKSOptions.SSHKey,
			Version:    pointer.String(createOpts.KubeVersion),
		},
	}
}

// getMachineDeployment
func getMachineDeployment(createOpts *v1.ClusterOptions) *capiv1beta1.MachineDeployment {
	return &capiv1beta1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       machineDeploymentKind,
			APIVersion: capiv1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        getMachineDeploymentName(createOpts),
			Namespace:   metav1.NamespaceDefault,
			ClusterName: getDefaultName(createOpts),
		},
		Spec: capiv1beta1.MachineDeploymentSpec{
			ClusterName: getDefaultName(createOpts),
			Replicas:    pointer.Int32Ptr(createOpts.Replicas),
			Template: capiv1beta1.MachineTemplateSpec{
				Spec: capiv1beta1.MachineSpec{
					Bootstrap: capiv1beta1.Bootstrap{
						ConfigRef: &corev1.ObjectReference{
							Kind:       eksConfigTemplateKind,
							Name:       getMachineDeploymentName(createOpts),
							APIVersion: awsbootstrap.GroupVersion.String(),
						},
					},
					ClusterName: getDefaultName(createOpts),
					InfrastructureRef: corev1.ObjectReference{
						Kind:       awsMachineTemplateKind,
						Name:       getMachineDeploymentName(createOpts),
						APIVersion: capiawsv1beta1.GroupVersion.String(),
					},
					Version: pointer.StringPtr(createOpts.KubeVersion),
				},
			},
		},
	}
}

func getMachineTemplate(createOpts *v1.ClusterOptions) *capiawsv1beta1.AWSMachineTemplate {
	return &capiawsv1beta1.AWSMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       awsMachineTemplateKind,
			APIVersion: capiawsv1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        getMachineDeploymentName(createOpts),
			Namespace:   metav1.NamespaceDefault,
			ClusterName: getDefaultName(createOpts),
		},
		Spec: capiawsv1beta1.AWSMachineTemplateSpec{
			Template: capiawsv1beta1.AWSMachineTemplateResource{
				Spec: capiawsv1beta1.AWSMachineSpec{
					IAMInstanceProfile: "nodes.cluster-api-provider-aws.sigs.k8s.io",
					InstanceType:       createOpts.EKSOptions.InstanceType,
					SSHKeyName:         pointer.StringPtr(createOpts.EKSOptions.SSHKey),
				},
			},
		},
	}
}

func getEksConfigTemplate(createOpts *v1.ClusterOptions) *awsbootstrap.EKSConfigTemplate {
	return &awsbootstrap.EKSConfigTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       eksConfigTemplateKind,
			APIVersion: awsbootstrap.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getMachineDeploymentName(createOpts),
			Namespace: metav1.NamespaceDefault,
		},
		Spec: awsbootstrap.EKSConfigTemplateSpec{
			Template: awsbootstrap.EKSConfigTemplateResource{},
		},
	}
}
