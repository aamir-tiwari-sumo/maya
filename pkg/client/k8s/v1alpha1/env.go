/*
Copyright 2018 The OpenEBS Authors

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
	// K8sMasterIPEnvironmentKey is the environment variable key used to
	// determine the kubernetes master IP address
	K8sMasterIPEnvironmentKey menv.ENVKey = "OPENEBS_IO_K8S_MASTER"
	// KubeConfigEnvironmentKey is the environment variable key used to
	// determine the kubernetes config
	KubeConfigEnvironmentKey menv.ENVKey = "OPENEBS_IO_KUBE_CONFIG"
)
