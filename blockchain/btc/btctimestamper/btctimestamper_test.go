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

package btctimestamper

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/blockchain/btc/btctesting"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetwork_NetworkTest3(t *testing.T) {
	t.Run("Handles uncompressed keys", func(t *testing.T) {
		ts, err := New(&Config{
			WIF: "91cH7CBcQDEhtfu12qPd9ZzeD2aGtKRBE5raJvig81GntJwfJ7R",
			Fee: int64(10000),
		})
		require.NoError(t, err)
		assert.Equal(t, btc.NetworkTest3.String(), ts.Network().String())
		assert.Equal(t, ts.address.EncodeAddress(), "mfcFnCqjbRafqPA3XHMfoHjHe1TCmHfjWZ")
	})

	t.Run("Handles compressed keys", func(t *testing.T) {
		ts, err := New(&Config{
			WIF: "cTqzoRGya8Dw5aBxcojLTtbSKnMo1LiAjS25JNjjjxCuopSHekmV",
			Fee: int64(10000),
		})
		require.NoError(t, err)
		assert.Equal(t, btc.NetworkTest3.String(), ts.Network().String())
		assert.Equal(t, ts.address.EncodeAddress(), "mxxP8z8DXp4Gzb5XAcTRFQnAphVZ13Qti8")
	})
}

func TestNetwork_NetworkMain(t *testing.T) {
	t.Run("Handles uncompressed keys", func(t *testing.T) {
		ts, err := New(&Config{
			WIF: "5JZSYpPRedE3Thajcm5ink8KJaX9hC87Gk9wFHGH9VDvzuDFFTR",
			Fee: int64(10000),
		})
		require.NoError(t, err)
		assert.Equal(t, btc.NetworkMain, ts.Network())
		assert.Equal(t, ts.address.EncodeAddress(), "1MTPL1Ni2iD34eD4WVMpPHzbX4V5QUq1em")
	})

	t.Run("Handles compressed keys", func(t *testing.T) {
		ts, err := New(&Config{
			WIF: "L2utUUeVtvkzduyTmvjRb2Mxy4jAbtJjtHLdqFa6Ph1WTGNQo13H",
			Fee: int64(10000),
		})
		require.NoError(t, err)
		assert.Equal(t, btc.NetworkMain, ts.Network())
		assert.Equal(t, ts.address.EncodeAddress(), "1EXGUstoazmXFu2E9VRRPThWz6T9C3vK2V")
	})

}

func TestTimestamperTimestampHash(t *testing.T) {
	feeAmount := 15000
	t.Run("Handles compressed keys", func(t *testing.T) {
		ctx := context.Background()
		mock := &btctesting.Mock{}
		mock.MockFindUnspent.Fn = func(context.Context, *types.ReversedBytes20, int64) (btc.UnspentResult, error) {
			PKScriptHex := "76a914bf1e72331f8018f66faec356a04ca98b35bf5ee288ac"
			PKScript, _ := hex.DecodeString(PKScriptHex)
			output := btc.Output{Index: 0, PKScript: PKScript, Value: 14745268}
			if err := output.TXHash.Unstring("e35297e10fde340e5d0e2200de20f314f3851ea683d06feccf2f8bef6dd337d5"); err != nil {
				return btc.UnspentResult{}, err
			}

			return btc.UnspentResult{
				Outputs: []btc.Output{output},
				Sum:     14745268,
			}, nil
		}

		ts, err := New(&Config{
			WIF:           "cQNA7W1beoBJsefQQeznRoYT6XH9HkpU98V2S4ZUaWNxVPPT1qEk",
			UnspentFinder: mock,
			Broadcaster:   mock,
			Fee:           int64(feeAmount),
		})
		require.NoError(t, err)

		_, err = ts.TimestampHash(ctx, chainscripttest.RandomHash())
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.MockBroadcast.CalledCount)
	})

	t.Run("Handles uncompressed keys", func(t *testing.T) {
		ctx := context.Background()
		mock := &btctesting.Mock{}
		mock.MockFindUnspent.Fn = func(context.Context, *types.ReversedBytes20, int64) (btc.UnspentResult, error) {
			PKScriptHex := "76a914105647e641ac3104eef924e16b77378964d2930b88ac"
			PKScript, _ := hex.DecodeString(PKScriptHex)
			output := btc.Output{Index: 0, PKScript: PKScript, Value: 14745268}
			if err := output.TXHash.Unstring("60c8c843f29be77134097d105743013093cc115d4468690d8a9c2f9c8950ed20"); err != nil {
				return btc.UnspentResult{}, err
			}

			return btc.UnspentResult{
				Outputs: []btc.Output{output},
				Sum:     14745268,
			}, nil
		}

		ts, err := New(&Config{
			WIF:           "92EPR478AMbsHGqkPqN2TdMEe6BboHRjPUhhM1qTxQEWsmrD461",
			UnspentFinder: mock,
			Broadcaster:   mock,
			Fee:           int64(feeAmount),
		})
		require.NoError(t, err)

		_, err = ts.TimestampHash(ctx, chainscripttest.RandomHash())
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.MockBroadcast.CalledCount)
	})

	t.Run("Only use UTXO greater than transaction fee", func(t *testing.T) {
		ctx := context.Background()
		feeAmount := 15000
		mock := &btctesting.Mock{}
		mock.MockFindUnspent.Fn = func(context.Context, *types.ReversedBytes20, int64) (btc.UnspentResult, error) {
			PKScriptHex := "76a914105647e641ac3104eef924e16b77378964d2930b88ac"
			PKScript, _ := hex.DecodeString(PKScriptHex)
			output1 := btc.Output{Index: 0, PKScript: PKScript, Value: feeAmount / 2}
			if err := output1.TXHash.Unstring("60c8c843f29be77134097d105743013093cc115d4468690d8a9c2f9c8950ed20"); err != nil {
				return btc.UnspentResult{}, err
			}
			output2 := btc.Output{Index: 0, PKScript: PKScript, Value: feeAmount / 2}
			if err := output2.TXHash.Unstring("60c8c843f29be77134097d105743013093cc115d4468690d8a9c2f9c8950ed20"); err != nil {
				return btc.UnspentResult{}, err
			}

			return btc.UnspentResult{
				Outputs: []btc.Output{output1, output2},
				Sum:     int64(feeAmount),
			}, nil
		}

		ts, err := New(&Config{
			WIF:           "92EPR478AMbsHGqkPqN2TdMEe6BboHRjPUhhM1qTxQEWsmrD461",
			UnspentFinder: mock,
			Broadcaster:   mock,
			Fee:           int64(feeAmount),
		})
		require.NoError(t, err)

		_, err = ts.TimestampHash(ctx, chainscripttest.RandomHash())
		assert.EqualError(t, err, "btc error 9: adress 0b93d2648937776be124f9ee0431ac41e6475610: no UTXO greater than transaction fee amount, refill required")
	})
}
