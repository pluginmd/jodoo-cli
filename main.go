// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT
//
// jodoo-cli — Jodoo (api.jodoo.com) CLI tool (Go implementation).
package main

import (
	"os"

	"jodoo-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
