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

	"github.com/rs/zerolog"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func init() {
	registerAction(api.ActionTypeSetMemberCondition, setMemberCondition)
}

func setMemberCondition(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionSetMemberCondition{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type actionSetMemberCondition struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

// Start starts the action for changing conditions on the provided member.
func (a actionSetMemberCondition) Start(ctx context.Context) (bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Info().Msg("can not set the condition because the member is gone already")
		return true, nil
	}

	if len(a.action.Params) == 0 {
		a.log.Info().Msg("can not start the action with the empty list of conditions")
		return true, nil
	}

	for condition, value := range a.action.Params {
		if value == "" {
			a.log.Debug().Msg("remove the condition")

			m.Conditions.Remove(api.ConditionType(condition))
		} else {
			set, err := strconv.ParseBool(value)
			if err != nil {
				a.log.Error().Err(err).Str("value", value).Msg("can not parse string to boolean")
				continue
			}

			a.log.Debug().Msg("set the condition")

			m.Conditions.Update(api.ConditionType(condition), set, a.action.Reason, "action set the member condition")
		}
	}

	if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
		return false, errors.Wrap(errors.WithStack(err), "can not update the member")
	}

	return true, nil
}
