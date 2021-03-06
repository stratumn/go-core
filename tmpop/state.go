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

package tmpop

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/bufferedbatch"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
	"github.com/stratumn/go-core/validation"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stratumn/merkle"
)

// State represents the app states, separating the committed state (for queries)
// from the working state (for CheckTx and DeliverTx).
type State struct {
	previousAppHash []byte
	// The same validator is used for a whole commit
	// When beginning a new block, the validator can
	// be updated.
	validator     validators.Validator
	validationCfg *validation.Config

	adapter            store.Adapter
	deliveredLinks     store.Batch
	deliveredLinksList []*chainscript.Link
	checkedLinks       store.Batch
}

// NewState creates a new State.
func NewState(ctx context.Context, a store.Adapter, config *Config) (*State, error) {
	deliveredLinks, err := a.NewBatch(ctx)
	if err != nil {
		return nil, err
	}

	// With transactional databases we cannot really use two transactions as they'd lock each other
	// (more exactly, checked links would lock out delivered links)
	checkedLinks := bufferedbatch.NewBatch(ctx, a)

	state := &State{
		adapter:        a,
		deliveredLinks: deliveredLinks,
		checkedLinks:   checkedLinks,
		validationCfg:  config.Validation,
	}

	return state, nil
}

// UpdateValidators updates validators if a new version is available.
func (s *State) UpdateValidators(ctx context.Context) {
	if s.validationCfg == nil {
		return
	}

	// We temporarily re-load validation rules for each block.
	// This will change once the new governance flows are implemented.
	v, err := validation.LoadFromFile(ctx, s.validationCfg)
	if err != nil {
		monitoring.TxLogEntry(ctx).
			WithError(err).
			Warn("could not load validation rules")
		return
	}

	mv := validators.NewMultiValidator(v.Flatten())
	s.validator = mv
}

// Check checks if creating this link is a valid operation
func (s *State) Check(ctx context.Context, link *chainscript.Link) *ABCIError {
	return s.checkLinkAndAddToBatch(ctx, link, s.checkedLinks)
}

// Deliver adds a link to the list of links to be committed
func (s *State) Deliver(ctx context.Context, link *chainscript.Link) *ABCIError {
	res := s.checkLinkAndAddToBatch(ctx, link, s.deliveredLinks)
	if res.IsOK() {
		s.deliveredLinksList = append(s.deliveredLinksList, link)
	}
	return res
}

// checkLinkAndAddToBatch validates the link's format and runs the validations (signatures, schema)
func (s *State) checkLinkAndAddToBatch(ctx context.Context, link *chainscript.Link, batch store.Batch) *ABCIError {
	if err := link.Validate(ctx); err != nil {
		return &ABCIError{
			Code: CodeTypeValidation,
			Log:  fmt.Sprintf("Link validation failed %v: %v", link, err),
		}
	}

	if err := validators.NewRefsValidator().Validate(ctx, batch, link); err != nil {
		return &ABCIError{
			Code: CodeTypeValidation,
			Log:  fmt.Sprintf("Link references validation failed %v: %v", link, err),
		}
	}

	if s.validator != nil {
		err := s.validator.Validate(ctx, batch, link)
		if err != nil {
			return &ABCIError{
				Code: CodeTypeValidation,
				Log:  fmt.Sprintf("Link validation rules failed: %v", err),
			}
		}
	}

	if _, err := batch.CreateLink(ctx, link); err != nil {
		return &ABCIError{
			Code: CodeTypeInternalError,
			Log:  err.Error(),
		}
	}

	return nil
}

// Commit commits the delivered links, resets delivered and checked state,
// and returns the hash for the commit and the list of committed links.
func (s *State) Commit(ctx context.Context) ([]byte, []*chainscript.Link, error) {
	appHash, err := s.computeAppHash()
	if err != nil {
		return nil, nil, err
	}

	if err := s.deliveredLinks.Write(ctx); err != nil {
		return nil, nil, err
	}

	if s.deliveredLinks, err = s.adapter.NewBatch(ctx); err != nil {
		return nil, nil, err
	}
	s.checkedLinks = bufferedbatch.NewBatch(ctx, s.adapter)

	committedLinks := make([]*chainscript.Link, len(s.deliveredLinksList))
	copy(committedLinks, s.deliveredLinksList)
	s.deliveredLinksList = nil

	return appHash, committedLinks, nil
}

func (s *State) computeAppHash() ([]byte, error) {
	var validatorHash []byte
	if s.validator != nil {
		h, err := s.validator.Hash()
		if err != nil {
			return nil, err
		}
		validatorHash = h
	}

	var merkleRoot []byte
	if len(s.deliveredLinksList) > 0 {
		var treeLeaves [][]byte
		for _, link := range s.deliveredLinksList {
			linkHash, _ := link.Hash()
			treeLeaves = append(treeLeaves, linkHash)
		}

		merkle, err := merkle.NewStaticTree(treeLeaves)
		if err != nil {
			return nil, types.WrapError(err, errorcode.InvalidArgument, Name, "could not create merkle tree")
		}

		merkleRoot = merkle.Root()
	}

	return ComputeAppHash(s.previousAppHash, validatorHash, merkleRoot), nil
}

// ComputeAppHash computes the app hash from its required parts.
func ComputeAppHash(previous []byte, validator []byte, root []byte) []byte {
	appHash := sha256.Sum256(append(
		previous,
		append(
			validator,
			root...,
		)...,
	))

	return appHash[:]
}
