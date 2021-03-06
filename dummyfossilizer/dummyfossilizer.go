// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

// Package dummyfossilizer implements a fossilizer that can be used for testing.
//
// It doesn't do much -- it just adds a timestamp.
package dummyfossilizer

import (
	"context"
	"sync"
	"time"

	"github.com/stratumn/go-core/dummyfossilizer/evidences"
	"github.com/stratumn/go-core/fossilizer"
)

const (
	// Description is the description set in the fossilizer's information.
	Description = "Stratumn's Dummy Fossilizer"
)

// Config contains configuration options for the store.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// DummyFossilizer is the type that implements
// github.com/stratumn/go-core/fossilizer.Adapter.
type DummyFossilizer struct {
	config *Config

	chansLock            sync.RWMutex
	fossilizerEventChans []chan *fossilizer.Event
}

// New creates an instance of a DummyFossilizer.
func New(config *Config) *DummyFossilizer {
	return &DummyFossilizer{config: config, fossilizerEventChans: nil}
}

// GetInfo implements github.com/stratumn/go-core/fossilizer.Adapter.GetInfo.
func (a *DummyFossilizer) GetInfo(ctx context.Context) (interface{}, error) {
	return &Info{
		Name:        evidences.Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// AddFossilizerEventChan implements
// github.com/stratumn/go-core/fossilizer.Adapter.AddFossilizerEventChan.
func (a *DummyFossilizer) AddFossilizerEventChan(fossilizerEventChan chan *fossilizer.Event) {
	a.chansLock.Lock()
	defer a.chansLock.Unlock()

	a.fossilizerEventChans = append(a.fossilizerEventChans, fossilizerEventChan)
}

// ListenersCount returns the number of registered listeners.
func (a *DummyFossilizer) ListenersCount() int {
	a.chansLock.RLock()
	defer a.chansLock.RUnlock()

	return len(a.fossilizerEventChans)
}

// Fossilize implements github.com/stratumn/go-core/fossilizer.Adapter.Fossilize.
func (a *DummyFossilizer) Fossilize(ctx context.Context, data []byte, meta []byte) error {
	proof := &evidences.DummyProof{
		Timestamp: uint64(time.Now().Unix()),
	}
	evidence, err := proof.Evidence(evidences.Name)
	if err != nil {
		return err
	}

	r := &fossilizer.Result{
		Evidence: *evidence,
		Fossil: fossilizer.Fossil{
			Data: data,
			Meta: meta,
		},
	}
	event := &fossilizer.Event{
		EventType: fossilizer.DidFossilize,
		Data:      r,
	}

	a.chansLock.RLock()
	defer a.chansLock.RUnlock()

	for _, c := range a.fossilizerEventChans {
		go func(c chan *fossilizer.Event) {
			c <- event
		}(c)
	}

	return nil
}
