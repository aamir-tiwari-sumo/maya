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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CStorPoolLister helps list CStorPools.
// All objects returned here must be treated as read-only.
type CStorPoolLister interface {
	// List lists all CStorPools in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.CStorPool, err error)
	// Get retrieves the CStorPool from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.CStorPool, error)
	CStorPoolListerExpansion
}

// cStorPoolLister implements the CStorPoolLister interface.
type cStorPoolLister struct {
	indexer cache.Indexer
}

// NewCStorPoolLister returns a new CStorPoolLister.
func NewCStorPoolLister(indexer cache.Indexer) CStorPoolLister {
	return &cStorPoolLister{indexer: indexer}
}

// List lists all CStorPools in the indexer.
func (s *cStorPoolLister) List(selector labels.Selector) (ret []*v1alpha1.CStorPool, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CStorPool))
	})
	return ret, err
}

// Get retrieves the CStorPool from the index for a given name.
func (s *cStorPoolLister) Get(name string) (*v1alpha1.CStorPool, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("cstorpool"), name)
	}
	return obj.(*v1alpha1.CStorPool), nil
}
