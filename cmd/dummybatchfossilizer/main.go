// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"

	"github.com/stratumn/go/fossilizer/fossilizerhttp"
	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/goprivate/batchfossilizer"
	"github.com/stratumn/goprivate/bcbatchfossilizer"
	"github.com/stratumn/goprivate/blockchain/dummytimestamper"
	"github.com/stratumn/goprivate/merkle"
)

var (
	port             = flag.String("port", fossilizerhttp.DefaultPort, "server port")
	certFile         = flag.String("tlscert", "", "TLS certificate file")
	keyFile          = flag.String("tlskey", "", "TLS private key file")
	numResultWorkers = flag.Int("workers", fossilizerhttp.DefaultNumResultWorkers, "number of result workers")
	callbackTimeout  = flag.Duration("callbacktimeout", fossilizerhttp.DefaultCallbackTimeout, "callback requests timeout")
	verbose          = flag.Bool("verbose", fossilizerhttp.DefaultVerbose, "verbose output")
	interval         = flag.Duration("interval", batchfossilizer.DefaultInterval, "batch interval")
	maxLeaves        = flag.Int("maxleaves", batchfossilizer.DefaultMaxLeaves, "maximum number of leaves in a Merkle tree")
	path             = flag.String("path", "", "an optional path to store files")
	archive          = flag.Bool("archive", batchfossilizer.DefaultArchive, "whether to archive completed batches (requires path)")
	exitBatch        = flag.Bool("exitbatch", batchfossilizer.DefaultStopBatch, "whether to do a batch on exit")
	fsync            = flag.Bool("fsync", batchfossilizer.DefaultFSync, "whether to fsync after saving a pending hash")
	version          = ""
	commit           = ""
)

func init() {
	log.SetPrefix("dummybatchfossilizer ")
}

func main() {
	flag.Parse()

	a, err := bcbatchfossilizer.New(&bcbatchfossilizer.Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		Version:   version,
		Commit:    commit,
		Interval:  *interval,
		MaxLeaves: *maxLeaves,
		Path:      *path,
		Archive:   *archive,
		StopBatch: *exitBatch,
		FSync:     *fsync,
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Fatal(a.Start())
	}()

	defer func() {
		if err := a.Stop(); err != nil {
			log.Printf("Error: %s\n", err)
		}
	}()

	c := &fossilizerhttp.Config{
		Config: jsonhttp.Config{
			Port:     *port,
			CertFile: *certFile,
			KeyFile:  *keyFile,
			Verbose:  *verbose,
		},
		NumResultWorkers: *numResultWorkers,
		CallbackTimeout:  *callbackTimeout,
		MinDataLen:       merkle.HashByteSize * 2,
		MaxDataLen:       merkle.HashByteSize * 2,
	}
	h := fossilizerhttp.New(a, c)

	log.Printf("Listening on %s", *port)

	if err := h.ListenAndServe(); err != nil {
		log.Fatalf("Fatal: %s\n")
	}
}
