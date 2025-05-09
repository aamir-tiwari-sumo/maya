/*
Copyright 2018-2019 The OpenEBS Authors

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

package v1alpha1

import (
	menv "github.com/aamir-tiwari-sumo/maya/pkg/env/v1alpha1"
)

const (
	// DefaultCstorSparsePool is the environment variable that
	// flags if default cstor pool should be configured or not
	//
	// If value is "true", default cstor pool will be
	// installed/configured else for "false" it will
	// not be configured
	DefaultCstorSparsePool menv.ENVKey = "OPENEBS_IO_INSTALL_DEFAULT_CSTOR_SPARSE_POOL"

	// CreateDefaultStorageConfig is the environment
	// variable that flags if default storage pools and/or storage
	// classes should be created.
	//
	// Default is "true"
	CreateDefaultStorageConfig menv.ENVKey = "OPENEBS_IO_CREATE_DEFAULT_STORAGE_CONFIG"

	// InstallCRD is the environment
	// variable that flags if maya apiserver should install the CRDs
	// As the installation moves towards helm 3, the responsibility of installing
	// CRDs can be pushed to helm.
	//
	// Default is "true"
	InstallCRD menv.ENVKey = "OPENEBS_IO_INSTALL_CRD"
)
