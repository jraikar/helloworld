/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/rest"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	v1 "github.com/aerospike/aerostation/api/v1"
	"github.com/aerospike/aerostation/pkg/utils"
	"github.com/aerospike/aerostation/pkg/utils/annotations"
	"github.com/aerospike/aerostation/pkg/utils/capi"
	corev1 "k8s.io/api/core/v1"

	v1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/cluster-api/util"

	fleet "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"sigs.k8s.io/cluster-api/util/patch"
)

const (
	AerostationFinalizerName = "aerostation.aerospike.com/finalizer"
	GithubSecret             = "github-ssh-auth"
)

// AeroClusterManagerReconciler reconciles a AeroClusterManager object
type AeroClusterManagerReconciler struct {
	client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	DatabaseReconciler *AeroDatabaseReconciler
	Kubectl            *kube.KubectlCmd
	Tracker            *remote.ClusterCacheTracker
	Config             *rest.Config
}

// SetupWithManager sets up the controller with the Manager.
func (r *AeroClusterManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.AeroClusterManager{}).
		Complete(r)
}

//+kubebuilder:rbac:groups=aerostation.aerospike.com,resources=aeroclustermanagers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aerostation.aerospike.com,resources=aeroclustermanagers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aerostation.aerospike.com,resources=aeroclustermanagers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;patch
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=fleet.cattle.io;machinedeployments.cluster.x-k8s.io;infrastructure.cluster.x-k8s.io;bootstrap.cluster.x-k8s.io;controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=*;machinedeployments;clusters;clusters/status,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrole,verbs=get;list;watch;create;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebinding,verbs=get;list;watch;create;patch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *AeroClusterManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	_ = r.Log.WithValues("aeroclustermanager", req.NamespacedName)

	manager := &v1.AeroClusterManager{}
	if err := r.Client.Get(ctx, req.NamespacedName, manager); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Make sure we cleanup if deleting
	if manager.ObjectMeta.DeletionTimestamp.IsZero() {
		if !utils.ContainsString(manager.GetFinalizers(), AerostationFinalizerName) {
			controllerutil.AddFinalizer(manager, AerostationFinalizerName)
			if err := r.Update(ctx, manager); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if utils.ContainsString(manager.GetFinalizers(), AerostationFinalizerName) {
			// our finalizer is present, lets delete the cluster
			if err := r.deleteExternalResources(ctx, manager); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(manager, AerostationFinalizerName)
			if err := r.Update(ctx, manager); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Return early if the object or Cluster is paused.
	if annotations.IsSuspended(manager, manager) {
		r.Log.Info("Reconciliation is suspended for this cluster")
		return ctrl.Result{}, nil
	}

	// Fetch the Cluster instance.
	cluster := &v1beta1.Cluster{}

	// Get remote cluster deployments
	// Initialize the patch helper.
	patchHelper, err := patch.NewHelper(manager, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	// ensure phase is always patched
	defer func() {
		// Always reconcile the Status.Phase field.
		r.reconcilePhase(manager, cluster)

		// Always attempt to Patch the Cluster object and status after each reconciliation.
		// Patch ObservedGeneration only if the reconciliation completed successfully
		patchOpts := []patch.Option{}
		if reterr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := patchCluster(ctx, patchHelper, manager, patchOpts...); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()
	return r.reconcile(ctx, manager, cluster)
}

func (r *AeroClusterManagerReconciler) deleteExternalResources(ctx context.Context, manager *v1.AeroClusterManager) error {
	// If this cluster is not managed by us. don't do anything with it.
	if !manager.Spec.Managed {
		return nil
	}

	clusters, err := r.getAllAerospikeClustersForKubeCluster(ctx, manager)
	if err != nil {
		return nil
	}

	for _, cluster := range clusters {
		err := r.Delete(ctx, &cluster)
		if err != nil {
			fmt.Printf("error %s\n", err.Error())
		}
	}

	cluster := &v1beta1.Cluster{}

	if err := r.Client.Get(ctx, manager.Spec.ClusterID.ToObjectKey(), cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return r.Client.Delete(ctx, cluster)
}

func (r *AeroClusterManagerReconciler) getAllAerospikeClustersForKubeCluster(ctx context.Context, manager *v1.AeroClusterManager) ([]v1.AeroDatabase, error) {
	var aerospikeClusters v1.AeroDatabaseList
	err := r.Client.List(ctx, &aerospikeClusters)
	if err != nil {
		return nil, err
	}

	clusters := make([]v1.AeroDatabase, 0)

	for _, cluster := range aerospikeClusters.Items {
		if cluster.Spec.Cluster.Name == manager.Name && cluster.Spec.Cluster.Namespace == manager.Namespace {
			clusters = append(clusters, cluster)
		}
	}

	return clusters, nil
}

// reconcile handles normal reconciliation.
func (r *AeroClusterManagerReconciler) reconcile(ctx context.Context, manager *v1.AeroClusterManager, cluster *v1beta1.Cluster) (ctrl.Result, error) {
	phases := []func(context.Context, *v1.AeroClusterManager, *v1beta1.Cluster) (ctrl.Result, error){
		r.reconcileCluster,
		r.reconcileGitRepo,
		r.reconcileFleet,
	}

	// TODO: shoudl also remove databases linked to this kubernetes cluster
	res := ctrl.Result{}
	errs := []error{}
	for _, phase := range phases {
		// Call the inner reconciliation methods.
		phaseResult, err := phase(ctx, manager, cluster)
		if err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			continue
		}
		res = util.LowestNonZeroResult(res, phaseResult)
	}
	if len(errs) > 0 {
		fmt.Printf("requeue in:%d%s\n", res.RequeueAfter, "seconds")
		fmt.Println("errors, requeueing")
		return res, kerrors.NewAggregate(errs)
	}

	fmt.Println("ending loops, should not requeue")
	return ctrl.Result{Requeue: false}, nil
}

// reconcileCluster handles reconciliation of the capi cluster
func (r *AeroClusterManagerReconciler) reconcileCluster(ctx context.Context, manager *v1.AeroClusterManager, cluster *v1beta1.Cluster) (ctrl.Result, error) {
	if err := r.Client.Get(ctx, manager.Spec.ClusterID.ToObjectKey(), cluster); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, this means we have not created it yet, lets do that now!
			var err error
			switch manager.Spec.ClusterOptions.Provider {
			case "docker":
				err = capi.ApplyDocker(r.Client, &manager.Spec, r.Config)
			case "eks":
				if manager.Spec.ClusterOptions.EKSOptions == nil {
					fmt.Println("Manager Spec has errors")
				}
				err = capi.ApplyEks(r.Client, &manager.Spec.ClusterOptions, r.Config)
			}

			if err != nil {
				fmt.Println(err)
				// Error Writing the capi object. Requeuing
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	var err error
	switch manager.Spec.ClusterOptions.Provider {
	case "docker":
		err = capi.ApplyDocker(r.Client, &manager.Spec, r.Config)
	case "eks":
		err = capi.ApplyEks(r.Client, &manager.Spec.ClusterOptions, r.Config)
	}
	if err != nil {
		fmt.Println(err)
		// Error Writing the cathub.com/rancher/fleet/pkg/apis v0.0.0pi object. Requeuing
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AeroClusterManagerReconciler) reconcileFleet(ctx context.Context, manager *v1.AeroClusterManager, cluster *v1beta1.Cluster) (_ ctrl.Result, reterr error) {

	fmt.Println("[DEBUG] reconciling fleet cluster = ", cluster.Name)
	//manager initiated cluster registration
	fleetObj := &fleet.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	}

	err := r.Client.Get(ctx, client.ObjectKey{Name: cluster.Name, Namespace: cluster.Namespace}, fleetObj)

	fleetObj = &fleet.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "fleet.cattle.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"env": "dev", // todo: take labels from user's input
			},
		},
		Spec: fleet.ClusterSpec{
			KubeConfigSecret: cluster.Name + "-kubeconfig",
		},
	}

	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, r.Client.Create(ctx, fleetObj)
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.Client.Update(ctx, fleetObj)
}

func (r *AeroClusterManagerReconciler) reconcileGitRepo(ctx context.Context, manager *v1.AeroClusterManager, cluster *v1beta1.Cluster) (_ ctrl.Result, reterr error) {

	fmt.Println("[DEBUG] reconciling gitRepo")
	//manager initiated cluster registration
	gitRepoObj := &fleet.GitRepo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aerostation-multi-cluster-fleet",
			Namespace: corev1.NamespaceDefault,
		},
	}

	err := r.Client.Get(ctx, client.ObjectKey{Name: cluster.Name, Namespace: cluster.Namespace}, gitRepoObj)

	gitRepoObj = &fleet.GitRepo{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitRepo",
			APIVersion: "fleet.cattle.io/v1alpha1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      "aerostation-multi-cluster-fleet",
			Namespace: cluster.Namespace,
		},

		//todo: take values from config
		Spec: fleet.GitRepoSpec{
			//Repo:                  "git@github.com:citrusleaf/aerostation-workload-deploy.git",
			Repo:                  "git@github.com:jmccormick2001/aerostation-workload-deploy-1.git",
			Branch:                "main",
			ClientSecretName:      GithubSecret,
			InsecureSkipTLSverify: false,
			Paths: []string{
				"dedicated/base/cert-manager",
				"dedicated/base/istio",
				"dedicated/base/istio-envoyfilters",
				"dedicated/base/istio-addons",
				"dedicated/base/ako",
			},
			Targets: []fleet.GitTarget{
				{
					Name:            "dev",
					ClusterSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"env": "dev"}}},
			},
		},
	}

	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, r.Client.Create(ctx, gitRepoObj)
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.Client.Update(ctx, gitRepoObj)
}

