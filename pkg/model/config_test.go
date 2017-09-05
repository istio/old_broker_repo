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
//	"errors"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
//	"github.com/golang/mock/gomock"

	brokerconfig "istio.io/api/broker/v1/config"
)

func TestConfigDescriptor(t *testing.T) {
	a := ProtoSchema{Type: "a", MessageName: "proxy.A"}
	descriptor := ConfigDescriptor{
		a,
		ProtoSchema{Type: "b", MessageName: "proxy.B"},
		ProtoSchema{Type: "c", MessageName: "proxy.C"},
	}
	want := []string{"a", "b", "c"}
	got := descriptor.Types()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("descriptor.Types() => got %+vwant %+v", spew.Sdump(got), spew.Sdump(want))
	}

	aType, aExists := descriptor.GetByType(a.Type)
	if !aExists || !reflect.DeepEqual(aType, a) {
		t.Errorf("descriptor.GetByType(a) => got %+v, want %+v", aType, a)
	}

	aSchema, aSchemaExists := descriptor.GetByMessageName(a.MessageName)
	if !aSchemaExists || !reflect.DeepEqual(aSchema, a) {
		t.Errorf("descriptor.GetByMessageName(a) => got %+v, want %+v", aType, a)
	}
	_, aSchemaNotExist := descriptor.GetByMessageName("blah")
	if aSchemaNotExist {
		t.Errorf("descriptor.GetByMessageName(blah) => got true, want false")
	}
}
/*
type testRegistry struct {
	ctrl     gomock.Controller
	mock     *MockConfigStore
	registry BrokerConfigStore
}

func initTestRegistry(t *testing.T) *testRegistry {
	ctrl := gomock.NewController(t)
	mock := NewMockConfigStore(ctrl)
	return &testRegistry{
		mock:     mock,
		registry: MakeIstioStore(mock),
	}
}

func (r *testRegistry) shutdown() {
	r.ctrl.Finish()
}

func TestIstioRegistryRouteRules(t *testing.T) {
	r := initTestRegistry(t)
	defer r.shutdown()

	cases := []struct {
		name      string
		mockError error
		mockObjs  []Config
		want      map[string]*proxyconfig.RouteRule
	}{
		{
			name:      "Empty object map with error",
			mockObjs:  nil,
			mockError: errors.New("foobar"),
			want:      map[string]*proxyconfig.RouteRule{},
		},
		{
			name: "Slice of unsorted RouteRules",
			mockObjs: []Config{
				{ConfigMeta: ConfigMeta{Name: "foo"}, Spec: routeRule1MatchNil},
				{ConfigMeta: ConfigMeta{Name: "bar"}, Spec: routeRule3SourceMismatch},
				{ConfigMeta: ConfigMeta{Name: "baz"}, Spec: routeRule2SourceEmpty},
			},
			want: map[string]*proxyconfig.RouteRule{
				"//foo": routeRule1MatchNil,
				"//bar": routeRule3SourceMismatch,
				"//baz": routeRule2SourceEmpty,
			},
		},
	}
	for _, c := range cases {
		r.mock.EXPECT().List(RouteRule.Type, "").Return(c.mockObjs, c.mockError)
		if got := r.registry.RouteRules(); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v with RouteRule failed: \ngot %+vwant %+v", c.name, spew.Sdump(got), spew.Sdump(c.want))
		}
	}
}

func TestIstioRegistryIngressRules(t *testing.T) {
	r := initTestRegistry(t)
	defer r.shutdown()

	rule := &proxyconfig.IngressRule{
		Name:        "sample-ingress",
		Destination: "a.svc",
	}

	r.mock.EXPECT().List(IngressRule.Type, "").Return([]Config{{
		ConfigMeta: ConfigMeta{
			Name: rule.Name,
		},
		Spec: rule,
	}}, nil)

	if got := r.registry.IngressRules(); !reflect.DeepEqual(got, map[string]*proxyconfig.IngressRule{
		"//" + rule.Name: rule,
	}) {
		t.Errorf("IngressRules failed: \ngot %+vwant %+v", spew.Sdump(got), spew.Sdump(rule))
	}

	r.mock.EXPECT().List(IngressRule.Type, "").Return(nil, errors.New("cannot list"))
	if got := r.registry.IngressRules(); len(got) > 0 {
		t.Errorf("IngressRules failed: \ngot %+vwant empty", spew.Sdump(got))
	}

}

func TestIstioRegistryRouteRulesBySource(t *testing.T) {
	r := initTestRegistry(t)
	defer r.shutdown()

	instances := []*ServiceInstance{serviceInstance1, serviceInstance2}

	mockObjs := []Config{
		{ConfigMeta: ConfigMeta{Name: "match-nil"}, Spec: routeRule1MatchNil},
		{ConfigMeta: ConfigMeta{Name: "source-empty"}, Spec: routeRule2SourceEmpty},
		{ConfigMeta: ConfigMeta{Name: "source-mismatch"}, Spec: routeRule3SourceMismatch},
		{ConfigMeta: ConfigMeta{Name: "source-match"}, Spec: routeRule4SourceMatch},
		{ConfigMeta: ConfigMeta{Name: "tag-subset-of-mismatch"}, Spec: routeRule5TagSubsetOfMismatch},
		{ConfigMeta: ConfigMeta{Name: "tag-subset-of-match"}, Spec: routeRule6TagSubsetOfMatch},
	}
	want := []*proxyconfig.RouteRule{
		routeRule6TagSubsetOfMatch,
		routeRule4SourceMatch,
		routeRule1MatchNil,
		routeRule2SourceEmpty,
	}

	r.mock.EXPECT().List(RouteRule.Type, "").Return(mockObjs, nil)
	got := r.registry.RouteRulesBySource(instances)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Failed \ngot %+vwant %+v", spew.Sdump(got), spew.Sdump(want))
	}
}

func TestIstioRegistryRouteRulesByDestination(t *testing.T) {
	r := initTestRegistry(t)
	defer r.shutdown()

	instances := []*ServiceInstance{serviceInstance2}

	mockObjs := []Config{
		{ConfigMeta: ConfigMeta{Name: "dest-foo"}, Spec: routeRule2SourceEmpty},
		{ConfigMeta: ConfigMeta{Name: "dest-two"}, Spec: routeRule7DestinationMatch},
	}
	want := []*proxyconfig.RouteRule{
		routeRule7DestinationMatch,
	}

	r.mock.EXPECT().List(RouteRule.Type, "").Return(mockObjs, nil)
	got := r.registry.RouteRulesByDestination(instances)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Failed \ngot %+vwant %+v", spew.Sdump(got), spew.Sdump(want))
	}
}

func TestIstioRegistryPolicies(t *testing.T) {
	r := initTestRegistry(t)
	defer r.shutdown()

	cases := []struct {
		name      string
		mockError error
		mockObjs  []Config
		want      []*proxyconfig.DestinationPolicy
	}{
		{
			name:      "Empty object map with error",
			mockObjs:  nil,
			mockError: errors.New("foobar"),
			want:      []*proxyconfig.DestinationPolicy{},
		},
		{
			name: "Slice of unsorted DestinationPolicy",
			mockObjs: []Config{
				{ConfigMeta: ConfigMeta{Name: "foo"}, Spec: dstPolicy1},
				{ConfigMeta: ConfigMeta{Name: "bar"}, Spec: dstPolicy2},
				{ConfigMeta: ConfigMeta{Name: "baz"}, Spec: dstPolicy3},
			},
			want: []*proxyconfig.DestinationPolicy{
				dstPolicy1, dstPolicy2, dstPolicy3,
			},
		},
	}
	makeSet := func(in []*proxyconfig.DestinationPolicy) map[*proxyconfig.DestinationPolicy]struct{} {
		out := map[*proxyconfig.DestinationPolicy]struct{}{}
		for _, c := range in {
			out[c] = struct{}{}
		}
		return out
	}

	for _, c := range cases {
		r.mock.EXPECT().List(DestinationPolicy.Type, "").Return(c.mockObjs, c.mockError)
		if got := r.registry.DestinationPolicies(); !reflect.DeepEqual(makeSet(got), makeSet(c.want)) {
			t.Errorf("%v failed: \ngot %+vwant %+v", c.name, spew.Sdump(got), spew.Sdump(c.want))
		}
	}
}

func TestIstioRegistryDestinationPolicies(t *testing.T) {
	r := initTestRegistry(t)
	defer r.shutdown()

	r.mock.EXPECT().Get(DestinationPolicy.Type, dstPolicy1.Destination, "").Return(&Config{
		Spec: dstPolicy1,
	}, true)
	want := dstPolicy1.Policy[0]
	if got := r.registry.DestinationPolicy(dstPolicy1.Destination, want.Tags); !reflect.DeepEqual(got, want) {
		t.Errorf("Failed: \ngot %+vwant %+v", spew.Sdump(got), spew.Sdump(want))
	}

	r.mock.EXPECT().Get(DestinationPolicy.Type, dstPolicy3.Destination, "").Return(nil, false)
	if got := r.registry.DestinationPolicy(dstPolicy3.Destination, nil); got != nil {
		t.Errorf("Failed: \ngot %+vwant nil", spew.Sdump(got))
	}
}

func TestEventString(t *testing.T) {
	cases := []struct {
		in   Event
		want string
	}{
		{EventAdd, "add"},
		{EventUpdate, "update"},
		{EventDelete, "delete"},
	}
	for _, c := range cases {
		if got := c.in.String(); got != c.want {
			t.Errorf("Failed: got %q want %q", got, c.want)
		}
	}
}

*/
func TestProtoSchemaConversions(t *testing.T) {
	s := &ProtoSchema{MessageName: ServiceClass.MessageName}

	msg := &brokerconfig.ServiceClass{
		Deployment: &brokerconfig.Deployment{
      Instance: "productpage",
		},
		Entry:  &brokerconfig.CatalogEntry{
			Name: "istio-bookinfo-productpage",
			Id: "4395a443-f49a-41b0-8d14-d17294cf612f",
			Description: "A book info service",
		},
	}

	wantYAML := "deployment:\n" +
    "  instance: productpage\n" +
    "entry:\n" +
		"  description: A book info service\n" +
    "  id: 4395a443-f49a-41b0-8d14-d17294cf612f\n" +
		"  name: istio-bookinfo-productpage\n"

	wantJSONMap := map[string]interface{}{
		"deployment": map[string]interface{}{
			"instance": "productpage",
		},
		"entry": map[string]interface{}{
			"name": "istio-bookinfo-productpage",
			"id": "4395a443-f49a-41b0-8d14-d17294cf612f",
			"description": "A book info service",
		},
	}

	badSchema := &ProtoSchema{MessageName: "bad-name"}
	if _, err := badSchema.FromYAML(wantYAML); err == nil {
		t.Errorf("FromYAML should have failed using ProtoSchema with bad MessageName")
	}

	gotYAML, err := s.ToYAML(msg)
	if err != nil {
		t.Errorf("ToYAML failed: %v", err)
	}
	if !reflect.DeepEqual(gotYAML, wantYAML) {
		t.Errorf("ToYAML failed: got %+v want %+v", spew.Sdump(gotYAML), spew.Sdump(wantYAML))
	}
	gotFromYAML, err := s.FromYAML(wantYAML)
	if err != nil {
		t.Errorf("FromYAML failed: %v", err)
	}
	if !reflect.DeepEqual(gotFromYAML, msg) {
		t.Errorf("FromYAML failed: got %+v want %+v", spew.Sdump(gotFromYAML), spew.Sdump(msg))
	}

	gotJSONMap, err := s.ToJSONMap(msg)
	if err != nil {
		t.Errorf("ToJSONMap failed: %v", err)
	}
	if !reflect.DeepEqual(gotJSONMap, wantJSONMap) {
		t.Errorf("ToJSONMap failed: \ngot %vwant %v", spew.Sdump(gotJSONMap), spew.Sdump(wantJSONMap))
	}
	gotFromJSONMap, err := s.FromJSONMap(wantJSONMap)
	if err != nil {
		t.Errorf("FromJSONMap failed: %v", err)
	}
	if !reflect.DeepEqual(gotFromJSONMap, msg) {
		t.Errorf("FromJSONMap failed: got %+v want %+v", spew.Sdump(gotFromJSONMap), spew.Sdump(msg))
	}
}
