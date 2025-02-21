//
// DISCLAIMER
//
// Copyright 2021 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Adam Janikowski
//

package arangomember

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Inspector interface {
	ArangoMember(name string) (*api.ArangoMember, bool)
	IterateArangoMembers(action ArangoMemberAction, filters ...ArangoMemberFilter) error
}

type ArangoMemberFilter func(pod *api.ArangoMember) bool
type ArangoMemberAction func(pod *api.ArangoMember) error

func FilterByDeploymentUID(uid types.UID) ArangoMemberFilter {
	return func(pod *api.ArangoMember) bool {
		return pod.Spec.DeploymentUID == "" || pod.Spec.DeploymentUID == uid
	}
}
