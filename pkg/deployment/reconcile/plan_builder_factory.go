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

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/rs/zerolog"
)

type planBuilder func(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan

type planBuilderCondition func(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) bool

type planBuilderSubPlan func(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext, w WithPlanBuilder, plans ...planBuilder) api.Plan

func NewWithPlanBuilder(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) WithPlanBuilder {
	return &withPlanBuilder{
		ctx:          ctx,
		log:          log,
		apiObject:    apiObject,
		spec:         spec,
		status:       status,
		cachedStatus: cachedStatus,
		context:      context,
	}
}

type WithPlanBuilder interface {
	Apply(p planBuilder) api.Plan
	ApplyWithCondition(c planBuilderCondition, p planBuilder) api.Plan
	ApplySubPlan(p planBuilderSubPlan, plans ...planBuilder) api.Plan
}

type withPlanBuilder struct {
	ctx          context.Context
	log          zerolog.Logger
	apiObject    k8sutil.APIObject
	spec         api.DeploymentSpec
	status       api.DeploymentStatus
	cachedStatus inspectorInterface.Inspector
	context      PlanBuilderContext
}

func (w withPlanBuilder) ApplyWithCondition(c planBuilderCondition, p planBuilder) api.Plan {
	if !c(w.ctx, w.log, w.apiObject, w.spec, w.status, w.cachedStatus, w.context) {
		return api.Plan{}
	}

	return w.Apply(p)
}

func (w withPlanBuilder) ApplySubPlan(p planBuilderSubPlan, plans ...planBuilder) api.Plan {
	return p(w.ctx, w.log, w.apiObject, w.spec, w.status, w.cachedStatus, w.context, w, plans...)
}

func (w withPlanBuilder) Apply(p planBuilder) api.Plan {
	return p(w.ctx, w.log, w.apiObject, w.spec, w.status, w.cachedStatus, w.context)
}
