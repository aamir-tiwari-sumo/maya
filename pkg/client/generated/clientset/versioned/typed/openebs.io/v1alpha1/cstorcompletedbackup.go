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
	"context"
	"time"

	v1alpha1 "github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/v1alpha1"
	scheme "github.com/aamir-tiwari-sumo/maya/pkg/client/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CStorCompletedBackupsGetter has a method to return a CStorCompletedBackupInterface.
// A group's client should implement this interface.
type CStorCompletedBackupsGetter interface {
	CStorCompletedBackups(namespace string) CStorCompletedBackupInterface
}

// CStorCompletedBackupInterface has methods to work with CStorCompletedBackup resources.
type CStorCompletedBackupInterface interface {
	Create(ctx context.Context, cStorCompletedBackup *v1alpha1.CStorCompletedBackup, opts v1.CreateOptions) (*v1alpha1.CStorCompletedBackup, error)
	Update(ctx context.Context, cStorCompletedBackup *v1alpha1.CStorCompletedBackup, opts v1.UpdateOptions) (*v1alpha1.CStorCompletedBackup, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.CStorCompletedBackup, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.CStorCompletedBackupList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.CStorCompletedBackup, err error)
	CStorCompletedBackupExpansion
}

// cStorCompletedBackups implements CStorCompletedBackupInterface
type cStorCompletedBackups struct {
	client rest.Interface
	ns     string
}

// newCStorCompletedBackups returns a CStorCompletedBackups
func newCStorCompletedBackups(c *OpenebsV1alpha1Client, namespace string) *cStorCompletedBackups {
	return &cStorCompletedBackups{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the cStorCompletedBackup, and returns the corresponding cStorCompletedBackup object, and an error if there is any.
func (c *cStorCompletedBackups) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.CStorCompletedBackup, err error) {
	result = &v1alpha1.CStorCompletedBackup{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cstorcompletedbackups").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CStorCompletedBackups that match those selectors.
func (c *cStorCompletedBackups) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.CStorCompletedBackupList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.CStorCompletedBackupList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cstorcompletedbackups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested cStorCompletedBackups.
func (c *cStorCompletedBackups) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("cstorcompletedbackups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a cStorCompletedBackup and creates it.  Returns the server's representation of the cStorCompletedBackup, and an error, if there is any.
func (c *cStorCompletedBackups) Create(ctx context.Context, cStorCompletedBackup *v1alpha1.CStorCompletedBackup, opts v1.CreateOptions) (result *v1alpha1.CStorCompletedBackup, err error) {
	result = &v1alpha1.CStorCompletedBackup{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("cstorcompletedbackups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(cStorCompletedBackup).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a cStorCompletedBackup and updates it. Returns the server's representation of the cStorCompletedBackup, and an error, if there is any.
func (c *cStorCompletedBackups) Update(ctx context.Context, cStorCompletedBackup *v1alpha1.CStorCompletedBackup, opts v1.UpdateOptions) (result *v1alpha1.CStorCompletedBackup, err error) {
	result = &v1alpha1.CStorCompletedBackup{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cstorcompletedbackups").
		Name(cStorCompletedBackup.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(cStorCompletedBackup).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the cStorCompletedBackup and deletes it. Returns an error if one occurs.
func (c *cStorCompletedBackups) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cstorcompletedbackups").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *cStorCompletedBackups) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cstorcompletedbackups").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched cStorCompletedBackup.
func (c *cStorCompletedBackups) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.CStorCompletedBackup, err error) {
	result = &v1alpha1.CStorCompletedBackup{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("cstorcompletedbackups").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
