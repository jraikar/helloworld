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

package main

import (
	"context"
	"flag"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	rbac "k8s.io/api/rbac/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	cliflag "k8s.io/component-base/cli/flag"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	aerostationv1 "github.com/aerospike/aerostation/api/v1"
	"github.com/aerospike/aerostation/controllers"
	"github.com/spf13/pflag"
	capiawsv1beta1 "sigs.k8s.io/cluster-api-provider-aws/api/v1beta1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	kubeadmcontrolplanev1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aerospike/aerospike-kubernetes-operator/api/v1beta1"
	aerov1b1 "github.com/aerospike/aerospike-kubernetes-operator/api/v1beta1"
	fleet "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	eksbootstrapv1beta1 "sigs.k8s.io/cluster-api-provider-aws/bootstrap/eks/api/v1beta1"
	ekscontrolplanev1beta1 "sigs.k8s.io/cluster-api-provider-aws/controlplane/eks/api/v1beta1"
	expv1beta1 "sigs.k8s.io/cluster-api-provider-aws/exp/api/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme     = runtime.NewScheme()
	setupLog   = ctrl.Log.WithName("setup")
	syncPeriod time.Duration
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(aerostationv1.AddToScheme(scheme))
	utilruntime.Must(capiv1beta1.AddToScheme(scheme))
	utilruntime.Must(capiawsv1beta1.AddToScheme(scheme))

	utilruntime.Must(kubeadmv1beta1.AddToScheme(scheme))
	utilruntime.Must(kubeadmcontrolplanev1beta1.AddToScheme(scheme))
	utilruntime.Must(rbac.AddToScheme(scheme))
	utilruntime.Must(v1beta1.AddToScheme(scheme))
	utilruntime.Must(ekscontrolplanev1beta1.AddToScheme(scheme))
	utilruntime.Must(eksbootstrapv1beta1.AddToScheme(scheme))
	utilruntime.Must(capiexp.AddToScheme(scheme))
	utilruntime.Must(expv1beta1.AddToScheme(scheme))
	utilruntime.Must(aerov1b1.AddToScheme(scheme))
	utilruntime.Must(fleet.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func InitFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&syncPeriod, "sync-period", 30*time.Minute,
		"The minimum interval at which watched resources are reconciled (e.g. 15m)")
}

func main() {
 
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader elecadftion for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	InitFlags(pflag.CommandLine)
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctx := ctrl.SetupSignalHandler()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		SyncPeriod:             &syncPeriod,
		LeaderElectionID:       "ddf80a57.aerospike.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	setupReconcilers(ctx, mgr)

	//+kubebuilder:scaffold:builder
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func setupReconcilers(ctx context.Context, mgr ctrl.Manager) {
	// Set up a ClusterCacheTracker to provide to controllers
	// requiring a connection to a remote cluster

	tracker, err := remote.NewClusterCacheTracker(
		mgr,
		remote.ClusterCacheTrackerOptions{
			Log: ctrl.Log.WithName("remote").WithName("ClusterCacheTracker"),
		},
	)
	if err != nil {
		setupLog.Error(err, "unable to create cluster cache tracker")
		os.Exit(1)
	}

	aeroDbReconciler := &controllers.AeroDatabaseReconciler{
		Client:  mgr.GetClient(),
		Log:     ctrl.Log.WithName("controllers").WithName("AeroDatabase"),
		Scheme:  mgr.GetScheme(),
		Tracker: tracker,
	}
	if err = aeroDbReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AeroDatabase")
		os.Exit(1)
	}
	if err = (&controllers.AeroClusterManagerReconciler{
		Client:             mgr.GetClient(),
		Log:                ctrl.Log.WithName("controllers").WithName("AeroClusterManager"),
		Scheme:             mgr.GetScheme(),
		DatabaseReconciler: aeroDbReconciler,
		Tracker:            tracker,
		Config:             mgr.GetConfig(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AeroClusterManager")
		os.Exit(1)
	}

 
}
import {
	
	"context"
	"flag"
	"os"
	"time"
}
