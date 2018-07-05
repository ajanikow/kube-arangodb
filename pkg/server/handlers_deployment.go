//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

// Deployment is the API implemented by an ArangoDeployment.
type Deployment interface {
	Name() string
	Namespace() string
	Mode() api.DeploymentMode
	PodCount() int
	ReadyPodCount() int
	VolumeCount() int
	ReadyVolumeCount() int
	StorageClasses() []string
	DatabaseURL() string
	DatabaseVersion() (string, string)
}

// DeploymentOperator is the API implemented by the deployment operator.
type DeploymentOperator interface {
	GetDeployments() ([]Deployment, error)
}

// DeploymentInfo is the information returned per deployment.
type DeploymentInfo struct {
	Name             string             `json:"name"`
	Namespace        string             `json:"namespace"`
	Mode             api.DeploymentMode `json:"mode"`
	PodCount         int                `json:"pod_count"`
	ReadyPodCount    int                `json:"ready_pod_count"`
	VolumeCount      int                `json:"volume_count"`
	ReadyVolumeCount int                `json:"ready_volume_count"`
	StorageClasses   []string           `json:"storage_classes"`
	DatabaseURL      string             `json:"database_url"`
	DatabaseVersion  string             `json:"database_version"`
	DatabaseLicense  string             `json:"database_license"`
}

// Handle a GET /api/deployment request
func (s *Server) handleGetDeployments(c *gin.Context) {
	if do := s.deps.Operators.DeploymentOperator(); do != nil {
		// Fetch deployments
		depls, err := do.GetDeployments()
		if err != nil {
			sendError(c, err)
		} else {
			result := make([]DeploymentInfo, len(depls))
			for i, d := range depls {
				version, license := d.DatabaseVersion()
				result[i] = DeploymentInfo{
					Name:             d.Name(),
					Namespace:        d.Namespace(),
					Mode:             d.Mode(),
					PodCount:         d.PodCount(),
					ReadyPodCount:    d.ReadyPodCount(),
					VolumeCount:      d.VolumeCount(),
					ReadyVolumeCount: d.ReadyVolumeCount(),
					StorageClasses:   d.StorageClasses(),
					DatabaseURL:      d.DatabaseURL(),
					DatabaseVersion:  version,
					DatabaseLicense:  license,
				}
			}
			c.JSON(http.StatusOK, gin.H{
				"deployments": result,
			})
		}
	}
}
