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
// Author Tomasz Mielech
//

package reconcile

import (
	"context"
	"strconv"

	"github.com/arangodb/go-driver"
	"github.com/rs/zerolog"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

// createCleanOutPlan creates clean out action if the server is cleaned out and the operator is not aware of it.
func createCleanOutPlan(ctx context.Context, log zerolog.Logger, _ k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, _ inspectorInterface.Inspector, planCtx PlanBuilderContext) api.Plan {

	if spec.GetMode() != api.DeploymentModeCluster {
		return nil
	}

	if !status.Conditions.IsTrue(api.ConditionTypeUpToDate) {
		// Do not consider to mark cleanedout servers when changes are propagating
		return nil
	}

	cluster, err := getCluster(ctx, planCtx)
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to get cluster")
		return nil
	}

	ctxChild, cancel := context.WithTimeout(ctx, arangod.GetRequestTimeout())
	defer cancel()
	health, err := cluster.Health(ctxChild)
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to get cluster health")
		return nil
	}

	var plan api.Plan

	for id, member := range health.Health {
		switch member.Role {
		case driver.ServerRoleDBServer:
			memberStatus, ok := status.Members.DBServers.ElementByID(string(id))
			if !ok {
				continue
			}

			if memberStatus.Conditions.IsTrue(api.ConditionTypeCleanedOut) {
				continue
			}

			if isCleanedOut, err := cluster.IsCleanedOut(ctx, string(id)); err != nil {
				log.Warn().Err(err).Str("id", string(id)).Msgf("Unable to get clean out status")
				return nil
			} else if isCleanedOut {
				log.Info().
					Str("role", string(member.Role)).
					Str("id", string(id)).
					Msgf("server is cleaned out so operator must do the same")

				action := api.NewAction(api.ActionTypeSetMemberCondition, api.ServerGroupDBServers, string(id),
					"server is cleaned out so operator must do the same").
					AddParam(string(api.ConditionTypeCleanedOut), strconv.FormatBool(true))

				plan = append(plan, action)
			}
		}
	}

	return plan
}
