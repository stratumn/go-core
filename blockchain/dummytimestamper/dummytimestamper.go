// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// LICENSE file.

// Package dummytimestamper implements a fake blockchain timestamper which can be used for testing.
package dummytimestamper

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/stratumn/goprivate/blockchain"
	"github.com/stratumn/sdk/types"
)

const networkString = "dummytimestamper"

// Network is the identifier of the dummy network.
type Network struct{}

// Timestamper is the type that implements fmt.Stringer.
func (Network) String() string {
	return networkString
}

// Timestamper is the type that implements github.com/stratumn/goprivate/blockchain.Timestamper.
type Timestamper struct{}

// Network implements fmt.Stringer.
func (Timestamper) Network() blockchain.Network {
	return Network{}
}

// Timestamp implements github.com/stratumn/goprivate/blockchain.Timestamper.
func (Timestamper) Timestamp(data interface{}) (types.TransactionID, error) {
	js, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(js)
	return sum[:], nil
}

// TimestampHash implements github.com/stratumn/goprivate/blockchain.HashTimestamper.
func (Timestamper) TimestampHash(hash *types.Bytes32) (types.TransactionID, error) {
	sum := sha256.Sum256(hash[:])
	return sum[:], nil
}