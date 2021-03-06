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

package storehttp

import (
	"fmt"

	"github.com/stratumn/go-core/jsonhttp"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

func newErrOffset(msg string) jsonhttp.ErrHTTP {
	if msg == "" {
		msg = "offset must be a positive integer"
	}

	return jsonhttp.NewErrHTTP(types.NewError(errorcode.InvalidArgument, store.Component, msg))
}

func newErrLimit(msg string) jsonhttp.ErrHTTP {
	if msg == "" {
		msg = fmt.Sprintf("limit must be a posive integer less than or equal to %d", store.MaxLimit)
	}

	return jsonhttp.NewErrHTTP(types.NewError(errorcode.InvalidArgument, store.Component, msg))
}

func newErrWithoutParent(msg string) jsonhttp.ErrHTTP {
	if msg == "" {
		msg = "withoutParent should be a boolean"
	}

	return jsonhttp.NewErrHTTP(types.NewError(errorcode.InvalidArgument, store.Component, msg))
}

func newErrPrevLinkHash(msg string) jsonhttp.ErrHTTP {
	if msg == "" {
		msg = "prevLinkHash must be a 64 byte long hexadecimal string"
	}

	return jsonhttp.NewErrHTTP(types.NewError(errorcode.InvalidArgument, store.Component, msg))
}

func newErrLinkHashes(msg string) jsonhttp.ErrHTTP {
	if msg == "" {
		msg = "linkHashes must be an array of 64 byte long hexadecimal string"
	}

	return jsonhttp.NewErrHTTP(types.NewError(errorcode.InvalidArgument, store.Component, msg))
}

func newErrReferencing(msg string) jsonhttp.ErrHTTP {
	if msg == "" {
		msg = "referencing linkHash must be a 64 byte long hexadecimal string"
	}

	return jsonhttp.NewErrHTTP(types.NewError(errorcode.InvalidArgument, store.Component, msg))
}
