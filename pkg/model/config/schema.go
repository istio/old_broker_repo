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

package config

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

const (
	dns1123LabelMaxLength int    = 63
	dns1123LabelFmt       string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
)

var (
	dns1123LabelRex = regexp.MustCompile("^" + dns1123LabelFmt + "$")
)

// Schema provides description of the configuration schema and its key function
type Schema struct {
	// Type refers to the short configuration type name
	Type string

	// Plural refers to the short plural configuration name
	Plural string

	// MessageName refers to the protobuf message type name corresponding to the type
	MessageName string

	// AdditionalValidate the protobuf message for this type. This is called within schema.Validate()
	// This can be nil.
	AdditionalValidate func(config proto.Message) error
}

// Make creates a new instance of the proto message
func (ps *Schema) Make() (proto.Message, error) {
	pbt := proto.MessageType(ps.MessageName)
	if pbt == nil {
		return nil, fmt.Errorf("unknown type %q", ps.MessageName)
	}
	return reflect.New(pbt.Elem()).Interface().(proto.Message), nil
}

// ToJSON marshals a proto to canonical JSON
func (ps *Schema) ToJSON(msg proto.Message) (string, error) {
	// Marshal from proto to json bytes
	m := jsonpb.Marshaler{}
	out, err := m.MarshalToString(msg)
	if err != nil {
		return "", err
	}
	return out, nil
}

// ToYAML marshals a proto to canonical YAML
func (ps *Schema) ToYAML(msg proto.Message) (string, error) {
	js, err := ps.ToJSON(msg)
	if err != nil {
		return "", err
	}
	yml, err := yaml.JSONToYAML([]byte(js))
	return string(yml), err
}

// ToJSONMap converts a proto message to a generic map using canonical JSON encoding
// JSON encoding is specified here: https://developers.google.com/protocol-buffers/docs/proto3#json
func (ps *Schema) ToJSONMap(msg proto.Message) (map[string]interface{}, error) {
	js, err := ps.ToJSON(msg)
	if err != nil {
		return nil, err
	}

	// Unmarshal from json bytes to go map
	var data map[string]interface{}
	err = json.Unmarshal([]byte(js), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// FromJSON converts a canonical JSON to a proto message
func (ps *Schema) FromJSON(js string) (proto.Message, error) {
	pb, err := ps.Make()
	if err != nil {
		return nil, err
	}
	if err = ApplyJSON(js, pb); err != nil {
		return nil, err
	}
	return pb, nil
}

// ApplyJSON unmarshals a JSON string into a proto message
func ApplyJSON(js string, pb proto.Message) error {
	return jsonpb.UnmarshalString(js, pb)
}

// FromYAML converts a canonical YAML to a proto message
func (ps *Schema) FromYAML(yml string) (proto.Message, error) {
	pb, err := ps.Make()
	if err != nil {
		return nil, err
	}
	if err = ApplyYAML(yml, pb); err != nil {
		return nil, err
	}
	return pb, nil
}

// ApplyYAML unmarshals a YAML string into a proto message
func ApplyYAML(yml string, pb proto.Message) error {
	js, err := yaml.YAMLToJSON([]byte(yml))
	if err != nil {
		return err
	}
	return ApplyJSON(string(js), pb)
}

// FromJSONMap converts from a generic map to a proto message using canonical JSON encoding
// JSON encoding is specified here: https://developers.google.com/protocol-buffers/docs/proto3#json
func (ps *Schema) FromJSONMap(data interface{}) (proto.Message, error) {
	// Marshal to YAML bytes
	str, err := yaml2.Marshal(data)
	if err != nil {
		return nil, err
	}
	out, err := ps.FromYAML(string(str))
	if err != nil {
		return nil, multierror.Prefix(err, fmt.Sprintf("YAML decoding error: %v", string(str)))
	}
	return out, nil
}

// isDNS1123Label tests for a string that conforms to the definition of a label in
// DNS (RFC 1123).
func isDNS1123Label(value string) bool {
	return len(value) <= dns1123LabelMaxLength && dns1123LabelRex.MatchString(value)
}

// Validate the basic config. Invokes AdditionalValidate() if set.
func (ps *Schema) Validate(config proto.Message) error {
	if !isDNS1123Label(ps.Type) {
		return fmt.Errorf("invalid type: %q", ps.Type)
	}
	if !isDNS1123Label(ps.Plural) {
		return fmt.Errorf("invalid plural: %q", ps.Plural)
	}
	if proto.MessageType(ps.MessageName) == nil {
		return fmt.Errorf("cannot discover proto message type: %q", ps.MessageName)
	}
	if ps.AdditionalValidate != nil {
		return ps.AdditionalValidate(config)
	}
	return nil
}

type JSONConfig struct {
	Meta

	// Spec is the content of the config
	Spec interface{} `json:"spec,omitempty"`
}

// Descriptor defines a group of config types.
type Descriptor interface {
	// Types lists all known types in the config schema
	Types() []string

	// GetByMessageName finds a schema by message name if it is available
	GetByMessageName(name string) (Schema, bool)

	// GetByType finds a schema by type if it is available
	GetByType(name string) (Schema, bool)

	// FromJSON deserializes and validates a JSON config object
	FromJSON(config JSONConfig) (*Entry, error)

	// FromYAML deserializes and validates a YAML config object
	FromYAML(content []byte) (*Entry, error)

	// ToYAML serializes a config into a YAML form
	ToYAML(config Entry) (string, error)
}
