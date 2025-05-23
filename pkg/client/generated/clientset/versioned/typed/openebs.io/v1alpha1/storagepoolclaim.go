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

// StoragePoolClaimsGetter has a method to return a StoragePoolClaimInterface.
// A group's client should implement this interface.
type StoragePoolClaimsGetter interface {
	StoragePoolClaims() StoragePoolClaimInterface
}

// StoragePoolClaimInterface has methods to work with StoragePoolClaim resources.
type StoragePoolClaimInterface interface {
	Create(ctx context.Context, storagePoolClaim *v1alpha1.StoragePoolClaim, opts v1.CreateOptions) (*v1alpha1.StoragePoolClaim, error)
	Update(ctx context.Context, storagePoolClaim *v1alpha1.StoragePoolClaim, opts v1.UpdateOptions) (*v1alpha1.StoragePoolClaim, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.StoragePoolClaim, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.StoragePoolClaimList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.StoragePoolClaim, err error)
	StoragePoolClaimExpansion
}

// storagePoolClaims implements StoragePoolClaimInterface
type storagePoolClaims struct {
	client rest.Interface
}

// newStoragePoolClaims returns a StoragePoolClaims
func newStoragePoolClaims(c *OpenebsV1alpha1Client) *storagePoolClaims {
	return &storagePoolClaims{
		client: c.RESTClient(),
	}
}

// Get takes name of the storagePoolClaim, and returns the corresponding storagePoolClaim object, and an error if there is any.
func (c *storagePoolClaims) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.StoragePoolClaim, err error) {
	result = &v1alpha1.StoragePoolClaim{}
	err = c.client.Get().
		Resource("storagepoolclaims").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of StoragePoolClaims that match those selectors.
func (c *storagePoolClaims) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.StoragePoolClaimList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.StoragePoolClaimList{}
	err = c.client.Get().
		Resource("storagepoolclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested storagePoolClaims.
func (c *storagePoolClaims) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("storagepoolclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a storagePoolClaim and creates it.  Returns the server's representation of the storagePoolClaim, and an error, if there is any.
func (c *storagePoolClaims) Create(ctx context.Context, storagePoolClaim *v1alpha1.StoragePoolClaim, opts v1.CreateOptions) (result *v1alpha1.StoragePoolClaim, err error) {
	result = &v1alpha1.StoragePoolClaim{}
	err = c.client.Post().
		Resource("storagepoolclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(storagePoolClaim).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a storagePoolClaim and updates it. Returns the server's representation of the storagePoolClaim, and an error, if there is any.
func (c *storagePoolClaims) Update(ctx context.Context, storagePoolClaim *v1alpha1.StoragePoolClaim, opts v1.UpdateOptions) (result *v1alpha1.StoragePoolClaim, err error) {
	result = &v1alpha1.StoragePoolClaim{}
	err = c.client.Put().
		Resource("storagepoolclaims").
		Name(storagePoolClaim.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(storagePoolClaim).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the storagePoolClaim and deletes it. Returns an error if one occurs.
func (c *storagePoolClaims) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("storagepoolclaims").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *storagePoolClaims) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("storagepoolclaims").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched storagePoolClaim.
func (c *storagePoolClaims) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.StoragePoolClaim, err error) {
	result = &v1alpha1.StoragePoolClaim{}
	err = c.client.Patch(pt).
		Resource("storagepoolclaims").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
