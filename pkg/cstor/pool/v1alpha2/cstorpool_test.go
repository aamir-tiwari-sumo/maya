// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha2

import (
	"reflect"
	"testing"

	apis "github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func mockAlwaysTrue(*CSP) bool  { return true }
func mockAlwaysFalse(*CSP) bool { return false }

func TestCStorPoolAll(t *testing.T) {
	tests := map[string]struct {
		Predicates     predicateList
		expectedOutput bool
	}{
		// Positive predicates
		"Positive Predicate 1": {[]predicate{mockAlwaysTrue}, true},
		"Positive Predicate 2": {[]predicate{mockAlwaysTrue, mockAlwaysTrue}, true},
		"Positive Predicate 3": {[]predicate{mockAlwaysTrue, mockAlwaysTrue, mockAlwaysTrue}, true},
		// Negative Predicates
		"Negative Predicate 1": {[]predicate{mockAlwaysFalse}, false},
		"Negative Predicate 2": {[]predicate{mockAlwaysTrue, mockAlwaysFalse}, false},
		"Negative Predicate 3": {[]predicate{mockAlwaysFalse, mockAlwaysTrue}, false},
		"Negative Predicate 4": {[]predicate{mockAlwaysFalse, mockAlwaysFalse}, false},
		"Negative Predicate 5": {[]predicate{mockAlwaysFalse, mockAlwaysTrue, mockAlwaysTrue}, false},
		"Negative Predicate 6": {[]predicate{mockAlwaysTrue, mockAlwaysFalse, mockAlwaysTrue}, false},
		"Negative Predicate 7": {[]predicate{mockAlwaysTrue, mockAlwaysTrue, mockAlwaysFalse}, false},
		"Negative Predicate 8": {[]predicate{mockAlwaysTrue, mockAlwaysFalse, mockAlwaysFalse}, false},
		"Negative Predicate 9": {[]predicate{mockAlwaysFalse, mockAlwaysFalse, mockAlwaysFalse}, false},
	}
	for name, mock := range tests {
		// pin it
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			if output := mock.Predicates.all(&CSP{}); output != mock.expectedOutput {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output)
			}
		})
	}
}

func TestCStorPoolIsNotUID(t *testing.T) {
	tests := map[string]struct {
		CSPuid         types.UID
		uids           []string
		expectedOutput bool
	}{
		// Positive Test
		"Positive 1": {"uid6", []string{"uid1", "uid2", "uid3", "uid4"}, true},
		"Positive 2": {"uid7", []string{"uid1", "uid2", "uid3", "uid4"}, true},
		"Positive 3": {"uid8", []string{"uid1", "uid2", "uid3", "uid4"}, true},
		"Positive 4": {"uid9", []string{"uid1", "uid2", "uid3", "uid4"}, true},
		"Positive 5": {"uid10", []string{"uid1", "uid2", "uid3", "uid4"}, true},

		// Negative Test
		"Negative 1": {"uid1", []string{"uid1", "uid2", "uid3", "uid4", "uid5"}, false},
		"Negative 2": {"uid2", []string{"uid1", "uid2", "uid3", "uid4", "uid5"}, false},
		"Negative 3": {"uid3", []string{"uid1", "uid2", "uid3", "uid4", "uid5"}, false},
		"Negative 4": {"uid4", []string{"uid1", "uid2", "uid3", "uid4", "uid5"}, false},
		"Negative 5": {"uid5", []string{"uid1", "uid2", "uid3", "uid4", "uid5"}, false},
	}
	for name, mock := range tests {
		// pin it
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			mockCSP := &CSP{&apis.CStorPool{ObjectMeta: metav1.ObjectMeta{UID: mock.CSPuid}}}
			p := IsNotUID(mock.uids...)
			if output := p(mockCSP); output != mock.expectedOutput {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output)
			}
		})
	}
}

func TestCStorPoolFilterUIDs(t *testing.T) {
	tests := map[string]struct {
		Predicates     predicateList
		UIDs           []string
		expectedOutput []string
	}{
		// With all Positive predicates
		"Positive 1": {[]predicate{mockAlwaysTrue}, []string{"uid1", "uid2", "uid3"}, []string{"uid1", "uid2", "uid3"}},
		"Positive 2": {[]predicate{mockAlwaysTrue, mockAlwaysTrue}, []string{"uid1", "uid2", "uid3"}, []string{"uid1", "uid2", "uid3"}},
		"Positive 3": {[]predicate{mockAlwaysTrue, mockAlwaysTrue}, []string{"uid1", "uid2"}, []string{"uid1", "uid2"}},
		//  With all negative predicates
		"Negative 1": {[]predicate{mockAlwaysFalse}, []string{"uid1", "uid2", "uid3"}, []string{}},
		"Negative 2": {[]predicate{mockAlwaysFalse, mockAlwaysFalse}, []string{"uid1", "uid2", "uid3"}, []string{}},
		"Negative 3": {[]predicate{mockAlwaysFalse, mockAlwaysFalse}, []string{"uid1", "uid2", "uid3"}, []string{}},
	}
	for name, mock := range tests {
		// pin it
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			CSPL := ListBuilder().WithUIDs(mock.UIDs...).List()
			output := CSPL.Filter(mock.Predicates...)
			if len(mock.expectedOutput) != len(output.GetPoolUIDs()) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output.GetPoolUIDs())
			}
			for index, val := range output.GetPoolUIDs() {
				if val != mock.expectedOutput[index] {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output.GetPoolUIDs())
				}
			}
		})
	}
}

