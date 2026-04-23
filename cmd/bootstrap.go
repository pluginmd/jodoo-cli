// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package cmd

import (
	"jodoo-cli/internal/cmdutil"
)

// BootstrapInvocationContext extracts the subset of top-level flags we
// must know before constructing the cobra root (--profile is enough for
// now). Reads them straight from os.Args because cobra hasn't run yet.
//
// Unrecognized args are ignored — cobra will validate them itself.
func BootstrapInvocationContext(args []string) (*cmdutil.Invocation, error) {
	inv := &cmdutil.Invocation{IO: cmdutil.SystemStreams()}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--profile":
			if i+1 < len(args) {
				inv.Profile = args[i+1]
				i++
			}
		default:
			if v, ok := splitEqual(args[i], "--profile="); ok {
				inv.Profile = v
			}
		}
	}
	return inv, nil
}

func splitEqual(arg, prefix string) (string, bool) {
	if len(arg) <= len(prefix) {
		return "", false
	}
	if arg[:len(prefix)] != prefix {
		return "", false
	}
	return arg[len(prefix):], true
}
