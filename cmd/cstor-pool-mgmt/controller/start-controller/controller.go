/*
Copyright 2018 The OpenEBS Authors.

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

package startcontroller

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	backupcontroller "github.com/aamir-tiwari-sumo/maya/cmd/cstor-pool-mgmt/controller/backup-controller"
	"github.com/aamir-tiwari-sumo/maya/cmd/cstor-pool-mgmt/controller/common"
	replicacontroller "github.com/aamir-tiwari-sumo/maya/cmd/cstor-pool-mgmt/controller/replica-controller"
	restorecontroller "github.com/aamir-tiwari-sumo/maya/cmd/cstor-pool-mgmt/controller/restore"
	"github.com/aamir-tiwari-sumo/maya/cmd/cstor-pool-mgmt/pool"

	clientset "github.com/aamir-tiwari-sumo/maya/pkg/client/generated/clientset/versioned"
	informers "github.com/aamir-tiwari-sumo/maya/pkg/client/generated/informers/externalversions"
	"github.com/aamir-tiwari-sumo/maya/pkg/signals"

	poolcontroller "github.com/aamir-tiwari-sumo/maya/cmd/cstor-pool-mgmt/controller/pool-controller"
	//poolcontroller2 "github.com/aamir-tiwari-sumo/maya/cmd/cstor-pool-mgmt/controller/new-pool-controller"
	//// for v1alpha2
	//clientset2 "github.com/aamir-tiwari-sumo/maya/pkg/client/generated/openebs.io/v1alpha2/clientset/internalclientset"
	//informers2 "github.com/aamir-tiwari-sumo/maya/pkg/client/generated/openebs.io/v1alpha2/informer/externalversions"
)

const (
	// NumThreads defines number of worker threads for resource watcher.
	NumThreads = 1
	// NumRoutinesThatFollow is for handling golang waitgroups.
	NumRoutinesThatFollow = 1
)

// StartControllers instantiates CStorPool and CStorVolumeReplica controllers
// and watches them.
func StartControllers(kubeconfig string) {
	// Set up signals to handle the first shutdown signal gracefully.
	stopCh := signals.SetupSignalHandler()

	cfg, err := getClusterConfig(kubeconfig)
	if err != nil {
		klog.Fatalf(err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building openebs clientset: %s", err.Error())
	}

	//TODO: Remove below code
	//openebsClient2, err := clientset.NewForConfig(cfg)
	//if err != nil {
	//	klog.Fatalf("Error building openebs clientset: %s", err.Error())
	//}

	common.Init()

	// Blocking call for checking status of zrepl running in cstor-pool container.
	pool.CheckForZreplInitial(common.InitialZreplRetryInterval)
	go func() {
		// CheckForZreplContinuous is continuous health checker for status of
		// zrepl in cstor-pool container.
		// When zrepl is getting terminated and restarted very fast: zpool status
		// goroutine may miss this failure. To resolve, we’ll give InitialTimeDelay y
		// for zrepl container such that the period(x) of the goroutine thread will
		// be half that of this initialTimeDelay y. (x = 1/2 y).
		pool.CheckForZreplContinuous(common.ContinuousZreplRetryInterval)
		klog.Errorf("Zrepl/Pool is not available, Shutting down")
		os.Exit(1)
	}()
	// Blocking call for checking status of CStorPool CRD.
	common.CheckForCStorPoolCRD(openebsClient)

	// Blocking call for checking status of CStorVolumeReplica CRD.
	common.CheckForCStorVolumeReplicaCRD(openebsClient)

	// NewSharedInformerFactory constructs a new instance of k8s sharedInformerFactory.
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, getSyncInterval())

	// openebsInformerFactory constructs a new instance of openebs sharedInformerFactory.
	openebsInformerFactory := informers.NewSharedInformerFactory(openebsClient, getSyncInterval())

	// openebsInformerFactory2 constructs a new instance of openebs sharedInformerFactory.
	//openebsInformerFactory2 := informers.NewSharedInformerFactory(openebsClient2, getSyncInterval())

	//// Instantiate the cStor Pool and VolumeReplica controllers.
	cStorPoolController := poolcontroller.NewCStorPoolController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	volumeReplicaController := replicacontroller.NewCStorVolumeReplicaController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	// Instantiate the cStor backup controller
	backupController := backupcontroller.NewCStorBackupController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	// Instantiate the cStor restore controller
	restoreController := restorecontroller.NewCStorRestoreController(kubeClient, openebsClient, kubeInformerFactory,
		openebsInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go openebsInformerFactory.Start(stopCh)
	//go openebsInformerFactory2.Start(stopCh)

	// Waitgroup for starting pool and VolumeReplica controller goroutines.
	var wg sync.WaitGroup
	//TODO: Remove below code
	wg.Add(NumRoutinesThatFollow)

	// Run controller for cStorPool.
	go func() {
		if err = cStorPoolController.Run(NumThreads, stopCh); err != nil {
			klog.Fatalf("Error running CStorPool controller: %s", err.Error())
		}
		wg.Done()
	}()

	// CheckForCStorPool tries to get pool name and blocks forever because
	// volumereplica can be created only if pool is present.
	common.CheckForCStorPool()

	wg.Add(NumRoutinesThatFollow)
	// Run controller for cStorVolumeReplica.
	go func() {
		if err = volumeReplicaController.Run(NumThreads, stopCh); err != nil {
			klog.Fatalf("Error running CStorVolumeReplica controller: %s", err.Error())
		}
		wg.Done()
	}()

	wg.Add(NumRoutinesThatFollow)
	// Run controller for CStorBackup
	go func() {
		if err = backupController.Run(NumThreads, stopCh); err != nil {
			klog.Fatalf("Error running CStorBackup controller: %s", err.Error())
		}
		wg.Done()
	}()

	wg.Add(NumRoutinesThatFollow)
	// Run controller for CStorRestore.
	go func() {
		if err = restoreController.Run(NumThreads, stopCh); err != nil {
			klog.Fatalf("Error running CStorRestore controller: %s", err.Error())
		}
		wg.Done()
	}()

	wg.Wait()
}

// GetClusterConfig return the config for k8s.
func getClusterConfig(kubeconfig string) (*rest.Config, error) {
	var masterURL string
	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("Failed to get k8s Incluster config. %+v", err)
		if len(kubeconfig) == 0 {
			return nil, fmt.Errorf("kubeconfig is empty: %v", err.Error())
		}
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("Error building kubeconfig: %s", err.Error())
		}
	}
	return cfg, err
}

// getSyncInterval gets the resync interval from environment variable.
// If missing or zero then default to SharedInformerInterval
// otherwise return the obtained value
func getSyncInterval() time.Duration {
	resyncInterval, err := strconv.Atoi(os.Getenv("RESYNC_INTERVAL"))
	if err != nil || resyncInterval == 0 {
		klog.Warningf("Incorrect resync interval %q obtained from env, defaulting to %q seconds", resyncInterval, common.SharedInformerInterval)
		return common.SharedInformerInterval
	}
	return time.Duration(resyncInterval) * time.Second
}
