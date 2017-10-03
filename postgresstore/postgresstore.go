// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

// Package postgresstore implements a store that saves all the segments in a
// PostgreSQL database. It requires PostgreSQL >= 9.5 for
// "ON CONFLICT DO UPDATE" support.
package postgresstore

import (
	"database/sql"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

const (
	// Name is the name set in the store's information.
	Name = "postgres"

	// Description is the description set in the store's information.
	Description = "Stratumn PostgreSQL Store"

	// DefaultURL is the default URL of the database.
	DefaultURL = "postgres://postgres@postgres/postgres?sslmode=disable"
)

const notFoundError = "sql: no rows in result set"

// Config contains configuration options for the store.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string

	// The URL of the PostgreSQL database, such as
	// "postgres://postgres@localhost/store?sslmode=disable".
	URL string
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// Store is the type that implements github.com/stratumn/sdk/store.Adapter.
type Store struct {
	*reader
	*writer
	config       *Config
	didSaveChans []chan *cs.Segment
	db           *sql.DB
	stmts        *stmts

	batches map[*Batch]*sql.Tx
}

// New creates an instance of a Store.
func New(config *Config) (*Store, error) {
	db, err := sql.Open("postgres", config.URL)
	if err != nil {
		return nil, err
	}
	return &Store{config: config, db: db, batches: make(map[*Batch]*sql.Tx)}, nil
}

// AddDidSaveChannel implements
// github.com/stratumn/sdk/fossilizer.Store.AddDidSaveChannel.
func (a *Store) AddDidSaveChannel(saveChan chan *cs.Segment) {
	a.didSaveChans = append(a.didSaveChans, saveChan)
}

// GetInfo implements github.com/stratumn/sdk/store.Adapter.GetInfo.
func (a *Store) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (a *Store) SaveSegment(segment *cs.Segment) error {
	curr, err := a.GetSegment(segment.GetLinkHash())
	if err != nil {
		return err
	}
	if curr != nil {
		segment, _ = curr.MergeMeta(segment)
	}

	a.writer.SaveSegment(segment)

	// Send saved segment to all the save channels without blocking.
	go func(chans []chan *cs.Segment) {
		for _, c := range chans {
			c <- segment
		}
	}(a.didSaveChans)

	return nil
}

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (a *Store) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	return a.reader.GetSegment(linkHash)
}

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (a *Store) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	return a.reader.FindSegments(filter)
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (a *Store) GetMapIDs(filter *store.MapFilter) ([]string, error) {
	return a.reader.GetMapIDs(filter)
}

// GetValue implements github.com/stratumn/sdk/store.Adapter.GetValue.
func (a *Store) GetValue(key []byte) ([]byte, error) {
	return a.reader.GetValue(key)
}

// NewBatch implements github.com/stratumn/sdk/store.Adapter.NewBatch.
func (a *Store) NewBatch() (store.Batch, error) {
	for b := range a.batches {
		if b.done {
			delete(a.batches, b)
		}
	}

	tx, err := a.db.Begin()
	if err != nil {
		return nil, err
	}
	b, err := NewBatch(tx)
	if err != nil {
		return nil, err
	}
	a.batches[b] = tx

	return b, nil
}

// Create creates the database tables and indexes.
func (a *Store) Create() error {
	for _, query := range sqlCreate {
		if _, err := a.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// Prepare prepares the database stmts.
// It should be called once before interacting with segments.
// It assumes the tables have been created using Create().
func (a *Store) Prepare() error {
	stmts, err := newStmts(a.db)
	if err != nil {
		return err
	}
	a.stmts = stmts
	a.reader = &reader{stmts: a.stmts.readStmts}
	a.writer = &writer{stmts: a.stmts.writeStmts}

	return nil
}

// Drop drops the database tables and indexes. It also rollbacks started batches.
func (a *Store) Drop() error {
	for b, tx := range a.batches {
		if !b.done {
			err := tx.Rollback()
			if err != nil {
				return err
			}
		}
	}

	for _, query := range sqlDrop {
		if _, err := a.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}
