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

package tmstore

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/store/storetestcases"
	"github.com/stratumn/go-core/tmpop"
	"github.com/stratumn/go-core/types"
	"github.com/stratumn/go-core/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
)

const itPrivKey = `-----BEGIN ED25519 PRIVATE KEY-----
MFACAQAwBwYDK2VwBQAEQgRAdkXy3jHVKv7T25wnQcji175T9TbxBt+KdLLwk6Tm
gncvywhyVrf3+9tkD4TOmXgq8VWi8Hn3iR4EM9o9Ua+auw==
-----END ED25519 PRIVATE KEY-----`

var (
	tmstore *TMStore
)

func newTestTMStore() (store.Adapter, error) {
	tmstore = NewTestClient()
	err := tmstore.RetryStartWebsocket(context.Background(), DefaultWsRetryInterval)
	if err != nil {
		return nil, err
	}

	return tmstore, nil
}

func resetTMPop(_ store.Adapter) {
	ResetNode()
}

func updateValidatorRulesFile(t *testing.T, in, out string) {
	rulesInputFile, err := os.OpenFile(in, os.O_RDONLY, 0666)
	defer rulesInputFile.Close()
	content, err := ioutil.ReadAll(rulesInputFile)
	require.NoErrorf(t, err, "Cannot read validator rule file %s", in)
	err = ioutil.WriteFile(out, content, 0666)
	require.NoErrorf(t, err, "Cannot write validator rule file %s", out)

	if testTmpop != nil {
		tmInfo := testTmpop.Info(abci.RequestInfo{})
		testTmpop.BeginBlock(abci.RequestBeginBlock{
			Hash: tmInfo.LastBlockAppHash,
			Header: abci.Header{
				AppHash: tmInfo.LastBlockAppHash,
				Height:  tmInfo.LastBlockHeight,
			},
		})
		time.Sleep(500 * time.Millisecond)
	}
}

func TestTMStore(t *testing.T) {
	rulesFilename := filepath.Join("testdata", "rules.test.json")
	updateValidatorRulesFile(t, "/dev/null", rulesFilename)
	testConfig := &tmpop.Config{Validation: &validation.Config{RulesPath: rulesFilename}}
	node := StartNode(testConfig)
	defer node.Wait()
	defer node.Stop()
	defer os.Remove(rulesFilename)

	t.Run("Store test cases", func(t *testing.T) {
		storetestcases.Factory{
			New:  newTestTMStore,
			Free: resetTMPop,
		}.RunStoreTests(t)
	})

	t.Run("Validation", func(t *testing.T) {
		tmstore.StartWebsocket(context.Background())
		updateValidatorRulesFile(t, filepath.Join("testdata", "rules.json"), rulesFilename)
		data := map[string]interface{}{"string": "test"}

		var err error
		t.Run("Validation succeeds", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess("testProcess").
				WithoutParent().
				WithStep("init").
				WithData(t, data).
				WithSignatureFromKey(t, []byte(itPrivKey), "").
				Build()

			_, err = tmstore.CreateLink(context.Background(), l)
			assert.NoError(t, err, "CreateLink() failed")
		})

		t.Run("Schema validation failed", func(t *testing.T) {
			badData := map[string]interface{}{"string": 42}
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess("testProcess").
				WithStep("init").
				WithoutParent().
				WithData(t, badData).
				Build()

			_, err = tmstore.CreateLink(context.Background(), l)
			assert.Error(t, err, "A validation error is expected")

			structErr, ok := err.(*types.Error)
			assert.True(t, ok, "Invalid error received: want types.Error")
			assert.Equal(t, errorcode.InvalidArgument, structErr.Code)
		})

		t.Run("Signature validation failed", func(t *testing.T) {
			// here we sign the link before modifying the state, making the signature out-of-date
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess("testProcess").
				WithStep("init").
				WithoutParent().
				WithSignature(t, "").
				WithData(t, data).
				Build()

			_, err = tmstore.CreateLink(context.Background(), l)
			assert.Error(t, err, "A validation error is expected")

			structErr, ok := err.(*types.Error)
			assert.True(t, ok, "Invalid error received: want types.Error")
			assert.Equal(t, errorcode.InvalidArgument, structErr.Code)
		})

		t.Run("Validation rules update succeeds", func(t *testing.T) {
			prevLink := chainscripttest.NewLinkBuilder(t).
				WithProcess("testProcess").
				WithStep("init").
				WithoutParent().
				WithData(t, data).
				Build()

			_, err = tmstore.CreateLink(context.Background(), prevLink)
			require.NoError(t, err, "CreateLink(init) failed")

			l := chainscripttest.NewLinkBuilder(t).
				Branch(t, prevLink).
				WithStep("processing").
				WithData(t, data).
				WithSignatureFromKey(t, []byte(itPrivKey), "").
				Build()

			_, err = tmstore.CreateLink(context.Background(), l)
			assert.NoError(t, err, "CreateLink() failed")

			updateValidatorRulesFile(t, filepath.Join("testdata", "rules.new.json"), rulesFilename)

			l = chainscripttest.NewLinkBuilder(t).
				WithProcess("testProcess").
				WithStep("processing").
				WithData(t, data).
				WithSignatureFromKey(t, []byte(itPrivKey), "").
				Build()

			_, err = tmstore.CreateLink(context.Background(), l)
			assert.Error(t, err, "CreateLink() should failed because signature is missing")
		})
	})

	// TestWebSocket tests how the web socket with Tendermint behaves
	t.Run("Websocket", func(t *testing.T) {
		t.Run("Start and stop websocket", func(t *testing.T) {
			err := tmstore.StartWebsocket(context.Background())
			assert.NoError(t, err)

			err = tmstore.StopWebsocket(context.Background())
			assert.NoError(t, err)
		})

		t.Run("Start websocket multiple times", func(t *testing.T) {
			err := tmstore.StartWebsocket(context.Background())
			assert.NoError(t, err)

			err = tmstore.StartWebsocket(context.Background())
			assert.NoError(t, err)

			err = tmstore.StopWebsocket(context.Background())
			assert.NoError(t, err)
		})

		t.Run("Stop already stopped websocket", func(t *testing.T) {
			err := tmstore.StopWebsocket(context.Background())
			assert.EqualError(t, err, "subscription not found")
		})
	})
}
