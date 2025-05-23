/*
Copyright 2019 The OpenEBS Authors.

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

package webhook

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	snapshot "github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/snapshot/v1"
	"github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/aamir-tiwari-sumo/maya/pkg/client/generated/clientset/versioned"
	ndmclientset "github.com/aamir-tiwari-sumo/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset"
	snapclient "github.com/aamir-tiwari-sumo/maya/pkg/client/generated/openebs.io/snapshot/v1/clientset/internalclientset"
	"github.com/pkg/errors"
	admissionv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

const (
	snapshotMetadataPVName = "SnapshotMetadata-PVName"
	skipValidation         = "openebs.io/skip-validations"
)

// Skip validation in special namespaces, i.e. in kube-system and kube-public
// namespaces the validation will be skipped
var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
	snapshotAnnotation = "snapshot.alpha.kubernetes.io/snapshot"
)

// webhook implements a validating webhook.
type webhook struct {
	//  Server defines parameters for running an golang HTTP server.
	Server *http.Server

	// kubeClient is a standard kubernetes clientset
	kubeClient kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// snapClientSet is a snaphot custom resource package generated from custom API group.
	snapClientSet snapclient.Interface

	ndmClientset ndmclientset.Interface
}

// Parameters are server configures parameters
type Parameters struct {
	// Port is webhook server port
	Port int
	//CertFile is path to the x509 certificate for https
	CertFile string
	//KeyFile is path to the x509 private key matching `CertFile`
	KeyFile string
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = v1.AddToScheme(runtimeScheme)
}

// New creates a new instance of a webhook. Prior to
// invoking this function, InitValidationServer function must be called to
// set up secret (for TLS certs) k8s resource. This function runs forever.
func New(p Parameters, kubeClient kubernetes.Interface,
	openebsClient clientset.Interface,
	snapClient snapclient.Interface,
	ndmClient ndmclientset.Interface) (
	*webhook, error) {

	admNamespace, err := getOpenebsNamespace()
	if err != nil {
		return nil, err
	}

	// Fetch certificate secret information
	certSecret, err := GetSecret(admNamespace, validatorSecret)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read secret(%s) object %v",
			validatorSecret,
			err,
		)
	}

	// extract cert information from the secret object
	certBytes, ok := certSecret.Data[appCrt]
	if !ok {
		return nil, fmt.Errorf(
			"%s value not found in %s secret",
			appCrt,
			validatorSecret,
		)
	}
	keyBytes, ok := certSecret.Data[appKey]
	if !ok {
		return nil, fmt.Errorf(
			"%s value not found in %s secret",
			appKey,
			validatorSecret,
		)
	}

	signingCertBytes, ok := certSecret.Data[rootCrt]
	if !ok {
		return nil, fmt.Errorf(
			"%s value not found in %s secret",
			rootCrt,
			validatorSecret,
		)
	}

	certPool := x509.NewCertPool()
	ok = certPool.AppendCertsFromPEM(signingCertBytes)
	if !ok {
		return nil, fmt.Errorf("failed to parse root certificate")
	}

	sCert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return nil, err
	}

	//	pair, err := tls.LoadX509KeyPair(p.CertFile, p.KeyFile)
	//	if err != nil {
	//		return nil, err
	//	}
	wh := &webhook{
		Server: &http.Server{
			Addr:      fmt.Sprintf(":%v", p.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{sCert}},
		},
		kubeClient:    kubeClient,
		clientset:     openebsClient,
		snapClientSet: snapClient,
		ndmClientset:  ndmClient,
	}
	return wh, nil
}

func admissionRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			klog.V(4).Infof("Skip validation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
			return false
		}
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	var required bool
	switch strings.ToLower(annotations[skipValidation]) {
	default:
		required = true
	case "y", "yes", "true":
		klog.Infof("Skipping validations for %s/%s due to PVC has skip validation", metadata.Namespace, metadata.Name)
		required = false
	}
	return required
}

func validationRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	required := admissionRequired(ignoredList, metadata)
	klog.V(4).Infof("Validation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

// validatePVCDeleteRequest validates the persistentvolumeclaim(PVC) delete request
func (wh *webhook) validatePVCDeleteRequest(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	response := &admissionv1.AdmissionResponse{}
	response.Allowed = true

	// ignore the Delete request of PVC if resource name is empty which
	// can happen as part of cleanup process of namespace
	if req.Name == "" {
		return response
	}

	klog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	// TODO* use raw object once validation webhooks support DELETE request
	// object as non nil value https://github.com/kubernetes/kubernetes/issues/66536
	//var pvc corev1.PersistentVolumeClaim
	//err := json.Unmarshal(req.Object.Raw, &pvc)
	//if err != nil {
	//	klog.Errorf("Could not unmarshal raw object: %v, %v", err, req.Object.Raw)
	//	status.Allowed = false
	//	status.Result = &metav1.Status{
	//		Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
	//		Message: err.Error(),
	//	}
	//	return status
	//}

	// fetch the pvc specifications
	pvc, err := wh.kubeClient.CoreV1().PersistentVolumeClaims(req.Namespace).
		Get(context.TODO(), req.Name, metav1.GetOptions{})
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("error retrieving PVC: %v", err.Error()),
		}
		return response
	}

	// If PVC is not yet bound to PV then don't even perform any checks
	if pvc.Spec.VolumeName == "" {
		klog.Infof("Skipping validations for %s/%s due to PVC is not yet bound", pvc.Namespace, pvc.Name)
		return response
	}

	// skip pvc validation if skip-validations annotation has been set to true
	// "openebs.io/skip-validations: true"
	if !validationRequired(ignoredNamespaces, &pvc.ObjectMeta) {
		klog.V(4).Infof("Skipping validation for %s/%s due to policy check", pvc.Namespace, pvc.Name)
		return response
	}

	err = wh.verifyExistenceOfSnapshots(pvc)
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("snapshots verification failed %s", err.Error()),
		}
		return response
	}

	// construct source-volume label to list all the matched cstorVolumes
	label := fmt.Sprintf("openebs.io/source-volume=%s", pvc.Spec.VolumeName)
	listOptions := metav1.ListOptions{
		LabelSelector: label,
	}

	// get the all CStorVolumes resources in all namespaces based on the
	// source-volume label to verify if there is any clone volume exists.
	// if source-volume label matches with name of PV, failed the pvc
	// deletion operation.

	cStorVolumes, err := wh.getCstorVolumes(listOptions)
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("error retrieving CstorVolumes: %v", err.Error()),
		}
		return response
	}

	if len(cStorVolumes.Items) != 0 {
		response.Allowed = false
		response.Result = &metav1.Status{
			Status: metav1.StatusFailure, Code: http.StatusForbidden, Reason: "PVC with cloned volumes can't be deleted",
			Message: fmt.Sprintf("pvc %q has '%v' cloned volume(s)", pvc.Name, len(cStorVolumes.Items)),
		}
		return response
	}

	return response
}

// validatePVCCreateRequest validates persistentvolumeclaim(PVC) create request
func (wh *webhook) validatePVCCreateRequest(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	response := &admissionv1.AdmissionResponse{}
	response.Allowed = true

	var pvc corev1.PersistentVolumeClaim
	err := json.Unmarshal(req.Object.Raw, &pvc)
	if err != nil {
		klog.Errorf("Could not unmarshal raw object: %v, %v", err, req.Object.Raw)
		response.Allowed = false
		response.Result = &metav1.Status{
			Status:  metav1.StatusFailure,
			Code:    http.StatusBadRequest,
			Reason:  metav1.StatusReasonBadRequest,
			Message: err.Error(),
		}
		return response
	}

	// If snapshot.alpha.kubernetes.io/snapshot annotation represents the clone pvc
	// create request
	snapname := pvc.Annotations[snapshotAnnotation]
	if len(snapname) == 0 {
		return response
	}

	klog.V(4).Infof("AdmissionReview for creating a clone volume Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)
	// get the snapshot object to get snapshotdata object
	// Note: If snapname is empty then below call will retrun error
	snapObj, err := wh.snapClientSet.VolumesnapshotV1().VolumeSnapshots(pvc.Namespace).
		Get(context.TODO(), snapname, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("failed to get the snapshot object for snapshot name: '%s' namespace: '%s' PVC: '%s'"+
			"error: '%v'", snapname, pvc.Namespace, pvc.Name, err)
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("Failed to get the snapshot object for snapshot name: '%s' namespace: '%s' "+
				"error: '%v'", snapname, pvc.Namespace, err.Error()),
		}
		return response
	}

	snapDataName := snapObj.Spec.SnapshotDataName
	if len(snapDataName) == 0 {
		klog.Errorf("Snapshotdata name is empty for snapshot: '%s' snapshot Namespace: '%s' PVC: '%s'",
			snapname, snapObj.ObjectMeta.Namespace, pvc.Name)
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("Snapshotdata name is empty for snapshot: '%s' snapshot Namespace: '%s'",
				snapname, snapObj.ObjectMeta.Namespace),
		}
		return response
	}
	klog.V(4).Infof("snapshotdata name: '%s'", snapDataName)

	// get the snapDataObj to get the snapshotdataname
	// Note: If snapDataName is empty then below call will return error
	snapDataObj, err := wh.snapClientSet.VolumesnapshotV1().VolumeSnapshotDatas().
		Get(context.TODO(), snapDataName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get the snapshotdata object for snapshotdata  name: '%s' "+
			"snapName: '%s' namespace: '%s' PVC: '%s' error: '%v'", snapDataName, snapname, snapObj.ObjectMeta.Namespace, pvc.Name, err)
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("Failed to get the snapshotdata object for snapshotdata  name: '%s' "+
				"snapName: '%s' namespace: '%s' error: '%v'", snapDataName, snapname, snapObj.ObjectMeta.Namespace, err.Error()),
		}
		return response
	}

	snapSizeString := snapDataObj.Spec.OpenEBSSnapshot.Capacity
	// If snapshotdata object doesn't consist Capacity field then we will log it and return false.
	if len(snapSizeString) == 0 {
		klog.Infof("snapshot size not found for snapshot name: '%s' snapshot namespace: '%s' snapshotdata name: '%s'",
			snapname, snapObj.ObjectMeta.Namespace, snapDataName)
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("PVC: '%s' creation requires upgrade of volumesnapshotdata name: '%s'", pvc.ObjectMeta.Name, snapDataName),
		}
		return response
	}

	snapCapacity := resource.MustParse(snapSizeString)
	pvcSize := pvc.Spec.Resources.Requests[corev1.ResourceName(corev1.ResourceStorage)]
	if pvcSize.Cmp(snapCapacity) != 0 {
		klog.Errorf("Requested pvc size not matched the snapshot size '%s' belongs to snapshot name: '%s' "+
			"snapshot Namespace: '%s' VolumeSnapshotData '%s'", snapSizeString, snapObj.ObjectMeta.Name, snapObj.ObjectMeta.Namespace, snapDataName)
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("Requested pvc size must be equal to snapshot size '%s' "+
				"which belongs to snapshot name: '%s' snapshot NameSpace: '%s' volumesnapshotdata: '%s'",
				snapSizeString, snapObj.ObjectMeta.Name, snapObj.ObjectMeta.Namespace, snapDataName),
		}
		return response
	}
	return response
}

// validate validates the persistentvolumeclaim(PVC) create, delete request
func (wh *webhook) validate(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request
	response := &admissionv1.AdmissionResponse{}
	response.Allowed = true
	klog.Info("Admission webhook request received")
	switch req.Kind.Kind {
	case "Namespace":
		klog.V(2).Infof("Admission webhook request for type %s", req.Kind.Kind)
		return wh.validateNamespace(ar)
	case "PersistentVolumeClaim":
		klog.V(2).Infof("Admission webhook request for type %s", req.Kind.Kind)
		return wh.validatePVC(ar)
	case "StoragePoolClaim":
		klog.V(2).Infof("Admission webhook request for type %s", req.Kind.Kind)
		return wh.validateSPC(ar)
	default:
		klog.V(2).Infof("Admission webhook not configured for type %s", req.Kind.Kind)
		return response
	}
}

func (wh *webhook) validatePVC(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request
	response := &admissionv1.AdmissionResponse{}
	response.Allowed = true
	// validates only if requested operation is CREATE or DELETE
	if req.Operation == admissionv1.Create {
		return wh.validatePVCCreateRequest(req)
	} else if req.Operation == admissionv1.Delete {
		return wh.validatePVCDeleteRequest(req)
	}
	klog.V(2).Info("Admission wehbook for PVC module not " +
		"configured for operations other than DELETE and CREATE")
	return response
}

func (wh *webhook) validateNamespace(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request
	response := &admissionv1.AdmissionResponse{}
	response.Allowed = true
	openebsNamespace, err := getOpenebsNamespace()
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("error getting OPENEBS_NAMESPACE env %s: %v", req.Name, err.Error()),
		}
		return response
	}
	// validates only if requested operation is DELETE
	if openebsNamespace == req.Name && req.Operation == admissionv1.Delete {
		return wh.validateNamespaceDeleteRequest(req)
	}
	klog.V(2).Info("Admission wehbook for Namespace module not " +
		"configured for operations other than DELETE")
	return response
}

func (wh *webhook) validateNamespaceDeleteRequest(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	response := &admissionv1.AdmissionResponse{}
	response.Allowed = true
	svcLabel := "openebs.io/controller-service=jiva-controller-svc"

	msg := fmt.Sprintf("either BDCs or services with the label %s exists in the namespace %s.", svcLabel, req.Name)

	// ignore the Delete request of Namespace if resource name is empty
	if req.Name == "" {
		return response
	}

	bdcList, err := wh.ndmClientset.OpenebsV1alpha1().
		BlockDeviceClaims(req.Name).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("error listing BDC in namespace %s: %v", req.Name, err.Error()),
		}
		return response
	}

	if len(bdcList.Items) != 0 {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: msg,
		}
		return response
	}

	svcList, err := wh.kubeClient.CoreV1().Services(req.Name).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: svcLabel,
		})
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("error listing svc in namespace %s: %v", req.Name, err.Error()),
		}
		return response
	}

	if len(svcList.Items) != 0 {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: msg,
		}
		return response
	}

	err = wh.kubeClient.AdmissionregistrationV1().
		ValidatingWebhookConfigurations().
		Delete(context.TODO(), validatorWebhook, metav1.DeleteOptions{})
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: err.Error(),
		}
		return response
	}

	return response
}

// getCstorVolumes gets the list of CstorVolumes based in the source-volume labels
func (wh *webhook) getCstorVolumes(listOptions metav1.ListOptions) (*v1alpha1.CStorVolumeList, error) {
	return wh.clientset.OpenebsV1alpha1().CStorVolumes("").List(context.TODO(), listOptions)
}

// Serve method for webhook server, handles http requests for webhooks
func (wh *webhook) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		klog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *admissionv1.AdmissionResponse
	ar := admissionv1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		klog.Errorf("Can't decode body: %v", err)
		admissionResponse = &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		if r.URL.Path == "/validate" {
			admissionResponse = wh.validate(&ar)
		}
	}

	admissionReview := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: admissionv1.SchemeGroupVersion.String(),
			Kind:       "AdmissionReview",
		},
	}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		klog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	klog.V(5).Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		klog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func (wh *webhook) validateSPC(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request
	response := &admissionv1.AdmissionResponse{}
	response.Allowed = true
	// validates only if requested operation is DELETE
	if req.Operation == admissionv1.Delete {
		return wh.validateSPCDeleteRequest(req)
	}
	klog.V(4).Info("Admission wehbook for SPC module not "+
		"configured for operations other than DELETE requested operation %s", req.Operation)
	return response
}

// verifyExistenceOfSnapshots returns error if there are volumesnapshots or
// volumesnapshotData exists for a volume or in case any API error occurs
// otherwise return nil
func (wh *webhook) verifyExistenceOfSnapshots(pvc *corev1.PersistentVolumeClaim) error {
	volumeSnapshotList, err := wh.snapClientSet.
		VolumesnapshotV1().
		VolumeSnapshots(pvc.Namespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: snapshotMetadataPVName + "=" + pvc.Spec.VolumeName,
		})
	if err != nil {
		return errors.Wrapf(err, "failed to list snapshots related to volume: %s", pvc.Spec.VolumeName)
	}
	if len(volumeSnapshotList.Items) > 0 {
		return errors.Errorf("pvc %s has '%d' number of dependent snapshot(s)", pvc.Name, len(volumeSnapshotList.Items))
	}
	volumeSnapshotDataList, err := wh.getVolumeSnapshotDataList(pvc)
	if err != nil {
		return err
	}
	if len(volumeSnapshotDataList.Items) > 0 {
		return errors.Errorf("pvc %s has '%d' number of dependent snapshotdata(s)", pvc.Name, len(volumeSnapshotDataList.Items))
	}
	return nil
}

// getVolumeSnapshotDataList return list of volumesnapshotdata list related
// to provided persistent volume claim
func (wh *webhook) getVolumeSnapshotDataList(
	pvc *corev1.PersistentVolumeClaim) (*snapshot.VolumeSnapshotDataList, error) {
	pvcSnapshotDataList := &snapshot.VolumeSnapshotDataList{}
	snapshotDataList, err := wh.snapClientSet.
		VolumesnapshotV1().
		VolumeSnapshotDatas().
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list volumesnapshotdata related to volume: %s", pvc.Spec.VolumeName)
	}
	for _, snapshotData := range snapshotDataList.Items {
		if snapshotData.Spec.PersistentVolumeRef != nil &&
			snapshotData.Spec.PersistentVolumeRef.Name == pvc.Spec.VolumeName {
			pvcSnapshotDataList.Items = append(pvcSnapshotDataList.Items, snapshotData)
		}
	}
	return pvcSnapshotDataList, nil
}
