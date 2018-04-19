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

package validator

import (
	"context"
	"crypto/sha256"
	"io/ioutil"
	"path"
	"plugin"
	"strings"

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

const (
	golang = "go"

	// ErrLoadingPlugin is the error returned in case the plugin could not be loaded
	ErrLoadingPlugin = "Error while loading validation script for process %s and type %s"

	// ErrBadPlugin is the error returned in case the plugin is missing exported symbols
	ErrBadPlugin = "script does not implement the ScriptValidatorFunc type"

	// ErrBadScriptType is the error returned when the type of script does not match the supported ones
	ErrBadScriptType = "Validation engine does not handle script of type %s, valid types are %v"
)

var (
	validScriptTypes = []string{golang}
)

// ScriptValidatorFunc is the function called when enforcing a custom validation rule
type ScriptValidatorFunc = func(store.SegmentReader, *cs.Link) error

type scriptValidator struct {
	script     ScriptValidatorFunc
	scriptHash [32]byte
	config     *validatorBaseConfig
}

func checkScriptType(cfg *scriptConfig) error {
	switch cfg.Type {
	case golang:
		return nil
	default:
		return errors.Errorf(ErrBadScriptType, cfg.Type, validScriptTypes)
	}
}

func newScriptValidator(baseConfig *validatorBaseConfig, scriptCfg *scriptConfig, pluginsPath string) (Validator, error) {
	if err := checkScriptType(scriptCfg); err != nil {
		return nil, err
	}
	pluginFile := path.Join(pluginsPath, scriptCfg.File)
	p, err := plugin.Open(pluginFile)
	if err != nil {
		return nil, errors.Wrapf(err, ErrLoadingPlugin, baseConfig.Process, baseConfig.LinkType)
	}

	symbol, err := p.Lookup(strings.Title(baseConfig.LinkType))
	if err != nil {
		return nil, errors.Wrapf(err, ErrLoadingPlugin, baseConfig.Process, baseConfig.LinkType)
	}

	customValidator, ok := symbol.(ScriptValidatorFunc)
	if !ok {
		return nil, errors.Wrapf(errors.New(ErrBadPlugin), ErrLoadingPlugin, baseConfig.Process, baseConfig.LinkType)
	}

	// here we ignore the error since there is no way we cannot read the file if the plugin has be loaded successfully
	b, _ := ioutil.ReadFile(pluginFile)
	return &scriptValidator{
		config:     baseConfig,
		script:     customValidator,
		scriptHash: sha256.Sum256(b),
	}, nil
}

func (sv scriptValidator) Hash() (*types.Bytes32, error) {
	b, err := cj.Marshal(struct {
		ScriptHash [32]byte
		Config     *validatorBaseConfig
	}{
		ScriptHash: sv.scriptHash,
		Config:     sv.config,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	validationsHash := types.Bytes32(sha256.Sum256(b))
	return &validationsHash, nil
}

func (sv scriptValidator) ShouldValidate(link *cs.Link) bool {
	return sv.config.ShouldValidate(link)
}

func (sv scriptValidator) Validate(_ context.Context, storeReader store.SegmentReader, link *cs.Link) error {
	return sv.script(storeReader, link)
}