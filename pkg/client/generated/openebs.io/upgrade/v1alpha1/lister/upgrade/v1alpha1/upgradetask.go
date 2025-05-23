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
	v1alpha1 "github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// UpgradeTaskLister helps list UpgradeTasks.
// All objects returned here must be treated as read-only.
type UpgradeTaskLister interface {
	// List lists all UpgradeTasks in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.UpgradeTask, err error)
	// UpgradeTasks returns an object that can list and get UpgradeTasks.
	UpgradeTasks(namespace string) UpgradeTaskNamespaceLister
	UpgradeTaskListerExpansion
}

// upgradeTaskLister implements the UpgradeTaskLister interface.
type upgradeTaskLister struct {
	indexer cache.Indexer
}

// NewUpgradeTaskLister returns a new UpgradeTaskLister.
func NewUpgradeTaskLister(indexer cache.Indexer) UpgradeTaskLister {
	return &upgradeTaskLister{indexer: indexer}
}

// List lists all UpgradeTasks in the indexer.
func (s *upgradeTaskLister) List(selector labels.Selector) (ret []*v1alpha1.UpgradeTask, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.UpgradeTask))
	})
	return ret, err
}

// UpgradeTasks returns an object that can list and get UpgradeTasks.
func (s *upgradeTaskLister) UpgradeTasks(namespace string) UpgradeTaskNamespaceLister {
	return upgradeTaskNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// UpgradeTaskNamespaceLister helps list and get UpgradeTasks.
// All objects returned here must be treated as read-only.
type UpgradeTaskNamespaceLister interface {
	// List lists all UpgradeTasks in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.UpgradeTask, err error)
	// Get retrieves the UpgradeTask from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.UpgradeTask, error)
	UpgradeTaskNamespaceListerExpansion
}

// upgradeTaskNamespaceLister implements the UpgradeTaskNamespaceLister
// interface.
type upgradeTaskNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all UpgradeTasks in the indexer for a given namespace.
func (s upgradeTaskNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.UpgradeTask, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.UpgradeTask))
	})
	return ret, err
}

// Get retrieves the UpgradeTask from the indexer for a given namespace and name.
func (s upgradeTaskNamespaceLister) Get(name string) (*v1alpha1.UpgradeTask, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("upgradetask"), name)
	}
	return obj.(*v1alpha1.UpgradeTask), nil
}