func TestCStorPoolWithUIDs(t *testing.T) {
	tests := map[string]struct {
		expectedUIDs []string
	}{
		"UID set 1":  {[]string{}},
		"UID set 2":  {[]string{"uid1"}},
		"UID set 3":  {[]string{"uid1", "uid2"}},
		"UID set 4":  {[]string{"uid1", "uid2", "uid3"}},
		"UID set 5":  {[]string{"uid1", "uid2", "uid3", "uid4"}},
		"UID set 6":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5"}},
		"UID set 7":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6"}},
		"UID set 8":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7"}},
		"UID set 9":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8"}},
		"UID set 10": {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8", "uid9"}},
	}

	for name, mock := range tests {
		// pin it
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			lb := ListBuilder().WithUIDs(mock.expectedUIDs...)
			if len(lb.list.Items) != len(mock.expectedUIDs) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs, lb.list.Items)
			}
			for index, val := range lb.list.Items {
				if string(val.Object.GetUID()) != mock.expectedUIDs[index] {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs[index], string(val.Object.GetUID()))
				}
			}
		})
	}
}

func TestCstorPoolList(t *testing.T) {
	tests := map[string]struct {
		expectedUIDs []string
	}{
		"UID set 1":  {[]string{}},
		"UID set 2":  {[]string{"uid1"}},
		"UID set 3":  {[]string{"uid1", "uid2"}},
		"UID set 4":  {[]string{"uid1", "uid2", "uid3"}},
		"UID set 5":  {[]string{"uid1", "uid2", "uid3", "uid4"}},
		"UID set 6":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5"}},
		"UID set 7":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6"}},
		"UID set 8":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7"}},
		"UID set 9":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8"}},
		"UID set 10": {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8", "uid9"}},
	}

	for name, mock := range tests {
		// pin it
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			lb := ListBuilder().WithUIDs(mock.expectedUIDs...).List()
			if len(lb.Items) != len(mock.expectedUIDs) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs, lb.Items)
			}
			for index, val := range lb.Items {
				if string(val.Object.GetUID()) != mock.expectedUIDs[index] {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs[index], string(val.Object.GetUID()))
				}
			}
		})
	}
}

func TestBuildWithListUids(t *testing.T) {
	tests := map[string]struct {
		expectedUIDs []string
	}{
		"UID set 1":  {[]string{}},
		"UID set 2":  {[]string{"uid1"}},
		"UID set 3":  {[]string{"uid1", "uid2"}},
		"UID set 4":  {[]string{"uid1", "uid2", "uid3"}},
		"UID set 5":  {[]string{"uid1", "uid2", "uid3", "uid4"}},
		"UID set 6":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5"}},
		"UID set 7":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6"}},
		"UID set 8":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7"}},
		"UID set 9":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8"}},
		"UID set 10": {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8", "uid9"}},
	}

	for name, mock := range tests {
		// pin it
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			lb := ListBuilder().WithUIDs(mock.expectedUIDs...).List()
			if len(lb.GetPoolUIDs()) != len(mock.expectedUIDs) {
				t.Fatalf("Test %v failed, Expected %v Got %v", name, lb.GetPoolUIDs(), mock.expectedUIDs)
			}

		})
	}
}

