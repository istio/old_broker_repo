// Copyright 2017 Istio Authors
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
package controller

import (
	"net/http"

	"github.com/golang/glog"
	"istio.io/broker/pkg/model"
	"istio.io/broker/pkg/utils"
)

const (
	demoCatalogFilePath = "example"
	catalogFileName     = "demo_catalog.json"
)

type Controller struct {
}

func CreateController() (*Controller, error) {
	return new(Controller), nil
}

func (c *Controller) Catalog(w http.ResponseWriter, r *http.Request) {
	glog.Infof("Get Service Broker Catalog...")
	var catalog model.Catalog

	err := utils.ReadAndUnmarshal(&catalog, demoCatalogFilePath, catalogFileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	utils.WriteResponse(w, http.StatusOK, catalog)
}
