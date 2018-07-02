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

// Package cs defines types to work with Chainscripts.
package cs

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"reflect"

	cj "github.com/gibson042/canonicaljson-go"
	jmespath "github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"
	sigencoding "github.com/stratumn/go-crypto/encoding"
	"github.com/stratumn/go-crypto/keys"
	"github.com/stratumn/go-crypto/signatures"
	"github.com/stratumn/go-indigocore/types"
)

// Segment contains a link and meta data about the link.
type Segment struct {
	Link Link        `json:"link"`
	Meta SegmentMeta `json:"meta"`
}

// GetLinkHash returns the link ID as bytes.
// It assumes the segment is valid.
func (s *Segment) GetLinkHash() *types.Bytes32 {
	return s.Meta.GetLinkHash()
}

// GetLinkHashString returns the link ID as a string.
// It assumes the segment is valid.
func (s *Segment) GetLinkHashString() string {
	return s.Meta.GetLinkHashString()
}

// HashLink hashes the segment link and stores it into the Meta
func (s *Segment) HashLink() (string, error) {
	return s.Link.HashString()
}

// SetLinkHash overwrites the segment LinkHash using HashLink()
func (s *Segment) SetLinkHash() error {
	linkHash, err := s.HashLink()
	if err != nil {
		return err
	}

	s.Meta.LinkHash = linkHash
	return nil
}

// IsEmpty checks if a segment is empty (nil)
func (s *Segment) IsEmpty() bool {
	return reflect.DeepEqual(*s, Segment{})
}

// Validate checks for errors in a segment
func (s *Segment) Validate(ctx context.Context, getSegment GetSegmentFunc) error {
	if s.Meta.LinkHash == "" {
		return errors.New("meta.linkHash should be a non empty string")
	}

	want, err := s.HashLink()
	if err != nil {
		return err
	}
	if got := s.GetLinkHashString(); want != got {
		return errors.New("meta.linkHash is not in sync with link")
	}

	return s.Link.Validate(ctx, getSegment)
}

// GetSegmentFunc is the function signature to retrieve a Segment
type GetSegmentFunc func(ctx context.Context, linkHash *types.Bytes32) (*Segment, error)

// SegmentMeta contains additional information about the segment and a proof of existence
type SegmentMeta struct {
	Evidences Evidences `json:"evidences"`
	LinkHash  string    `json:"linkHash"`
}

// GetLinkHash returns the link ID as bytes.
// It assumes the segment is valid.
func (s *SegmentMeta) GetLinkHash() *types.Bytes32 {
	b, _ := types.NewBytes32FromString(s.LinkHash)
	return b
}

// GetLinkHashString returns the link ID as a string.
// It assumes the segment is valid.
func (s *SegmentMeta) GetLinkHashString() string {
	return s.LinkHash
}

// AddEvidence sets the segment evidence
func (s *SegmentMeta) AddEvidence(evidence Evidence) error {
	return s.Evidences.AddEvidence(evidence)
}

// GetEvidence gets an evidence from a provider
func (s *SegmentMeta) GetEvidence(provider string) *Evidence {
	return s.Evidences.GetEvidence(provider)
}

// FindEvidences find all evidences generated by a specified backend ("TMPop", "bcbatchfossilizer"...)
func (s *SegmentMeta) FindEvidences(backend string) (res Evidences) {
	return s.Evidences.FindEvidences(backend)
}

// SegmentReference is a reference to a segment or a linkHash
type SegmentReference struct {
	Process  string `json:"process"`
	LinkHash string `json:"linkHash"`
}

// LinkMeta contains the typed meta data of a Link and data
type LinkMeta struct {
	MapID        string                 `json:"mapId"`
	Process      string                 `json:"process"`
	Action       string                 `json:"action"`
	Type         string                 `json:"type"`
	Inputs       []interface{}          `json:"inputs"`
	Tags         []string               `json:"tags"`
	Priority     float64                `json:"priority"`
	PrevLinkHash string                 `json:"prevLinkHash"`
	Refs         []SegmentReference     `json:"refs"`
	Data         map[string]interface{} `json:"data"`
}

// Link contains a state and meta data about the state.
type Link struct {
	State      map[string]interface{} `json:"state"`
	Meta       LinkMeta               `json:"meta"`
	Signatures []*Signature           `json:"signatures"`
}

// Hash hashes the link
func (l *Link) Hash() (*types.Bytes32, error) {
	jsonLink, err := cj.Marshal(l)
	if err != nil {
		return nil, err
	}
	byteLinkHash := sha256.Sum256(jsonLink)
	linkHash := types.Bytes32(byteLinkHash)
	return &linkHash, nil
}

// HashString hashes the link and returns a string
func (l *Link) HashString() (string, error) {
	hash, err := l.Hash()
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}

// GetPrevLinkHash returns the previous link hash as a bytes.
// It will return nil if the previous link hash is null.
func (m *LinkMeta) GetPrevLinkHash() *types.Bytes32 {
	if m.PrevLinkHash != "" {
		b, _ := types.NewBytes32FromString(m.PrevLinkHash)
		return b
	}
	return nil
}

