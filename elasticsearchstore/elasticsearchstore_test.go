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

package elasticsearchstore

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/store/storetestcases"
	"github.com/stratumn/go-core/testutil"
	"github.com/stratumn/go-core/tmpop/tmpoptestcases"
	"github.com/stratumn/go-core/types"
	"github.com/stratumn/go-core/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	test *testing.T
)

const (
	domain = "0.0.0.0"
	port   = "9200"
)

func TestMain(m *testing.M) {
	flag.Parse()
	// ElasticSearch container configuration.
	imageName := "docker.elastic.co/elasticsearch/elasticsearch:6.2.1"
	containerName := "sdk_elasticsearchstore_test"
	p, _ := nat.NewPort("tcp", port)
	exposedPorts := map[nat.Port]struct{}{p: {}}
	portBindings := nat.PortMap{
		p: []nat.PortBinding{
			{
				HostIP:   domain,
				HostPort: port,
			},
		},
	}

	// Stop container if it is already running, swallow error.
	testutil.KillContainer(containerName)

	// Start elasticsearch container.
	env := []string{"discovery.type=single-node"}
	if err := testutil.RunContainerWithEnv(containerName, imageName, env, exposedPorts, portBindings); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	// Retry until container is ready.
	if err := util.Retry(func(attempt int) (bool, error) {
		_, err := http.Get(fmt.Sprintf("http://%s:%s", domain, port))
		if err != nil {
			time.Sleep(1 * time.Second)
			return true, err
		}
		return false, err
	}, 60); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	// Run tests.
	testResult := m.Run()

	// Stop elasticsearch container.
	if err := testutil.KillContainer(containerName); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	os.Exit(testResult)
}

func TestElasticSearchStore(t *testing.T) {
	test = t
	factory := storetestcases.Factory{
		New:               newTestElasticSearchStoreAdapter,
		NewKeyValueStore:  newTestElasticSearchStoreKeyValue,
		Free:              freeTestElasticSearchStoreAdapter,
		FreeKeyValueStore: freeTestElasticSearchStoreKeyValue,
	}

	factory.RunStoreTests(t)
	factory.RunKeyValueStoreTests(t)
}

func TestElasticSearchTMPop(t *testing.T) {
	tmpoptestcases.Factory{
		New:  newTestElasticSearchStoreTMPop,
		Free: freeTestElasticSearchStoreTMPop,
	}.RunTests(t)
}

func verifyResultsCount(t *testing.T, err error, segments *types.PaginatedSegments, expectedCount int) {
	require.NoError(t, err)
	require.NotNil(t, segments)
	assert.Len(t, segments.Segments, expectedCount, "Invalid number of results")
	assert.Condition(t, func() bool { return len(segments.Segments) <= segments.TotalCount }, "Invalid total count of results")
}

