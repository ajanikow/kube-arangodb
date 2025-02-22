//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
)

func withMaintenance(plan ...api.Action) api.Plan {
	if !features.Maintenance().Enabled() {
		return plan
	}

	return withMaintenanceStart(plan...).After(api.NewAction(api.ActionTypeDisableMaintenance, api.ServerGroupUnknown, "", "Disable maintenance after actions"))
}
func withMaintenanceStart(plan ...api.Action) api.Plan {
	if !features.Maintenance().Enabled() {
		return plan
	}

	return api.AsPlan(plan).Before(
		api.NewAction(api.ActionTypeEnableMaintenance, api.ServerGroupUnknown, "", "Enable maintenance before actions"),
		api.NewAction(api.ActionTypeSetMaintenanceCondition, api.ServerGroupUnknown, "", "Enable maintenance before actions"))
}

func withResignLeadership(group api.ServerGroup, member api.MemberStatus, reason string, plan ...api.Action) api.Plan {
	if member.Image == nil {
		return plan
	}

	return api.AsPlan(plan).Before(api.NewAction(api.ActionTypeResignLeadership, group, member.ID, reason))
}

func cleanOutMember(group api.ServerGroup, m api.MemberStatus) api.Plan {
	var plan api.Plan

	if group == api.ServerGroupDBServers {
		plan = append(plan,
			api.NewAction(api.ActionTypeCleanOutMember, group, m.ID),
		)
	}
	plan = append(plan,
		api.NewAction(api.ActionTypeShutdownMember, group, m.ID),
		api.NewAction(api.ActionTypeRemoveMember, group, m.ID),
	)

	return plan
}
