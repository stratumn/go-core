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

package agenttestcases

import (
	"encoding/hex"
	"testing"

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateSegmentOK tests the client's ability to handle a CreateSegment request.
func (f Factory) TestCreateSegmentOK(t *testing.T) {
	process, action := "test", "test"
	parent, _ := f.Client.CreateMap(process, nil, "test")

	segment, err := f.Client.CreateSegment(process, parent.LinkHash(), action, nil, "test")
	assert.NoError(t, err)
	assert.NotNil(t, segment)

	var data map[string]string
	err = segment.Link.StructurizeData(&data)
	require.NoError(t, err)

	assert.Equal(t, "test", data["title"])
}

// TestCreateSegmentWithRefs tests the client's ability to handle a CreateSegment request
// when a reference is passed.
func (f Factory) TestCreateSegmentWithRefs(t *testing.T) {
	process, action := "test", "test"
	parent, _ := f.Client.CreateMap(process, nil, "test")
	refs := []chainscript.LinkReference{{Process: "other", LinkHash: chainscripttest.RandomHash()}}

	segment, err := f.Client.CreateSegment(process, parent.LinkHash(), action, refs, "one")
	assert.NoError(t, err)
	assert.NotNil(t, segment)
	assert.NotNil(t, segment.Link.Meta.Refs)
	want, _ := cj.Marshal(refs)
	got, _ := cj.Marshal(segment.Link.Meta.Refs)
	assert.Equal(t, want, got)
}

// TestCreateSegmentWithBadRefs tests the client's ability to handle a CreateSegment request
// when a reference is passed.
func (f Factory) TestCreateSegmentWithBadRefs(t *testing.T) {
	process, action, arg := "test", "test", "wrongref"
	parent, _ := f.Client.CreateMap(process, nil, "test")
	refs := []chainscript.LinkReference{{Process: "wrong"}}

	segment, err := f.Client.CreateSegment(process, parent.LinkHash(), action, refs, arg)
	assert.Error(t, err, "missing segment or (process and linkHash)")
	assert.Contains(t, err.Error(), "linkHash should be a non empty string")
	assert.Nil(t, segment)
}

// TestCreateSegmentHandlesWrongProcess tests the client's ability to handle a CreateSegment request
// when the provided process does not exist.
func (f Factory) TestCreateSegmentHandlesWrongProcess(t *testing.T) {
	process, linkHash, action := "wrong", chainscripttest.RandomHash(), "test"
	segment, err := f.Client.CreateSegment(process, linkHash, action, nil, "test")
	assert.EqualError(t, err, "process 'wrong' does not exist")
	assert.Nil(t, segment)
}

// TestCreateSegmentHandlesWrongLinkHash tests the client's ability to handle a CreateSegment request
// when the provided parent's linkHash does not exist.
func (f Factory) TestCreateSegmentHandlesWrongLinkHash(t *testing.T) {
	linkHash, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	process, action := "test", "test"
	segment, err := f.Client.CreateSegment(process, linkHash, action, nil, "test")
	assert.EqualError(t, err, "Not Found")
	assert.Nil(t, segment)
}

// TestCreateSegmentHandlesWrongAction tests the client's ability to handle a CreateSegment request
// when the provided action does not exist.
func (f Factory) TestCreateSegmentHandlesWrongAction(t *testing.T) {
	process, action := "test", "wrong"
	parent, _ := f.Client.CreateMap(process, nil, "test")

	segment, err := f.Client.CreateSegment(process, parent.LinkHash(), action, nil, "test")
	assert.EqualError(t, err, "not found")
	assert.Nil(t, segment)
}

// TestCreateSegmentHandlesWrongActionArgs tests the client's ability to handle a CreateSegment request
// when the provided action's arguments do not match the actual ones.
func (f Factory) TestCreateSegmentHandlesWrongActionArgs(t *testing.T) {
	process, action := "test", "test"
	parent, _ := f.Client.CreateMap(process, nil, "test")

	segment, err := f.Client.CreateSegment(process, parent.LinkHash(), action, nil)
	assert.EqualError(t, err, "a title is required")
	assert.Nil(t, segment)
}
