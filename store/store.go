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

// Package store defines types to implement a store.
package store

import (
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/types"
)

const (
	// DefaultLimit is the default pagination limit.
	DefaultLimit = 20

	// MaxLimit is the maximum pagination limit.
	MaxLimit = 200
)

// Writer is the interface that wraps the Write methods of a Store.
type Writer interface {
	// Creates or updates a segment. Segments passed to this method are
	// assumed to be valid.
	SaveSegment(segment *cs.Segment) error

	// Deletes a segment by link hash. Returns the removed segment or nil
	// if not found.
	DeleteSegment(linkHash *types.Bytes32) (*cs.Segment, error)

	// Saves a value at a key.
	SaveValue(key []byte, value []byte) error

	// Deletes a value at a key.
	DeleteValue(key []byte) ([]byte, error)
}

// Reader is the interface that wraps the Read methods of a Store.
type Reader interface {
	// Get a segment by link hash. Returns nil if no match is found.
	GetSegment(linkHash *types.Bytes32) (*cs.Segment, error)

	// Find segments. Returns an empty slice if there are no results.
	FindSegments(filter *SegmentFilter) (cs.SegmentSlice, error)

	// Get all the existing map IDs.
	GetMapIDs(filter *MapFilter) ([]string, error)

	// Gets a value at a key
	GetValue(key []byte) ([]byte, error)
}

// Batch represents a database transaction
type Batch interface {
	Reader
	Writer

	// Write definitely writes the content of the Batch
	Write() error
}

// Adapter must be implemented by a store.
type Adapter interface {
	Reader
	Writer

	// Returns arbitrary information about the adapter.
	GetInfo() (interface{}, error)

	// Adds a channel that receives segments whenever they are saved.
	AddDidSaveChannel(chan *cs.Segment)

	// Creates a new Batch
	NewBatch() (Batch, error)
}

/* Updated store interfaces.
 * Existing stores should migrate to these interfaces.
 * New stores should implement only this interface.
 */

// SegmentReader is the interface for reading Segments from a store.
type SegmentReader interface {
	// Get a segment by link hash. Returns nil if no match is found.
	// Will return link and evidences (if there are some in that store).
	GetSegment(linkHash *types.Bytes32) (*cs.Segment, error)

	// Find segments. Returns an empty slice if there are no results.
	// Will return links and evidences (if there are some).
	FindSegments(filter *SegmentFilter) (cs.SegmentSlice, error)

	// Get all the existing map IDs.
	GetMapIDs(filter *MapFilter) ([]string, error)
}

// LinkWriter is the interface for writing links to a store.
// Links are immutable and cannot be deleted.
type LinkWriter interface {
	// Create the immutable part of a segment.
	// The input link is expected to be valid.
	// Returns the link hash or an error.
	CreateLink(link *cs.Link) (*types.Bytes32, error)
}

// EvidenceReader is the interface for reading segment evidence from a store.
type EvidenceReader interface {
	// Get the evidences for a segment.
	// Can return a nil error with an empty evidence slice if
	// the segment currently doesn't have evidence.
	GetEvidences(linkHash *types.Bytes32) (*cs.Evidences, error)
}

// EvidenceWriter is the interface for adding evidence to a segment in a store.
type EvidenceWriter interface {
	// Add an evidence to a segment.
	AddEvidence(linkHash *types.Bytes32, evidence *cs.Evidence) error
}

// EvidenceStore is the interface for storing and reading segment evidence.
type EvidenceStore interface {
	EvidenceReader
	EvidenceWriter
}

// BatchV2 represents a database transaction.
// It will be renamed to Batch when the old interface will be removed.
type BatchV2 interface {
	SegmentReader
	LinkWriter

	// Write definitely writes the content of the Batch
	Write() error
}

// EventType lets you know the kind of event received.
// A client should ignore events it doesn't care about or doesn't understand.
type EventType int

const (
	// SavedLink means that a segment link was saved.
	SavedLink EventType = iota
	// SavedEvidence means that a segment evidence was saved.
	SavedEvidence
)

// Event is the object stores send to notify of important events.
type Event struct {
	EventType EventType
	Details   interface{}
}