// GetTagMap returns the tags as a map of string to empty structs (a set).
func (m *LinkMeta) GetTagMap() map[string]struct{} {
	tags := map[string]struct{}{}
	for _, v := range m.Tags {
		tags[v] = struct{}{}
	}
	return tags
}

// Validate checks for errors in a link.
// It checks the validity of: format, signatures and references.
func (l *Link) Validate(ctx context.Context, getSegment GetSegmentFunc) error {
	if l.Meta.Process == "" {
		return errors.New("link.meta.process should be a non empty string")
	}
	if l.Meta.MapID == "" {
		return errors.New("link.meta.mapId should be a non empty string")
	}

	for _, tag := range l.Meta.Tags {
		if tag == "" {
			return errors.New("link.meta.tags should be an array of non empty string")
		}
	}

	if _, err := l.Hash(); err != nil {
		return err
	}

	if err := l.validateSignatures(); err != nil {
		return err
	}

	return l.validateReferences(ctx, getSegment)
}

func (l *Link) validateReferences(ctx context.Context, getSegment GetSegmentFunc) error {
	for refIdx, ref := range l.Meta.Refs {
		if ref.Process == "" {
			return errors.Errorf("link.meta.refs[%d].process should be a non empty string", refIdx)
		}
		linkHash, err := types.NewBytes32FromString(ref.LinkHash)
		if err != nil {
			return errors.Errorf("link.meta.refs[%d].linkHash should be a bytes32 field", refIdx)
		}
		if l.Meta.Process == ref.Process && getSegment != nil {
			if seg, err := getSegment(ctx, linkHash); err != nil {
				return errors.Wrapf(err, "link.meta.refs[%d] segment should be retrieved", refIdx)
			} else if seg == nil {
				return errors.Errorf("link.meta.refs[%d] segment is nil", refIdx)
			}
		}
		// Segment from another process is not retrieved because it could be in another store
	}
	return nil
}

func (l *Link) validateSignatures() error {
	if l.Signatures != nil {
		for _, sig := range l.Signatures {
			if sig.Type == "" {
				return errors.New("signature.Type cannot be empty")
			} else if _, err := keys.ParsePublicKey([]byte(sig.PublicKey)); err != nil {
				return errors.Wrapf(err, "failed to parse public key [%s]", sig.PublicKey)
			} else if _, err := sigencoding.DecodePEM([]byte(sig.Signature), signatures.SignaturePEMLabel); err != nil {
				return errors.Wrapf(err, "failed to parse signature [%s]", sig.Signature)
			} else if _, err := jmespath.Compile(sig.Payload); err != nil {
				return errors.Errorf("signature.Payload [%s] has to be a JMESPATH expression, got: %s", sig.Payload, err.Error())
			}

			if err := sig.Verify(l); err != nil {
				return err
			}
		}
	}
	return nil
}

// Segmentify returns a segment from a link,
// filling the link hash and appropriate meta.
func (l Link) Segmentify() *Segment {
	linkHash, _ := l.HashString()
	return &Segment{
		Link: l,
		Meta: SegmentMeta{
			LinkHash: linkHash,
		},
	}
}

// Clone returns a copy of the link.
// Since it uses the json Marshaler, the state is automatically
// converted to a map[string]interface if it is a custom struct.
func (l *Link) Clone() (*Link, error) {
	var clone Link
	js, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(js, &clone); err != nil {
		return nil, err
	}

	return &clone, nil
}

// Search executes a JMESPATH query on the link and return the matching payload.
func (l *Link) Search(jsonQuery string) (interface{}, error) {
	// Cloning the link allows the state to be serialized to
	// JSON which is needed for the jmespath query.
	cloned, err := l.Clone()
	if err != nil {
		return nil, err
	}
	return jmespath.Search(jsonQuery, cloned)
}

// SegmentSlice is a slice of segment pointers.
type SegmentSlice []*Segment

// Len implements sort.Interface.Len.
func (s SegmentSlice) Len() int {
	return len(s)
}

// Swap implements sort.Interface.Swap.
func (s SegmentSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implements sort.Interface.Less.
func (s SegmentSlice) Less(i, j int) bool {
	var (
		s1 = s[i]
		s2 = s[j]
		p1 = s1.Link.Meta.Priority
		p2 = s2.Link.Meta.Priority
	)

	if p1 > p2 {
		return true
	}

	if p1 < p2 {
		return false
	}

	return s1.GetLinkHashString() < s2.GetLinkHashString()
}

// LinkHashes is a slice of Bytes32-formatted link hashes
type LinkHashes []*types.Bytes32

// NewLinkHashesFromStrings creates a slice of bytes-formatted link hashes
// from a slice of string-formatted link hashes
func NewLinkHashesFromStrings(linkHashesStr []string) (LinkHashes, error) {
	linkHashes := LinkHashes{}
	for _, lh := range linkHashesStr {
		linkHashBytes, err := types.NewBytes32FromString(lh)
		if err != nil {
			return nil, err
		}
		linkHashes = append(linkHashes, linkHashBytes)
	}
	return linkHashes, nil
}
