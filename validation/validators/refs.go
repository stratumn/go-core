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

package validators

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

// Errors returned by the RefsValidator.
var (
	ErrMapIDMismatch   = errors.New("mapID doesn't match previous link")
	ErrParentNotFound  = errors.New("parent is missing from store")
	ErrProcessMismatch = errors.New("process doesn't match referenced link")
	ErrRefNotFound     = errors.New("reference is missing from store")
)

// RefsValidator validates link references (parent and refs).
type RefsValidator struct{}

// NewRefsValidator creates a new RefsValidator.
func NewRefsValidator() Validator {
	return &RefsValidator{}
}

// Validate all references (parent and refs).
func (v *RefsValidator) Validate(ctx context.Context, r store.SegmentReader, l *chainscript.Link) error {
	if len(l.PrevLinkHash()) > 0 {
		s, err := r.GetSegment(ctx, l.PrevLinkHash())
		if err != nil {
			return errors.WithStack(err)
		}

		if s == nil || s.Link == nil {
			return ErrParentNotFound
		}

		if s.Link.Meta.Process.Name != l.Meta.Process.Name {
			return ErrProcessMismatch
		}

		if s.Link.Meta.MapId != l.Meta.MapId {
			return ErrMapIDMismatch
		}
	}

	if len(l.Meta.Refs) > 0 {
		var lhs []chainscript.LinkHash
		for _, ref := range l.Meta.Refs {
			lhs = append(lhs, ref.LinkHash)
		}

		segments, err := r.FindSegments(ctx, &store.SegmentFilter{
			Pagination: store.Pagination{Limit: len(lhs)},
			LinkHashes: lhs,
		})
		if err != nil {
			return errors.WithStack(err)
		}

		if len(segments.Segments) != len(l.Meta.Refs) {
			return ErrRefNotFound
		}

		for _, ref := range l.Meta.Refs {
			found := false

			for _, s := range segments.Segments {
				if bytes.Equal(ref.LinkHash, s.LinkHash()) {
					found = true
					if ref.Process != s.Link.Meta.Process.Name {
						return ErrProcessMismatch
					}

					break
				}
			}

			if !found {
				return ErrRefNotFound
			}
		}
	}

	return nil
}

// ShouldValidate always evaluates to true, as all links should validate their
// references.
func (v *RefsValidator) ShouldValidate(*chainscript.Link) bool {
	return true
}

// Hash returns an empty hash since RefsValidator doesn't have any state.
func (v *RefsValidator) Hash() (*types.Bytes32, error) {
	return nil, nil
}