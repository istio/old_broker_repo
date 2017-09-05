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

package model

import (
	"errors"
	"fmt"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"

	brokerconfig "istio.io/api/broker/v1/config"
	"istio.io/broker/pkg/model/test"
)

// ConfigMeta is metadata attached to each configuration unit.
// The revision is optional, and if provided, identifies the
// last update operation on the object.
type ConfigMeta struct {
	// Type is a short configuration name that matches the content message type
	// (e.g. "route-rule")
	Type string `json:"type,omitempty"`

	// Name is a unique immutable identifier in a namespace
	Name string `json:"name,omitempty"`

	// Namespace defines the space for names (optional for some types),
	// applications may choose to use namespaces for a variety of purposes
	// (security domains, fault domains, organizational domains)
	Namespace string `json:"namespace,omitempty"`

	// Namespace where istio control plane is installed
	IstioNamespace string `json:"istioNamespace,omitempty"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	Annotations map[string]string `json:"annotations,omitempty"`

	// ResourceVersion is an opaque identifier for tracking updates to the config registry.
	// The implementation may use a change index or a commit log for the revision.
	// The config client should not make any assumptions about revisions and rely only on
	// exact equality to implement optimistic concurrency of read-write operations.
	//
	// The lifetime of an object of a particular revision depends on the underlying data store.
	// The data store may compactify old revisions in the interest of storage optimization.
	//
	// An empty revision carries a special meaning that the associated object has
	// not been stored and assigned a revision.
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// Config is a configuration unit consisting of the type of configuration, the
// key identifier that is unique per type, and the content represented as a
// protobuf message.
type Config struct {
	ConfigMeta

	// Spec holds the configuration object as a protobuf message
	Spec proto.Message
}

// ConfigStore describes a set of platform agnostic APIs that must be supported
// by the underlying platform to store and retrieve Istio configuration.
//
// Configuration key is defined to be a combination of the type, name, and
// namespace of the configuration object. The configuration key is guaranteed
// to be unique in the store.
//
// The storage interface presented here assumes that the underlying storage
// layer supports _Get_ (list), _Update_ (update), _Create_ (create) and
// _Delete_ semantics but does not guarantee any transactional semantics.
//
// _Update_, _Create_, and _Delete_ are mutator operations. These operations
// are asynchronous, and you might not see the effect immediately (e.g. _Get_
// might not return the object by key immediately after you mutate the store.)
// Intermittent errors might occur even though the operation succeeds, so you
// should always check if the object store has been modified even if the
// mutating operation returns an error.  Objects should be created with
// _Create_ operation and updated with _Update_ operation.
//
// Resource versions record the last mutation operation on each object. If a
// mutation is applied to a different revision of an object than what the
// underlying storage expects as defined by pure equality, the operation is
// blocked.  The client of this interface should not make assumptions about the
// structure or ordering of the revision identifier.
//
// Object references supplied and returned from this interface should be
// treated as read-only. Modifying them violates thread-safety.
type ConfigStore interface {
	// ConfigDescriptor exposes the configuration type schema known by the config store.
	// The type schema defines the bidrectional mapping between configuration
	// types and the protobuf encoding schema.
	ConfigDescriptor() ConfigDescriptor

	// Get retrieves a configuration element by a type and a key
	Get(typ, name, namespace string) (config *Config, exists bool)

	// List returns objects by type and namespace.
	// Use "" for the namespace to list across namespaces.
	List(typ, namespace string) ([]Config, error)

	// Create adds a new configuration object to the store. If an object with the
	// same name and namespace for the type already exists, the operation fails
	// with no side effects.
	Create(config Config) (revision string, err error)

	// Update modifies an existing configuration object in the store.  Update
	// requires that the object has been created.  Resource version prevents
	// overriding a value that has been changed between prior _Get_ and _Put_
	// operation to achieve optimistic concurrency. This method returns a new
	// revision if the operation succeeds.
	Update(config Config) (newRevision string, err error)

	// Delete removes an object from the store by key
	Delete(typ, name, namespace string) error
}

// Key function for the configuration objects
func Key(typ, name, namespace string) string {
	return fmt.Sprintf("%s/%s/%s", typ, namespace, name)
}

// Key is the unique identifier for a configuration object
func (config *Config) Key() string {
	return Key(config.Type, config.Name, config.Namespace)
}

// ConfigDescriptor defines the bijection between the short type name and its
// fully qualified protobuf message name
type ConfigDescriptor []ProtoSchema

// Types lists all known types in the config schema
func (descriptor ConfigDescriptor) Types() []string {
	types := make([]string, 0, len(descriptor))
	for _, t := range descriptor {
		types = append(types, t.Type)
	}
	return types
}

// GetByMessageName finds a schema by message name if it is available
func (descriptor ConfigDescriptor) GetByMessageName(name string) (ProtoSchema, bool) {
	for _, schema := range descriptor {
		if schema.MessageName == name {
			return schema, true
		}
	}
	return ProtoSchema{}, false
}

// GetByType finds a schema by type if it is available
func (descriptor ConfigDescriptor) GetByType(name string) (ProtoSchema, bool) {
	for _, schema := range descriptor {
		if schema.Type == name {
			return schema, true
		}
	}
	return ProtoSchema{}, false
}

// BrokerConfigStore is a specialized interface to access config store using
// Broker configuration types.
type BrokerConfigStore interface {
	// ServiceClasses lists all service classes.
	ServiceClasses() map[string]*brokerconfig.ServiceClass

	// ServicePlans lists all service plans.
	ServicePlans() map[string]*brokerconfig.ServicePlan
}

const (
	// IstioAPIGroup defines API group name for Istio configuration resources
	IstioAPIGroup = "config.istio.io"

	// IstioAPIVersion defines API group version
	IstioAPIVersion = "v1alpha2"
)

var (
	// MockConfig is used purely for testing
	MockConfig = ProtoSchema{
		Type:        "mock-config",
		Plural:      "mock-configs",
		MessageName: "test.MockConfig",
		Validate: func(config proto.Message) error {
			if config.(*test.MockConfig).Key == "" {
				return errors.New("empty key")
			}
			return nil
		},
	}

	// ServiceClass describes service class
	ServiceClass = ProtoSchema{
		Type:        "service-class",
		Plural:      "service-classes",
		MessageName: "istio.broker.v1.config.ServiceClass",
		Validate:    nil,
	}

	// ServicePlan describes service plan
	ServicePlan = ProtoSchema{
		Type:        "service-plan",
		Plural:      "service-plans",
		MessageName: "istio.broker.v1.config.ServicePlan",
		Validate:    nil,
	}
)

// brokerConfigStore provides a simple adapter for Broker configuration types
// from the generic config registry
type brokerConfigStore struct {
	ConfigStore
}

// MakeBrokerStore creates a wrapper around a store
func MakeIstioStore(store ConfigStore) BrokerConfigStore {
	return &brokerConfigStore{store}
}

func (i brokerConfigStore) ServiceClasses() map[string]*brokerconfig.ServiceClass {
	out := make(map[string]*brokerconfig.ServiceClass)
	rs, err := i.List(ServiceClass.Type, "")
	if err != nil {
		glog.V(2).Infof("ServiceClasses => %v", err)
	}
	for _, r := range rs {
		if c, ok := r.Spec.(*brokerconfig.ServiceClass); ok {
			out[r.Key()] = c
		}
	}
	return out
}

func (i brokerConfigStore) ServicePlans() map[string]*brokerconfig.ServicePlan {
	out := make(map[string]*brokerconfig.ServicePlan)
	rs, err := i.List(ServicePlan.Type, "")
	if err != nil {
		glog.V(2).Infof("ServicePlans => %v", err)
	}
	for _, r := range rs {
		if c, ok := r.Spec.(*brokerconfig.ServicePlan); ok {
			out[r.Key()] = c
		}
	}
	return out
}
