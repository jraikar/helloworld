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

	"github.com/aerospike/aerospike-kubernetes-operator/api/v1beta1"
	"github.com/aerospike/aerostation/pkg/utils"

	"github.com/aerospike/aerostation/pkg/utils/ako"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/cluster-api/util/patch"

	kerrors "k8s.io/apimachinery/pkg/util/errors"

	v1 "github.com/aerospike/aerostation/api/v1"
)

// AeroDatabaseReconciler reconciles a AeroDatabase object
type AeroDatabaseReconciler struct {
	client.Client
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Tracker *remote.ClusterCacheTracker
}

//+kubebuilder:rbac:groups=aerostation.aerospike.com,resources=aerodatabases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aerostation.aerospike.com,resources=aerodatabases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aerostation.aerospike.com,resources=aerodatabases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AeroDatabase object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *AeroDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("aerodatabase", req.NamespacedName)

	// need to know if this is an update, or a create. Maybe seperate phases?

	db := &v1.AeroDatabase{}
	if err := r.Client.Get(ctx, req.NamespacedName, db); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "failed to get aerodatabase for database %s/%s ", req.Namespace, req.Name)
	}

	cluster, err := r.getDatabaseForCluster(context.Background(), db)

	// Make sure we cleanup if deleting
	if db.ObjectMeta.DeletionTimestamp.IsZero() {
		if !utils.ContainsString(db.GetFinalizers(), AerostationFinalizerName) {
			controllerutil.AddFinalizer(db, AerostationFinalizerName)
			if err := r.Update(ctx, db); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if utils.ContainsString(db.GetFinalizers(), AerostationFinalizerName) {
			// our finalizer is present, lets delete the cluster
			if err := r.deleteExternalResources(ctx, db, cluster); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(db, AerostationFinalizerName)
			if err := r.Update(ctx, db); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "failed to find cluster for db %s/%s", req.Namespace, req.Name)
		// TODO, this should be a failure status, we can't create this DB until it's cluster is crated
	}

	var reterr error

	patchHelper, err := patch.NewHelper(db, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		patchOpts := []patch.Option{}
		if reterr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := patchHelper.Patch(ctx, db, patchOpts...); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		if reterr != nil {
			db.Status.LastError = reterr.Error()
		}

		/*
			if err := patchCluster(ctx, patchHelper, db, patchOpts...); err != nil {
				reterr = kerrors.NewAggregate([]error{reterr, err})
			}
		*/
	}()

	/**
	if cluster.Status.Phase != string(v1.ManagerPhaseProvisioned) {
		//Cluster is not ready, waiting for cluster to be ready. requeueing
		fmt.Println("waiting for cluster to be ready")
		db.Status.SetTypedPhase(v1.DBPhase(v1.DBPhasePending))
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Minute}, nil
	}
	*/

	if err != nil {
		return ctrl.Result{}, err
	}

	err = utils.ApplyDatabase(ctx, r.Client, db, ako.GetMemoryDB(db.Spec))

	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "failed to patch database %s/%s", req.Namespace, req.Name)
	}
	db.Status.SetTypedPhase(v1.DBPhase(v1.DBPhaseDeployed))

	return ctrl.Result{}, reterr
}

// SetupWithManager sets up the controller with the Manager.
func (r *AeroDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.AeroDatabase{}).
		Complete(r)
}

func (r *AeroDatabaseReconciler) DatabaseToCluster(o client.Object) []ctrl.Request {
	result := []ctrl.Request{}

	ms, ok := o.(*v1.AeroDatabase)

	if !ok {
		panic(fmt.Sprintf("Expected a AeroClusterManager but got a %T", o))
	}

	c, err := r.getDatabaseForCluster(context.Background(), ms)
	if err != nil {
		return result
	}

	name := client.ObjectKey{Namespace: c.Namespace, Name: c.Name}
	result = append(result, ctrl.Request{NamespacedName: name})

	return result
}

func (r *AeroDatabaseReconciler) getDatabaseForCluster(ctx context.Context, db *v1.AeroDatabase) (*v1.AeroClusterManager, error) {
	c := &v1.AeroClusterManager{}

	err := r.Get(ctx, db.Spec.Cluster.ToObjectKey(), c)

	return c, err
}

func (r *AeroDatabaseReconciler) deleteExternalResources(ctx context.Context, db *v1.AeroDatabase, kube *v1.AeroClusterManager) error {
	remoteClient, err := r.Tracker.GetClient(ctx, db.Spec.Cluster.ToObjectKey())
	if err != nil {
		// if kubernetes cluster does not exist, bad state.
		return nil
	}

	// TODO: Clean up dangling Feature Secret

	obj := &v1beta1.AerospikeCluster{}
	obj.Name = db.Spec.Name
	obj.Namespace = db.Spec.Namespace
	err = remoteClient.Delete(ctx, obj)

	if err != nil {
		return err
	}

	return nil
}
