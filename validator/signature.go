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

package validator

import (
	"encoding/base64"

	cj "github.com/gibson042/canonicaljson-go"
	jmespath "github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/validator/signature"
)

var (

	// ErrMissingSignature is returned when there are no signatures in the link.
	ErrMissingSignature = errors.New("signature validation requires link.signatures to contain at least one element")

	// ErrEmptyPayload is returned when the JMESPATH query didn't match any element of the link.
	ErrEmptyPayload = errors.New("JMESPATH query does not match any link data")
)

// signatureValidatorConfig contains everything a signatureValidator needs to
// validate links.
type signatureValidatorConfig struct {
	*validatorBaseConfig
	pki *PKI
}

// newSignatureValidatorConfig creates a signatureValidatorConfig for a given process and type.
func newSignatureValidatorConfig(process, id, linkType string, pki *PKI) (*signatureValidatorConfig, error) {
	baseConfig, err := newValidatorBaseConfig(process, id, linkType)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &signatureValidatorConfig{validatorBaseConfig: baseConfig, pki: pki}, nil
}

// signatureValidator validates the json signature of a link's state.
type signatureValidator struct {
	config *signatureValidatorConfig
}

func newSignatureValidator(config *signatureValidatorConfig) validator {
	return &signatureValidator{config: config}
}

// Validate validates the signature of a link's state.
func (sv signatureValidator) Validate(_ store.SegmentReader, link *cs.Link) error {
	if !sv.config.shouldValidate(link) {
		return nil
	}

	if len(link.Signatures) == 0 {
		return ErrMissingSignature
	}

	for _, sig := range link.Signatures {

		// don't check decoding errors here, this is done in link.Validate() beforehand
		keyBytes, _ := base64.StdEncoding.DecodeString(sig.PublicKey)
		sigBytes, _ := base64.StdEncoding.DecodeString(sig.Signature)

		payload, err := jmespath.Search(sig.Payload, link)
		if err != nil {
			return errors.Wrapf(err, "failed to execute jmespath query")
		}
		if payload == nil {
			return ErrEmptyPayload
		}

		payloadBytes, err := cj.Marshal(payload)
		if err != nil {
			return errors.WithStack(err)
		}

		if err := signature.Verify(sig.Type, keyBytes, sigBytes, payloadBytes); err != nil {
			return errors.WithStack(err)
		}
	}
	// TODO: check that
	// - public keys match PKI of rules.json
	// - required signatures for this action are present/valid

	return nil
}
