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

package store

import (
	"github.com/pkg/errors"
)

// Common errors that can be used by store implementations.
var (
	ErrLinkAlreadyExists       = errors.New("link already exists")
	ErrOutDegreeNotSupported   = errors.New("out degree is not supported by the current implementation")
	ErrUniqueMapEntry          = errors.New("unique map entry is set and map already has an initial link")
	ErrReferencingNotSupported = errors.New("filtering on referencing segments is not supported by the current implementation")
	ErrBatchFailed             = errors.New("cannot add to batch: failures have been detected")
)