func TestElasticSearchStoreSearch(t *testing.T) {
	a, err := newTestElasticSearchStore()
	require.NoError(t, err, "newTestElasticSearchStore()")
	require.NotNil(t, a, "ES adapter")
	defer freeTestElasticSearchStore(a)

	data1 := map[string]interface{}{"nested": map[string]interface{}{
		"first":  "hector",
		"last":   "salazar",
		"common": "stratumn",
	}}
	link1 := chainscripttest.NewLinkBuilder(t).
		WithProcess("something crazy").
		WithTags("one", "two", "three").
		WithMapID("foo bar").
		WithData(t, data1).
		Build()
	a.CreateLink(context.Background(), link1)
	hash1, _ := link1.Hash()

	data2 := map[string]interface{}{"nested": map[string]interface{}{
		"first":  "james",
		"last":   "daniel",
		"common": "stratumn",
	}}
	link2 := chainscripttest.NewLinkBuilder(t).
		WithProcess("fly emirates").
		WithTags("urban", "paranoia", "city").
		WithMapID("stupid madness").
		WithData(t, data2).
		Build()
	a.CreateLink(context.Background(), link2)
	hash2, _ := link2.Hash()

	t.Run("Simple Search Query", func(t *testing.T) {
		t.Run("Should find segment based on partial state match", func(t *testing.T) {
			slice, err := a.SimpleSearchQuery(
				context.Background(),
				&SearchQuery{
					SegmentFilter: store.SegmentFilter{
						Pagination: store.Pagination{
							Limit: 5,
						},
					},
					Query: "sala*",
				})
			verifyResultsCount(t, err, slice, 1)
			assert.Equal(t, hash1, slice.Segments[0].LinkHash(), "Wrong link was found")
		})

		t.Run("Should find segment based on mapId", func(t *testing.T) {
			slice, err := a.SimpleSearchQuery(
				context.Background(),
				&SearchQuery{
					SegmentFilter: store.SegmentFilter{
						Pagination: store.Pagination{
							Limit: 5,
						},
					},
					Query: "emirates",
				})
			verifyResultsCount(t, err, slice, 1)
			assert.Equal(t, hash2, slice.Segments[0].LinkHash(), "Wrong link was found")
		})

		t.Run("Should filter on Process", func(t *testing.T) {
			slice, err := a.SimpleSearchQuery(
				context.Background(),
				&SearchQuery{
					SegmentFilter: store.SegmentFilter{
						Pagination: store.Pagination{
							Limit: 5,
						},
						Process: "fly emirates",
					},
					Query: "stratu*",
				})
			verifyResultsCount(t, err, slice, 1)
			assert.Equal(t, hash2, slice.Segments[0].LinkHash(), "Wrong link was found")
		})

		t.Run("Should filter on one MapId", func(t *testing.T) {
			slice, err := a.SimpleSearchQuery(
				context.Background(),
				&SearchQuery{
					SegmentFilter: store.SegmentFilter{
						Pagination: store.Pagination{
							Limit: 5,
						},
						MapIDs: []string{"foo bar"},
					},
					Query: "stratu*",
				})
			verifyResultsCount(t, err, slice, 1)
			assert.Equal(t, hash1, slice.Segments[0].LinkHash(), "Wrong link was found")
		})

		t.Run("Should filter on multiple MapIds", func(t *testing.T) {
			slice, err := a.SimpleSearchQuery(
				context.Background(),
				&SearchQuery{
					SegmentFilter: store.SegmentFilter{
						Pagination: store.Pagination{
							Limit: 5,
						},
						MapIDs: []string{"foo bar", "stupid madness"},
					},
					Query: "stratu*",
				})
			verifyResultsCount(t, err, slice, 2)
			h1, h2 := slice.Segments[0].LinkHash(), slice.Segments[1].LinkHash()
			assert.Contains(t, []chainscript.LinkHash{hash1, hash2}, h1, "Wrong link was found")
			assert.Contains(t, []chainscript.LinkHash{hash1, hash2}, h2, "Wrong link was found")
			assert.NotEqual(t, h1, h2, "The two results are the same")
		})
	})

	t.Run("Multi Match Query", func(t *testing.T) {
		t.Run("Should find segments based on multiple words", func(t *testing.T) {
			slice, err := a.MultiMatchQuery(
				context.Background(),
				&SearchQuery{
					SegmentFilter: store.SegmentFilter{
						Pagination: store.Pagination{
							Limit: 5,
						},
					},
					Query: "salazar daniel",
				})
			verifyResultsCount(t, err, slice, 2)
			h1, h2 := slice.Segments[0].LinkHash(), slice.Segments[1].LinkHash()
			assert.Contains(t, []chainscript.LinkHash{hash1, hash2}, h1, "Wrong link was found")
			assert.Contains(t, []chainscript.LinkHash{hash1, hash2}, h2, "Wrong link was found")
			assert.NotEqual(t, h1, h2, "The two results are the same")
		})
	})

	t.Run("Should extract all value tokens", func(t *testing.T) {
		l := chainscripttest.RandomLink(t)
		data := []byte(`{
			"string": "hello",
			"bool": true,
			"num": 0.54,
			"array": ["abc", 1, true, {"name": "james"}, ["def", 1, true]],
			"object": {
				"string": "world",
				"bool": false,
				"num": 23
			}
		}`)

		dataObj := map[string]interface{}{}
		err := json.Unmarshal(data, &dataObj)
		require.NoError(t, err)

		err = l.SetData(dataObj)
		require.NoError(t, err)

		expectedTokens := []string{"hello", "abc", "james", "def", "world"}

		doc, err := fromLink(l)
		assert.NoError(t, err, "fromLink")
		require.NotNil(t, doc, "fromLink")
		assert.Equal(t, len(expectedTokens), len(doc.DataTokens), "Invalid number of tokens")
		for _, token := range expectedTokens {
			assert.Contains(t, doc.DataTokens, token, "Invalid tokens extracted")
		}
	})
}

func newTestElasticSearchStore() (*ESStore, error) {
	config := &Config{
		URL: fmt.Sprintf("http://%s:%s", domain, port),
	}
	return New(config)
}

func newTestElasticSearchStoreAdapter() (store.Adapter, error) {
	return newTestElasticSearchStore()
}

func newTestElasticSearchStoreKeyValue() (store.KeyValueStore, error) {
	return newTestElasticSearchStore()
}

func newTestElasticSearchStoreTMPop() (store.Adapter, store.KeyValueStore, error) {
	a, err := newTestElasticSearchStore()
	return a, a, err
}

func freeTestElasticSearchStore(a *ESStore) {
	if err := a.deleteIndex(linksIndex); err != nil {
		test.Fatal(err)
	}
	if err := a.deleteIndex(evidencesIndex); err != nil {
		test.Fatal(err)
	}
	if err := a.deleteIndex(valuesIndex); err != nil {
		test.Fatal(err)
	}
}

func freeTestElasticSearchStoreAdapter(a store.Adapter) {
	freeTestElasticSearchStore(a.(*ESStore))
}

func freeTestElasticSearchStoreKeyValue(a store.KeyValueStore) {
	freeTestElasticSearchStore(a.(*ESStore))
}

func freeTestElasticSearchStoreTMPop(a store.Adapter, _ store.KeyValueStore) {
	freeTestElasticSearchStoreAdapter(a)
}
