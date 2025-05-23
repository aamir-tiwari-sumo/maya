/*
Copyright 2020 The OpenEBS Authors

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

package volume

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/v1alpha1"
	cv "github.com/aamir-tiwari-sumo/maya/pkg/cstor/volume/v1alpha1"
	cvr "github.com/aamir-tiwari-sumo/maya/pkg/cstor/volumereplica/v1alpha1"
	pod "github.com/aamir-tiwari-sumo/maya/pkg/kubernetes/pod/v1alpha1"
	"github.com/aamir-tiwari-sumo/maya/tests"
	"github.com/aamir-tiwari-sumo/maya/tests/cstor"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// This Test Can Be Run By Using Following Command
// ginkgo -v -focus="\[WAITFORFIRSTCONSUMER\] " -- -kubeconfig=<path_to_kube_config> -cstor-maxpools=<no.of_pools> -cstor-replicas=<no.of_storagereplicas>

var _ = Describe("[WAITFORFIRSTCONSUMER] CStor Volume Provisioning", func() {
	When("SPC is created", func() {
		It("cStor Pools Should be Provisioned ", func() {

			By("Build And Create StoragePoolClaim object")
			// Populate configurations and create
			spcConfig := &tests.SPCConfig{
				Name:                spcName,
				DiskType:            "sparse",
				PoolCount:           cstor.PoolCount,
				IsThickProvisioning: true,
				PoolType:            "striped",
			}
			ops.Config = spcConfig
			spcObj = ops.BuildAndCreateSPC()
			By("Creating SPC, Desired Number of CSP Should Be Created", func() {
				ops.VerifyDesiredCSPCount(spcObj, cstor.PoolCount)
			})
		})
	})

	When("Apply Volume Related Artifacts", func() {
		It("Volume Should be Created and Provisioned", func() {
			By("Build And Create StorageClass")

			casConfig := strings.Replace(
				openebsCASConfigValue, "$spcName", spcObj.Name, 1)
			casConfig = strings.Replace(
				casConfig, "$count", strconv.Itoa(cstor.ReplicaCount), 1)
			annotations[string(apis.CASTypeKey)] = string(apis.CstorVolume)
			annotations[string(apis.CASConfigKey)] = casConfig
			scConfig := &tests.SCConfig{
				Name:              scName,
				Annotations:       annotations,
				Provisioner:       openebsProvisioner,
				VolumeBindingMode: storagev1.VolumeBindingWaitForFirstConsumer,
			}
			ops.Config = scConfig
			scObj = ops.CreateStorageClass()

			pvcConfig := &tests.PVCConfig{
				Name:        pvcName,
				Namespace:   nsObj.Name,
				SCName:      scObj.Name,
				Capacity:    "5G",
				AccessModes: accessModes,
			}
			ops.Config = pvcConfig
			pvcObj = ops.BuildAndCreatePVC()
		})
	})

	When("Deploying BusyBox Application", func() {
		It("CStor Volume Related Resources Should Be Created and Become Healthy", func() {
			var err error
			// Deploying Application
			By("Building a busybox app pod deployment using above volume")
			appDeployment, err = ops.BuildAndDeployBusyBoxPod(
				"busybox-cstor",
				pvcObj.Name, pvcObj.Namespace,
				map[string]string{"app": "busybox"})
			Expect(err).ShouldNot(HaveOccurred(), "while building app deployement {%v}", err)

			By("Verifying pvc status as bound")

			// Verify health of CStor Volume Related CR's
			ops.VerifyVolumeStatus(pvcObj,
				cstor.ReplicaCount,
				cvr.PredicateList{cvr.IsHealthy()},
				cv.PredicateList{cv.IsHealthy()},
			)
			pvcObj, err = ops.PVCClient.
				WithNamespace(pvcObj.Namespace).
				Get(pvcObj.Name, metav1.GetOptions{})
			Expect(err).To(BeNil())
		})
	})

	When("Deleting Application, PVC and SPC", func() {
		It("Should Delete Volume and Pools Related to test", func() {
			err := ops.DeployClient.WithNamespace(nsObj.Name).
				Delete(appDeployment.Name, &metav1.DeleteOptions{})
			Expect(err).ShouldNot(HaveOccurred(), "while deleting application pod")
			cnt := ops.GetPodCountEventually(nsObj.Name, "app=busybox", pod.PredicateList{}, 0)
			Expect(cnt).Should(BeNumerically("==", 0), fmt.Sprintf("While waiting for application pod to scale down"))

			ops.DeletePersistentVolumeClaim(pvcObj.Name, pvcObj.Namespace)
			ops.VerifyVolumeResources(pvcObj.Spec.VolumeName, openebsNamespace, cvr.PredicateList{}, cv.PredicateList{})
			err = ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())

			err = ops.SPCClient.Delete(
				spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting spc {%s}", spcObj.Name)
		})
	})
})
