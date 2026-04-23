// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

// Package jodoo declares all curated shortcuts that wrap Jodoo APIs.
package jodoo

import "jodoo-cli/shortcuts/common"

// Shortcuts returns every shortcut in this package, in display order.
func Shortcuts() []common.Shortcut {
	out := []common.Shortcut{}
	out = append(out, appShortcuts()...)
	out = append(out, dataShortcuts()...)
	out = append(out, fileShortcuts()...)
	out = append(out, workflowShortcuts()...)
	out = append(out, contactShortcuts()...)
	return out
}