func TestNewListFromUIDNode(t *testing.T) {
	tests := map[string]struct {
		UIDNodeMap     map[string]string
		UIDCapacityMap map[string]string
		expectedPools  []string
	}{
		"Test 1": {map[string]string{"Pool 1": "host 1"}, map[string]string{"Pool 1": "9.40G"}, []string{"Pool 1"}},
		"Test 2": {map[string]string{"Pool 1": "host 1", "Pool 2": "host 2"}, map[string]string{"Pool 1": "9.40G"}, []string{"Pool 1", "Pool 2"}},
		"Test 3": {map[string]string{"Pool 1": "host 1", "Pool 2": "host 2", "Pool 3": "host 3"}, map[string]string{"Pool 1": "9.40G"}, []string{"Pool 1", "Pool 2", "Pool 3"}},
		"Test 4": {map[string]string{"Pool 1": "host 1", "Pool 2": "host 2", "Pool 3": "host 3", "Pool 4": "host 4"}, map[string]string{"Pool 1": "9.40G"}, []string{"Pool 1", "Pool 2", "Pool 3", "Pool 4"}},
		"Test 5": {map[string]string{"Pool 1": "host 1", "Pool 2": "host 2", "Pool 3": "host 3", "Pool 4": "host 4", "Pool 5": "host 5"}, map[string]string{"Pool 1": "9.40G"}, []string{"Pool 1", "Pool 2", "Pool 3", "Pool 4", "Pool 5"}},
	}

	for name, mock := range tests {
		// pin it
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			output := newListFromUIDNode(mock.UIDNodeMap, mock.UIDCapacityMap).GetPoolUIDs()
			if len(output) != len(mock.expectedPools) {
				t.Fatalf("Test %v failed: Expected %v but got %v", name, mock.expectedPools, output)
			}

		})
	}
}

func TestNewListFromUIDs(t *testing.T) {
	tests := map[string]struct {
		PoolUIDs []string
	}{
		"Test 1": {[]string{"Pool 1"}},
		"Test 2": {[]string{"Pool 1", "Pool 2"}},
		"Test 3": {[]string{"Pool 1", "Pool 2", "Pool 3"}},
		"Test 4": {[]string{"Pool 1", "Pool 2", "Pool 3", "Pool 4"}},
		"Test 5": {[]string{"Pool 1", "Pool 2", "Pool 3", "Pool 4", "Pool 5"}},
	}

	for name, mock := range tests {
		// pin it
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			output := newListFromUIDs(mock.PoolUIDs).GetPoolUIDs()
			if len(output) != len(mock.PoolUIDs) {
				t.Fatalf("Test %v failed: Expected %v but got %v", name, mock.PoolUIDs, output)
			}

		})
	}
}

func TestTemplateFunctionsCount(t *testing.T) {
	tests := map[string]struct {
		expectedLength int
	}{
		"Test 1": {2},
	}

	for name, test := range tests {
		// pin it
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			p := TemplateFunctions()
			if len(p) != test.expectedLength {
				t.Fatalf("test %q failed: expected Items %v but got %v", name, test.expectedLength, len(p))
			}
		})
	}
}

func TestHasAnnotation(t *testing.T) {
	tests := map[string]struct {
		availableAnnotations       map[string]string
		checkForKey, checkForValue string
		hasAnnotation              bool
	}{
		"Test 1": {map[string]string{"Anno 1": "Val 1"}, "Anno 1", "Val 1", true},
		"Test 2": {map[string]string{"Anno 1": "Val 1"}, "Anno 1", "Val 2", false},
		"Test 3": {map[string]string{"Anno 1": "Val 1", "Anno 2": "Val 2"}, "Anno 0", "Val 2", false},
		"Test 4": {map[string]string{"Anno 1": "Val 1", "Anno 2": "Val 2"}, "Anno 1", "Val 1", true},
		"Test 5": {map[string]string{"Anno 1": "Val 1", "Anno 2": "Val 2", "Anno 3": "Val 3"}, "Anno 1", "Val 1", true},
	}

	for name, test := range tests {
		// pin it
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakeCSP := &CSP{&apis.CStorPool{ObjectMeta: metav1.ObjectMeta{Annotations: test.availableAnnotations}}}
			ok := HasAnnotation(test.checkForKey, test.checkForValue)(fakeCSP)
			if ok != test.hasAnnotation {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.availableAnnotations, fakeCSP.Object.GetAnnotations())
			}
		})
	}
}

func TestWithAPIList(t *testing.T) {
	tests := map[string]struct {
		expectedPoolIDs []string
	}{
		"Test 1": {[]string{"Pool 1"}},
		"Test 2": {[]string{"Pool 1", "Pool 2"}},
		"Test 3": {[]string{"Pool 1", "Pool 2", "Pool 3"}},
		"Test 4": {[]string{"Pool 1", "Pool 2", "Pool 3", "Pool 4"}},
		"Test 5": {[]string{"Pool 1", "Pool 2", "Pool 3", "Pool 4", "Pool 5"}},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			poolItems := []apis.CStorPool{}
			for _, p := range mock.expectedPoolIDs {
				poolItems = append(poolItems, apis.CStorPool{ObjectMeta: metav1.ObjectMeta{Name: p}})
			}

			b := ListBuilder().WithAPIList(&apis.CStorPoolList{Items: poolItems})
			for index, ob := range b.list.Items {
				if !reflect.DeepEqual(*ob.Object, poolItems[index]) {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, poolItems[index], ob.Object)
				}
			}
		})
	}

}
