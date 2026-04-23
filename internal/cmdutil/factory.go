// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package cmdutil

import (
	"sync"

	"jodoo-cli/internal/client"
	"jodoo-cli/internal/core"
	"jodoo-cli/internal/credential"
	"jodoo-cli/internal/output"
)

// Invocation captures parsed top-level flags + IO. Built by the bootstrap
// helper before the cobra root is constructed.
type Invocation struct {
	Profile string
	IO      *IOStreams
}

// Factory is passed to every command. It lazily resolves config + client
// so commands that don't need API access (--help, completion, doctor with
// no key) don't trigger errors.
type Factory struct {
	IOStreams *IOStreams

	// ProfileOverride is the value of the --profile flag (or empty).
	ProfileOverride string

	mu     sync.Mutex
	config *core.CliConfig
	cfgErr error
}

// NewDefault returns a Factory using the system streams + parsed invocation.
func NewDefault(inv *Invocation) *Factory {
	io := inv.IO
	if io == nil {
		io = SystemStreams()
	}
	return &Factory{IOStreams: io, ProfileOverride: inv.Profile}
}

// Config loads (and caches) the resolved profile.
//
// On the first call after the user runs `jodoo-cli config init`, the
// keychain may hold the API key while config.json doesn't. We escalate to
// the keychain transparently and stamp the resolved key onto the cached
// CliConfig so subsequent calls in the same process don't re-query.
func (f *Factory) Config() (*core.CliConfig, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.config != nil {
		return f.config, nil
	}
	if f.cfgErr != nil {
		return nil, f.cfgErr
	}
	cfg, err := core.LoadResolved(f.ProfileOverride)
	if err != nil {
		// Try the keychain before giving up — `cfg` is still populated
		// with profile / base URL even when the API key is missing.
		if cfg != nil {
			if v, kerr := credential.GetFromKeychain(cfg.Profile); kerr == nil && v != "" {
				cfg.APIKey = v
				f.config = cfg
				return cfg, nil
			}
		}
		f.cfgErr = err
		return nil, err
	}
	f.config = cfg
	return cfg, nil
}

// MustConfig is a helper for commands that cannot proceed without a key.
// It returns an *output.ExitError so the root error handler can write the
// standard envelope.
func (f *Factory) MustConfig() (*core.CliConfig, error) {
	cfg, err := f.Config()
	if err != nil {
		var ce *core.ConfigError
		if asErr, ok := err.(*core.ConfigError); ok {
			ce = asErr
		} else if asErr2, ok := err.(interface{ Unwrap() error }); ok {
			if asErr3, ok := asErr2.Unwrap().(*core.ConfigError); ok {
				ce = asErr3
			}
		}
		if ce != nil {
			return nil, output.ErrWithHint(ce.Code, ce.Type, ce.Message, ce.Hint)
		}
		return nil, err
	}
	return cfg, nil
}

// NewAPIClient constructs an HTTP client bound to the resolved profile.
func (f *Factory) NewAPIClient() (*client.APIClient, error) {
	cfg, err := f.MustConfig()
	if err != nil {
		return nil, err
	}
	return client.New(cfg), nil
}

// ResetConfig drops the cached config (used by tests + config commands
// that just mutated the file on disk).
func (f *Factory) ResetConfig() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.config = nil
	f.cfgErr = nil
}
