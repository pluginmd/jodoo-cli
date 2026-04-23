// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package cmdutil

import "github.com/spf13/cobra"

// tipsAnnotation key — stored under cobra.Command.Annotations.
const tipsAnnotation = "jodoo-cli/tips"

// SetTips attaches a list of tip strings to a cobra command. The root
// command's help func appends them under a TIPS heading.
func SetTips(cmd *cobra.Command, tips []string) {
	if len(tips) == 0 {
		return
	}
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	out := tips[0]
	for _, t := range tips[1:] {
		out += "\n" + t
	}
	cmd.Annotations[tipsAnnotation] = out
}

// GetTips returns the tips attached via SetTips (empty if none).
func GetTips(cmd *cobra.Command) []string {
	if cmd == nil || cmd.Annotations == nil {
		return nil
	}
	v := cmd.Annotations[tipsAnnotation]
	if v == "" {
		return nil
	}
	var out []string
	cur := ""
	for _, r := range v {
		if r == '\n' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
