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
	"errors"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"istio.io/mixer/bazel-mixer/external/com_github_docker_distribution/registry/client"
	"istio.io/mixer/bazel-mixer/external/com_github_prometheus_common/model"
)

type Controller struct {
}

func CreateController(cloudName string, instanceMap map[string]*model.ServiceInstance, bindingMap map[string]*model.ServiceBinding) (*Controller, error) {
	return new(Controller), nil
}

func (c *Controller) Catalog(w http.ResponseWriter, r *http.Request) {
	glog.Infof("Get Service Broker Catalog...")
	var catalog model.Catalog
	catalogFileName := "dummy.catalog.json"

	if c.cloudName == utils.AWS {
		catalogFileName = "catalog.AWS.json"
	} else if c.cloudName == utils.SOFTLAYER || c.cloudName == utils.SL {
		catalogFileName = "catalog.SoftLayer.json"
	}

	err := utils.ReadAndUnmarshal(&catalog, conf.CatalogPath, catalogFileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	utils.WriteResponse(w, http.StatusOK, catalog)
}

func readJson(object interface{}, dir string, fileName string) error {
	path := dir + string(os.PathSeparator) + fileName

	bytes, err := ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, object)
	if err != nil {
		return err
	}

	return nil
}