// reconcileOperator handles the reconcilation of operator related tasks
/**
func (r *AeroClusterManagerReconciler) reconcileOperator(ctx context.Context, manager *v1.AeroClusterManager, cluster *v1beta1.Cluster) (_ ctrl.Result, reterr error) {
	clusterkey := util.ObjectKey(cluster)

	cli, reterr := r.Tracker.GetClient(ctx, clusterkey)

	if reterr != nil {
		return ctrl.Result{}, reterr
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "aerospike",
		},
	}

	err := cli.Get(ctx, client.ObjectKey{Name: "aerospike"}, namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = cli.Create(ctx, namespace)

			if err != nil {
				_ = r.Log.WithValues("aeroclustermanager", "unable to create namespace")
				// we cant really do more at this point.
				return ctrl.Result{}, errors.Wrapf(err, "failed to get remote client %s/%s", manager.Namespace, manager.Name)
			}
		} else {
			// Unknown Error
			return ctrl.Result{}, err
		}
	}

	akoKey := client.ObjectKey{
		Name: "aerospikeclusters.aerospike.com",
	}
	exists := &unstructured.Unstructured{}
	err = cli.Get(ctx, akoKey, exists)

	if err != nil || exists == nil {

		err = utils.ApplyCertManager(ctx, r.Client, clusterkey)

		if err != nil {
			fmt.Println("Failed to apply CertManager")
		}

		err = utils.ApplyOperatorator(ctx, r.Client, clusterkey)

		if err != nil {

			fmt.Println("Failed to apply operator")
			return ctrl.Result{}, err
		}

		// TODO: check everything is deployed correctly?
		manager.Status.AerospikeOperator.Running = true

	}

	fmt.Println("Operator loop finished")

	_ = r.Log.WithValues("aeroclustermanager", "remote cluster now has the aerospike kubernetes operator installed")

	return ctrl.Result{}, reterr
}
*/

