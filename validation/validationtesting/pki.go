// Copyright 2017 Stratumn SAS. All rights reserved.
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

package validationtesting

import (
	"encoding/json"
	"fmt"

	"github.com/stratumn/go-core/validation/validators"
)

// These are the test key pairs of Alice and Bob.
const (
	AlicePrivateKey = "-----BEGIN ED25519 PRIVATE KEY-----\nME4CAQAwBQYDK2VwBEIEQByVNUFScxEsQJbHFjiV49lQ0OSGWqxXGSEV9CfD3RLc\n4HuxduXhOjSyr657IqXX4WBsj++R4pgRmqwJa9PN3W4=\n-----END ED25519 PRIVATE KEY-----\n"
	AlicePublicKey  = `-----BEGIN ED25519 PUBLIC KEY-----\nMCowBQYDK2VwAyEA4HuxduXhOjSyr657IqXX4WBsj++R4pgRmqwJa9PN3W4=\n-----END ED25519 PUBLIC KEY-----\n`

	BobPrivateKey = "-----BEGIN ED25519 PRIVATE KEY-----\nME4CAQAwBQYDK2VwBEIEQBNjYUKZhIQCu1a2DZde6jM5kSltWKqRXkim3MUeWyUT\nPAtD68Uo/tTD6zVSMpxdWb0J1SA7sVHumDI3LZRDGEM=\n-----END ED25519 PRIVATE KEY-----\n"
	BobPublicKey  = `-----BEGIN ED25519 PUBLIC KEY-----\nMCowBQYDK2VwAyEAPAtD68Uo/tTD6zVSMpxdWb0J1SA7sVHumDI3LZRDGEM=\n-----END ED25519 PUBLIC KEY-----\n`
)

// ValidAuctionJSONPKIConfig is a valid PKI schema for the auction process.
var ValidAuctionJSONPKIConfig = fmt.Sprintf(`
{
	"alice.vandenbudenmayer@stratumn.com": {
		"keys": ["%s"],
		"roles": ["employee"]
	},
	"Bob Wagner": {
		"keys": ["%s"],
		"roles": ["manager", "it"]
	}
}
`, AlicePublicKey, BobPublicKey)

// ValidChatJSONPKIConfig is a valid PKI schema for the chat process.
var ValidChatJSONPKIConfig = fmt.Sprintf(`
{
	"Bob Wagner": {
		"keys": ["%s"],
		"roles": ["manager", "it"]
	}
}
`, BobPublicKey)

// LoadPKI unmarshals a JSON-formatted PKI into a PKI struct.
func LoadPKI(rawPKI []byte) (*validators.PKI, error) {
	var pki validators.PKI
	if err := json.Unmarshal(rawPKI, &pki); err != nil {
		return nil, err
	}

	return &pki, nil
}