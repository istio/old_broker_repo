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

package crd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	multierror "github.com/hashicorp/go-multierror"
	yaml2 "gopkg.in/yaml.v2"
)

// Descriptor defines a group of config types.
type Descriptor struct {
  schemas []Schema
}

// Types lists all known types in the config schema
func (d Descriptor) Types() []string {
	ts := make([]string, 0, len(d.schemas))
	for _, t := range d.schemas {
		ts = append(ts, t.Type)
	}
	return ts
}

// GetByMessageName finds a schema by message name if it is available
func (d Descriptor) GetByMessageName(name string) (Schema, bool) {
	for _, s := range d.schemas {
		if s.MessageName == name {
			return s, true
		}
	}
	return Schema{}, false
}

// GetByType finds a schema by type if it is available
func (d Descriptor) GetByType(name string) (Schema, bool) {
	for _, s := range d.schemas {
		if s.Type == name {
			return s, true
		}
	}
	return Schema{}, false
}

// JSONConfig is the JSON serialized form of the config unit
type JSONConfig struct {
	Meta

	// Spec is the content of the config
	Spec interface{} `json:"spec,omitempty"`
}

// FromJSON deserializes and validates a JSON config object
func (d Descriptor) FromJSON(config JSONConfig) (*Entry, error) {
	s, ok := d.GetByType(config.Type)
	if !ok {
		return nil, fmt.Errorf("unknown spec type %s", config.Type)
	}

	m, err := s.FromJSONMap(config.Spec)
	if err != nil {
		return nil, fmt.Errorf("cannot parse proto message: %v", err)
	}

	if err = s.Validate(m); err != nil {
		return nil, err
	}
	return &Entry{
		Meta: config.Meta,
		Spec: m,
	}, nil
}

// FromYAML deserializes and validates a YAML config object
func (d Descriptor) FromYAML(content []byte) (*Entry, error) {
	out := JSONConfig{}
	err := yaml.Unmarshal(content, &out)
	if err != nil {
		return nil, err
	}
	return d.FromJSON(out)
}

// ToYAML serializes a config into a YAML form
func (d Descriptor) ToYAML(config Entry) (string, error) {
	s, ok := d.GetByType(config.Type)
	if !ok {
		return "", fmt.Errorf("missing type %q", config.Type)
	}

	spec, err := s.ToJSONMap(config.Spec)
	if err != nil {
		return "", err
	}

	out := JSONConfig{
		Meta: config.Meta,
		Spec: spec,
	}

	bytes, err := yaml.Marshal(out)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