func (r *AeroClusterManagerReconciler) reconcilePhase(manager *v1.AeroClusterManager, cluster *v1beta1.Cluster) {
	if manager.Status.Phase == "" {
		manager.Status.SetTypedPhase(v1.ManagerPhasePending)
	}

	// Does nto handle cluster error states just yet
	if cluster == nil || v1beta1.ClusterPhase(cluster.Status.Phase) != v1beta1.ClusterPhaseProvisioned {
		manager.Status.SetTypedPhase(v1.ManagerPhaseClusterCreating)
	}

	if v1beta1.ClusterPhase(cluster.Status.Phase) == v1beta1.ClusterPhaseProvisioned {
		if manager.Spec.ClusterOptions.Provider == "docker" {
			fmt.Println("checking on CNI...")
			err := capi.ApplyCNI(context.Background(), r.Client, r.Tracker, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace})
			if err != nil {
				fmt.Printf("error in applying CNI %s\n", err.Error())
			}
		}

		manager.Status.SetTypedPhase(v1.ManagerPhaseOperatorInstalling)
	}

	if manager.Status.AerospikeOperator.Running {
		manager.Status.SetTypedPhase(v1.ManagerPhaseProvisioned)
	}
}

// ClusterToCluster Maps capi clusters to aerospike clusters
func (r *AeroClusterManagerReconciler) ClusterToCluster(o client.Object) []ctrl.Request {
	result := []ctrl.Request{}

	fmt.Println("Updating aerocluster manager based on capi cluster update")

	ms, ok := o.(*v1beta1.Cluster)

	if !ok {
		panic(fmt.Sprintf("Expected a Capi Cluster but got a %T", o))
	}

	c, err := r.getClusterForCluster(context.Background(), ms)
	if err != nil {
		return result
	}

	name := client.ObjectKey{Namespace: c.Namespace, Name: c.Name}
	result = append(result, ctrl.Request{NamespacedName: name})

	return result
}

func (r *AeroClusterManagerReconciler) getClusterForCluster(ctx context.Context, cluster *v1beta1.Cluster) (*v1.AeroClusterManager, error) {
	c := &v1.AeroClusterManager{}

	err := r.Get(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, c)

	return c, err
}

func patchCluster(ctx context.Context, patchHelper *patch.Helper, cluster *v1.AeroClusterManager, options ...patch.Option) error {
	// TODO: understand how/why we update the ready condition also what options
	// need set. something like patch.WithOwnedConditions
	return patchHelper.Patch(ctx, cluster, options...)
}
