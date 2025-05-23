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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/aamir-tiwari-sumo/maya/pkg/client/generated/clientset/versioned/scheme"
	rest "k8s.io/client-go/rest"
)

type OpenebsV1alpha1Interface interface {
	RESTClient() rest.Interface
	CASTemplatesGetter
	CStorBackupsGetter
	CStorCompletedBackupsGetter
	CStorPoolsGetter
	CStorRestoresGetter
	CStorVolumesGetter
	CStorVolumeReplicasGetter
	RunTasksGetter
	StoragePoolsGetter
	StoragePoolClaimsGetter
}

// OpenebsV1alpha1Client is used to interact with features provided by the openebs.io group.
type OpenebsV1alpha1Client struct {
	restClient rest.Interface
}

func (c *OpenebsV1alpha1Client) CASTemplates() CASTemplateInterface {
	return newCASTemplates(c)
}

func (c *OpenebsV1alpha1Client) CStorBackups(namespace string) CStorBackupInterface {
	return newCStorBackups(c, namespace)
}

func (c *OpenebsV1alpha1Client) CStorCompletedBackups(namespace string) CStorCompletedBackupInterface {
	return newCStorCompletedBackups(c, namespace)
}

func (c *OpenebsV1alpha1Client) CStorPools() CStorPoolInterface {
	return newCStorPools(c)
}

func (c *OpenebsV1alpha1Client) CStorRestores(namespace string) CStorRestoreInterface {
	return newCStorRestores(c, namespace)
}

func (c *OpenebsV1alpha1Client) CStorVolumes(namespace string) CStorVolumeInterface {
	return newCStorVolumes(c, namespace)
}

func (c *OpenebsV1alpha1Client) CStorVolumeReplicas(namespace string) CStorVolumeReplicaInterface {
	return newCStorVolumeReplicas(c, namespace)
}

func (c *OpenebsV1alpha1Client) RunTasks(namespace string) RunTaskInterface {
	return newRunTasks(c, namespace)
}

func (c *OpenebsV1alpha1Client) StoragePools() StoragePoolInterface {
	return newStoragePools(c)
}

func (c *OpenebsV1alpha1Client) StoragePoolClaims() StoragePoolClaimInterface {
	return newStoragePoolClaims(c)
}

// NewForConfig creates a new OpenebsV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*OpenebsV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &OpenebsV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new OpenebsV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *OpenebsV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new OpenebsV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *OpenebsV1alpha1Client {
	return &OpenebsV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *OpenebsV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
