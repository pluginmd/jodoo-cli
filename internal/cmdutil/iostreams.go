// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package cmdutil holds the dependency-injection factory + IO plumbing
// shared by every cobra command.
package cmdutil

import (
	"io"
	"os"
)

// IOStreams holds the standard input / output / error streams.
// Tests inject buffers; production wires it to os.{Stdin,Stdout,Stderr}.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// SystemStreams returns IOStreams pointing at the OS stdio handles.
func SystemStreams() *IOStreams {
	return &IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
}
