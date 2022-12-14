/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package state

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/dapr/components-contrib/state"
	"github.com/dapr/dapr/pkg/components"
)

type State struct {
	Name          string
	FactoryMethod func() state.Store
}

func New(name string, factoryMethod func() state.Store) State {
	return State{
		Name:          name,
		FactoryMethod: factoryMethod,
	}
}

// Registry is an interface for a component that returns registered state store implementations.
type Registry interface {
	Register(components ...State)
	Create(name, version string) (state.Store, error)
}

type stateStoreRegistry struct {
	stateStores map[string]func() state.Store
}

// NewRegistry is used to create state store registry.
func NewRegistry() Registry {
	return &stateStoreRegistry{
		stateStores: map[string]func() state.Store{},
	}
}

// // Register registers a new factory method that creates an instance of a StateStore.
// // The key is the name of the state store, eg. redis.
func (s *stateStoreRegistry) Register(components ...State) {
	for _, component := range components {
		s.stateStores[createFullName(component.Name)] = component.FactoryMethod
	}
}

func (s *stateStoreRegistry) Create(name, version string) (state.Store, error) {
	if method, ok := s.getStateStore(name, version); ok {
		return method(), nil
	}
	return nil, errors.Errorf("couldn't find state store %s/%s", name, version)
}

func (s *stateStoreRegistry) getStateStore(name, version string) (func() state.Store, bool) {
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	stateStoreFn, ok := s.stateStores[nameLower+"/"+versionLower]
	if ok {
		return stateStoreFn, true
	}
	if components.IsInitialVersion(versionLower) {
		stateStoreFn, ok = s.stateStores[nameLower]
	}
	return stateStoreFn, ok
}

func createFullName(name string) string {
	return strings.ToLower("state." + name)
}
