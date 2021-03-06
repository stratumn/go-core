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

package postgresstore_test

import (
	"testing"

	"github.com/stratumn/go-core/postgresstore"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/store/storetestcases"
	"github.com/stratumn/go-core/tmpop/tmpoptestcases"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	factory := storetestcases.Factory{
		New:               createAdapter,
		NewKeyValueStore:  createKeyValueStore,
		Free:              freeAdapter,
		FreeKeyValueStore: freeKeyValueStore,
	}

	factory.RunStoreTests(t)
	factory.RunKeyValueStoreTests(t)
}

func TestPostgresTMPop(t *testing.T) {
	tmpoptestcases.Factory{
		New:  createAdapterTMPop,
		Free: freeAdapterTMPop,
	}.RunTests(t)
}

func createStore() (*postgresstore.Store, error) {
	a, err := postgresstore.New(&postgresstore.Config{
		URL: "postgres://postgres@localhost:5433/sdk_test?sslmode=disable",
	})
	if err := a.Create(); err != nil {
		return nil, err
	}
	if err := a.Prepare(); err != nil {
		return nil, err
	}
	return a, err
}

func createAdapter() (store.Adapter, error) {
	return createStore()
}

func createKeyValueStore() (store.KeyValueStore, error) {
	return createStore()
}

func freeStore(s *postgresstore.Store) {
	if err := s.Drop(); err != nil {
		panic(err)
	}
	if err := s.Close(); err != nil {
		panic(err)
	}
}

func freeAdapter(s store.Adapter) {
	freeStore(s.(*postgresstore.Store))
}

func freeKeyValueStore(s store.KeyValueStore) {
	freeStore(s.(*postgresstore.Store))
}

func createAdapterTMPop() (store.Adapter, store.KeyValueStore, error) {
	a, err := createStore()
	return a, a, err
}

func freeAdapterTMPop(a store.Adapter, _ store.KeyValueStore) {
	freeAdapter(a)
}

func TestCreatePrepare(t *testing.T) {
	// If create and prepare have already been called, we should not fail.
	_, err := createStore()
	require.NoError(t, err)

	_, err = createStore()
	require.NoError(t, err)
}
