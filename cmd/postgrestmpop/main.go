// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

// The command postgrestmpop starts a tmpop node with a postgresstore.
package main

import (
	"flag"

	"github.com/stratumn/sdk/postgresstore"
	"github.com/stratumn/sdk/tendermint"
	"github.com/stratumn/sdk/tmpop"
	"github.com/stratumn/sdk/validator"
)

var (
	cacheSize         = flag.Int("cacheSize", tmpop.DefaultCacheSize, "size of the cache of the storage tree")
	validatorFilename = flag.String("rules_filename", validator.DefaultFilename, "Path to filename containing validation rules")
	version           = "0.1.0"
	commit            = "00000000000000000000000000000000"
)

func init() {
	tendermint.RegisterFlags()
	postgresstore.RegisterFlags()
}

func main() {
	flag.Parse()

	a := postgresstore.InitializeWithFlags(version, commit)

	tmpopConfig := &tmpop.Config{Commit: commit, Version: version, CacheSize: *cacheSize, ValidatorFilename: *validatorFilename}

	tmpop.Run(a, tmpopConfig)
}