// AdapterV2 is the new store interface. Once all stores are migrated
// it will be renamed to Adapter and the old interface will be removed.
type AdapterV2 interface {
	SegmentReader
	LinkWriter
	EvidenceStore

	// Returns arbitrary information about the adapter.
	GetInfo() (interface{}, error)

	// Adds a channel that receives events from the store.
	AddStoreEventChannel(chan *Event)

	// Creates a new Batch
	NewBatch() (BatchV2, error)
}

// KeyValueReader is the interface for reading key-value pairs.
type KeyValueReader interface {
	GetValue(key []byte) ([]byte, error)
}

// KeyValueWriter is the interface for writing key-value pairs.
type KeyValueWriter interface {
	SetValue(key []byte, value []byte) error
	DeleteValue(key []byte) ([]byte, error)
}

// KeyValueStore is the interface for a key-value store.
// Some store adapters will implement this interface, but not all.
type KeyValueStore interface {
	KeyValueReader
	KeyValueWriter
}

/* End of updated store interfaces. */

// Pagination contains pagination options.
type Pagination struct {
	// Index of the first entry.
	Offset int `json:"offset"`

	// Maximum number of entries.
	Limit int `json:"limit"`
}

// SegmentFilter contains filtering options for segments.
// If PrevLinkHash is not nil, MapID is ignored because a previous link hash
// implies the map ID of the previous segment.
type SegmentFilter struct {
	Pagination `json:"pagination"`

	// Map IDs the segments must have.
	MapIDs []string `json:"mapIds"`

	// Process name is optionnal.
	Process string `json:"process"`

	// A previous link hash the segments must have.
	// nil makes this attribute as optional
	// empty string is to search Segments without parent
	PrevLinkHash *string `json:"prevLinkHash"`

	// A slice of linkHashes to search Segments.
	// This attribute is optional.
	LinkHashes []*types.Bytes32 `json:"linkHashes"`

	// A slice of tags the segments must all contain.
	Tags []string `json:"tags"`
}

// MapFilter contains filtering options for segments.
type MapFilter struct {
	Pagination `json:"pagination"`

	// Process name is optionnal.
	Process string `json:"process"`
}

// PaginateStrings paginates a list of strings
func (p *Pagination) PaginateStrings(a []string) []string {
	l := len(a)
	if p.Offset >= l {
		return []string{}
	}

	end := min(l, p.Offset+p.Limit)
	return a[p.Offset:end]
}

// PaginateSegments paginate a list of segments
func (p *Pagination) PaginateSegments(a cs.SegmentSlice) cs.SegmentSlice {
	l := len(a)
	if p.Offset >= l {
		return cs.SegmentSlice{}
	}

	end := min(l, p.Offset+p.Limit)
	return a[p.Offset:end]
}

// Min of two ints, duh.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Match check if segment matches with filter
func (filter SegmentFilter) Match(segment *cs.Segment) bool {
	if segment == nil {
		return false
	}

	if filter.PrevLinkHash != nil {
		prevLinkHash := segment.Link.GetPrevLinkHash()
		if *filter.PrevLinkHash == "" {
			if prevLinkHash != nil {
				return false
			}
		} else {
			filterPrevLinkHash, err := types.NewBytes32FromString(*filter.PrevLinkHash)
			if err != nil || prevLinkHash == nil || *filterPrevLinkHash != *prevLinkHash {
				return false
			}
		}
	}

	if len(filter.LinkHashes) > 0 {
		var match bool
		for _, linkHash := range filter.LinkHashes {
			if linkHash.Compare(segment.GetLinkHash()) == 0 {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	if filter.Process != "" && filter.Process != segment.Link.GetProcess() {
		return false
	}

	if len(filter.MapIDs) > 0 {
		var match = false
		mapID := segment.Link.GetMapID()
		for _, filterMapIDs := range filter.MapIDs {
			match = match || filterMapIDs == mapID
		}
		if !match {
			return false
		}
	}

	if len(filter.Tags) > 0 {
		tags := segment.Link.GetTagMap()
		for _, tag := range filter.Tags {
			if _, ok := tags[tag]; !ok {
				return false
			}
		}
	}
	return true
}

// Match check if segment matches with filter
func (filter MapFilter) Match(segment *cs.Segment) bool {
	if segment == nil {
		return false
	}
	if filter.Process != "" && filter.Process != segment.Link.GetProcess() {
		return false
	}
	return true
}
