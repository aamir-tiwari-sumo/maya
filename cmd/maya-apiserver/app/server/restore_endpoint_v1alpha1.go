/*
Copyright 2019 The OpenEBS Authors

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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/aamir-tiwari-sumo/maya/pkg/client/generated/clientset/versioned"
	"github.com/aamir-tiwari-sumo/maya/pkg/volume"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

type restoreAPIOps struct {
	req  *http.Request
	resp http.ResponseWriter
}

// restoreV1alpha1SpecificRequest deals with restore API requests
func (s *HTTPServer) restoreV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	restoreOp := &restoreAPIOps{
		req:  req,
		resp: resp,
	}

	switch req.Method {
	case "POST":
		return restoreOp.create()
	case "GET":
		return restoreOp.get()
	}
	return nil, CodedError(405, ErrInvalidMethod)
}

// Create is http handler which handles restore-create request
func (rOps *restoreAPIOps) create() (interface{}, error) {
	var err error
	var openebsClient *versioned.Clientset
	restore := &v1alpha1.CStorRestore{}
	err = decodeBody(rOps.req, restore)
	if err != nil {
		return nil, err
	}

	// namespace is expected
	if !restore.Spec.Local && len(strings.TrimSpace(restore.Namespace)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create restore '%v': missing namespace", restore.Name))
	}

	// restore name is expected
	if len(strings.TrimSpace(restore.Spec.RestoreName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create restore: missing restore name "))
	}

	// volume name is expected
	if len(strings.TrimSpace(restore.Spec.VolumeName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create restore '%v': missing volume name", restore.Name))
	}

	// restoreIP is expected
	if len(strings.TrimSpace(restore.Spec.RestoreSrc)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create restore '%v': missing restoreSrc", restore.Name))
	}

	// storageClass is expected if restore is for local snapshot
	if restore.Spec.Local && len(strings.TrimSpace(restore.Spec.StorageClass)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create restore '%v': missing storageClass", restore.Name))
	}

	// size is expected if restore is for local snapshot
	if restore.Spec.Local && len(strings.TrimSpace(restore.Spec.Size.String())) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create restore '%v': missing size", restore.Name))
	}

	openebsClient, _, err = loadClientFromServiceAccount()
	if err != nil {
		return nil, CodedError(400, fmt.Sprintf("Failed to load openebs client:{%v}", err))
	}

	cvol, err := createVolumeForRestore(restore)
	if err != nil {
		return nil, CodedError(400, fmt.Sprintf("Failed to create resources for volume: {%v}", err))
	}
	klog.Infof("Restore volume '%v' created successfully ", cvol.Name)

	if restore.Spec.Local {
		return cvol, nil
	}

	return createRestoreResource(openebsClient, restore, cvol)
}

// createRestoreResource create restore CR for volume's CVR
func createRestoreResource(openebsClient *versioned.Clientset, rst *v1alpha1.CStorRestore, cvol *v1alpha1.CASVolume) (interface{}, error) {
	//Get List of cvr's related to this pvc
	listOptions := v1.ListOptions{
		LabelSelector: "openebs.io/persistent-volume=" + rst.Spec.VolumeName,
	}
	cvrList, err := openebsClient.OpenebsV1alpha1().CStorVolumeReplicas("").List(context.TODO(), listOptions)
	if err != nil {
		return nil, CodedError(500, err.Error())
	}

	for _, cvr := range cvrList.Items {
		rst.Name = rst.Spec.RestoreName + "-" + string(uuid.NewUUID())
		oldrst, err := openebsClient.OpenebsV1alpha1().CStorRestores(rst.Namespace).Get(context.TODO(), rst.Name, v1.GetOptions{})
		if err != nil {
			rst.Status = v1alpha1.RSTCStorStatusPending
			rst.ObjectMeta.Labels = map[string]string{
				"cstorpool.openebs.io/uid":     cvr.ObjectMeta.Labels["cstorpool.openebs.io/uid"],
				"openebs.io/persistent-volume": cvr.ObjectMeta.Labels["openebs.io/persistent-volume"],
				"openebs.io/restore":           rst.Spec.RestoreName,
			}

			_, err = openebsClient.OpenebsV1alpha1().CStorRestores(rst.Namespace).
				Create(context.TODO(), rst, v1.CreateOptions{})
			if err != nil {
				klog.Errorf("Failed to create restore CR(volume:%s CSP:%s) : error '%s'",
					rst.Spec.VolumeName, cvr.ObjectMeta.Labels["cstorpool.openebs.io/uid"],
					err.Error())
				return nil, CodedError(500, err.Error())
			}
			klog.Infof("Restore:%s created for volume %q poolUUID:%v", rst.Name,
				rst.Spec.VolumeName,
				rst.ObjectMeta.Labels["cstorpool.openebs.io/uid"])
		} else {
			oldrst.Status = v1alpha1.RSTCStorStatusPending
			oldrst.Spec = rst.Spec
			_, err = openebsClient.OpenebsV1alpha1().CStorRestores(oldrst.Namespace).
				Update(context.TODO(), oldrst, v1.UpdateOptions{})
			if err != nil {
				klog.Errorf("Failed to re-initialize old existing restore CR(volume:%s CSP:%s) : error '%s'",
					rst.Spec.VolumeName, cvr.ObjectMeta.Labels["cstorpool.openebs.io/uid"],
					err.Error())
				return nil, CodedError(500, err.Error())
			}
			klog.Infof("Re-initializing old restore:%s  %q poolUUID:%v", rst.Name,
				rst.Spec.VolumeName,
				rst.ObjectMeta.Labels["cstorpool.openebs.io/uid"])
		}
	}

	return cvol, nil
}

// get is http handler which handles backup get request
func (rOps *restoreAPIOps) get() (interface{}, error) {
	var err error
	var rstatus v1alpha1.CStorRestoreStatus
	var resp []byte

	rst := &v1alpha1.CStorRestore{}

	err = decodeBody(rOps.req, rst)
	if err != nil {
		return nil, err
	}

	// backup name is expected
	if len(strings.TrimSpace(rst.Spec.RestoreName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("Failed to get restore: missing restore name "))
	}

	// namespace is expected
	if len(strings.TrimSpace(rst.Namespace)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("Failed to get restore '%v': missing namespace", rst.Spec.RestoreName))
	}

	// volume name is expected
	if len(strings.TrimSpace(rst.Spec.VolumeName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("Failed to get restore '%v': missing volume name", rst.Spec.RestoreName))
	}

	rstatus, err = getRestoreStatus(rst)
	if err != nil {
		return nil, CodedError(400, fmt.Sprintf("Failed to fetch status '%v'", err))
	}

	resp, err = json.Marshal(rstatus)
	if err == nil {
		_, err = rOps.resp.Write(resp)
		if err != nil {
			return nil, CodedError(400, fmt.Sprintf("Failed to send response data"))
		}
		return nil, nil
	}

	return nil, CodedError(400, fmt.Sprintf("Failed to encode response data"))
}

func getRestoreStatus(rst *v1alpha1.CStorRestore) (v1alpha1.CStorRestoreStatus, error) {
	rstStatus := v1alpha1.RSTCStorStatusEmpty

	openebsClient, k8sClient, err := loadClientFromServiceAccount()
	if err != nil {
		return rstStatus, CodedError(400, fmt.Sprintf("Failed to create openEBSClient '%v'", err))
	}

	listOptions := v1.ListOptions{
		LabelSelector: "openebs.io/restore=" + rst.Spec.RestoreName + ",openebs.io/persistent-volume=" + rst.Spec.VolumeName,
	}

	rlist, err := openebsClient.OpenebsV1alpha1().CStorRestores(rst.Namespace).
		List(context.TODO(), listOptions)
	if err != nil {
		return v1alpha1.RSTCStorStatusEmpty, CodedError(400, fmt.Sprintf("Failed to fetch restore error:%v", err))
	}

	for _, nr := range rlist.Items {
		rstStatus = getCVRRestoreStatus(k8sClient, nr)

		switch rstStatus {
		case v1alpha1.RSTCStorStatusInProgress:
			rstStatus = v1alpha1.RSTCStorStatusInProgress
		case v1alpha1.RSTCStorStatusFailed, v1alpha1.RSTCStorStatusInvalid:
			if nr.Status != rstStatus {
				// Restore for given CVR may failed due to node failure or pool failure
				// Let's update status for given CVR's restore to failed
				updateRestoreStatus(openebsClient, nr, rstStatus)
			}
			rstStatus = v1alpha1.RSTCStorStatusFailed
		case v1alpha1.RSTCStorStatusDone:
			if rstStatus != v1alpha1.RSTCStorStatusFailed {
				rstStatus = v1alpha1.RSTCStorStatusDone
			}
		}

		klog.Infof("Restore:%v status is %v", nr.Name, nr.Status)

		if rstStatus == v1alpha1.RSTCStorStatusInProgress {
			break
		}
	}
	return rstStatus, nil
}

func getCVRRestoreStatus(k8sClient *kubernetes.Clientset, rst v1alpha1.CStorRestore) v1alpha1.CStorRestoreStatus {
	if rst.Status != v1alpha1.RSTCStorStatusDone && rst.Status != v1alpha1.RSTCStorStatusFailed {
		// check if node is running or not
		bkpNodeDown := checkIfCSPPoolNodeDown(k8sClient, rst.Labels["cstorpool.openebs.io/uid"])
		// check if cstor-pool-mgmt container is running or not
		bkpPodDown := checkIfCSPPoolPodDown(k8sClient, rst.Labels["cstorpool.openebs.io/uid"])

		if bkpNodeDown || bkpPodDown {
			// Backup is stalled, assume status as failed
			return v1alpha1.RSTCStorStatusFailed
		}
	}
	return rst.Status
}

// updateRestoreStatus will update the restore status to given status
func updateRestoreStatus(clientset versioned.Interface, rst v1alpha1.CStorRestore, status v1alpha1.CStorRestoreStatus) {
	rst.Status = status

	_, err := clientset.OpenebsV1alpha1().CStorRestores(rst.Namespace).
		Update(context.TODO(), &rst, v1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update restore:%s with status:%v", rst.Name, status)
		return
	}
}

func createVolumeForRestore(r *v1alpha1.CStorRestore) (*v1alpha1.CASVolume, error) {
	vol := &v1alpha1.CASVolume{}
	vol.Name = r.Spec.VolumeName
	vol.Labels = map[string]string{
		string(v1alpha1.StorageClassKey): r.Spec.StorageClass,
	}
	vol.Spec.Capacity = r.Spec.Size.String()

	if r.Spec.Local {
		vol.CloneSpec.IsClone = true
		vol.CloneSpec.SourceVolume = r.Spec.RestoreSrc
		vol.CloneSpec.SnapshotName = r.Spec.RestoreName
	} else {
		vol.Annotations = map[string]string{
			v1alpha1.PVCreatedByKey: "restore",
		}
	}

	vOps, err := volume.NewOperation(vol)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create volume operation")
	}

	// If Restore is from remote backup then volume creation is handled by velero-plugin
	// So let's check if volume exist or not
	cvol, err := vOps.Read()
	if err != nil {
		if !isNotFound(err) {
			return nil, errors.Wrapf(err, "Failed to get restore volume details")
		}
	} else {
		return cvol, nil
	}

	return vOps.Create()
}
