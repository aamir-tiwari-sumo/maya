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

package fake

import (
	"context"

	v1alpha1 "github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCStorVolumes implements CStorVolumeInterface
type FakeCStorVolumes struct {
	Fake *FakeOpenebsV1alpha1
	ns   string
}

var cstorvolumesResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorvolumes"}

var cstorvolumesKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "CStorVolume"}

// Get takes name of the cStorVolume, and returns the corresponding cStorVolume object, and an error if there is any.
func (c *FakeCStorVolumes) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.CStorVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(cstorvolumesResource, c.ns, name), &v1alpha1.CStorVolume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorVolume), err
}

// List takes label and field selectors, and returns the list of CStorVolumes that match those selectors.
func (c *FakeCStorVolumes) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.CStorVolumeList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(cstorvolumesResource, cstorvolumesKind, c.ns, opts), &v1alpha1.CStorVolumeList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CStorVolumeList{ListMeta: obj.(*v1alpha1.CStorVolumeList).ListMeta}
	for _, item := range obj.(*v1alpha1.CStorVolumeList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cStorVolumes.
func (c *FakeCStorVolumes) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(cstorvolumesResource, c.ns, opts))

}

// Create takes the representation of a cStorVolume and creates it.  Returns the server's representation of the cStorVolume, and an error, if there is any.
func (c *FakeCStorVolumes) Create(ctx context.Context, cStorVolume *v1alpha1.CStorVolume, opts v1.CreateOptions) (result *v1alpha1.CStorVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(cstorvolumesResource, c.ns, cStorVolume), &v1alpha1.CStorVolume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorVolume), err
}

// Update takes the representation of a cStorVolume and updates it. Returns the server's representation of the cStorVolume, and an error, if there is any.
func (c *FakeCStorVolumes) Update(ctx context.Context, cStorVolume *v1alpha1.CStorVolume, opts v1.UpdateOptions) (result *v1alpha1.CStorVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(cstorvolumesResource, c.ns, cStorVolume), &v1alpha1.CStorVolume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorVolume), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeCStorVolumes) UpdateStatus(ctx context.Context, cStorVolume *v1alpha1.CStorVolume, opts v1.UpdateOptions) (*v1alpha1.CStorVolume, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(cstorvolumesResource, "status", c.ns, cStorVolume), &v1alpha1.CStorVolume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorVolume), err
}

// Delete takes name of the cStorVolume and deletes it. Returns an error if one occurs.
func (c *FakeCStorVolumes) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(cstorvolumesResource, c.ns, name), &v1alpha1.CStorVolume{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCStorVolumes) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(cstorvolumesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.CStorVolumeList{})
	return err
}

// Patch applies the patch and returns the patched cStorVolume.
func (c *FakeCStorVolumes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.CStorVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(cstorvolumesResource, c.ns, name, pt, data, subresources...), &v1alpha1.CStorVolume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorVolume), err
}
