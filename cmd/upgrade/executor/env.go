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

package executor

import (
	menv "github.com/aamir-tiwari-sumo/maya/pkg/env/v1alpha1"
)

//This file defines the environement variable names that are specific
// to this provisioner. In addition to the variables defined in this file,
// provisioner also uses the following:
//   OPENEBS_NAMESPACE

func getOpenEBSNamespace() string {
	return menv.Get(menv.OpenEBSNamespace)
}

func getUpgradeTaskLabel() string {
	return menv.Get("UPGRADE_TASK_LABEL")
}
