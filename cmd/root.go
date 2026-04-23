// Copyright (c) 2026 Jodoo CLI Authors
// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"jodoo-cli/cmd/api"
	"jodoo-cli/cmd/auth"
	cmdconfig "jodoo-cli/cmd/config"
	"jodoo-cli/cmd/doctor"
	jodoomcp "jodoo-cli/cmd/mcp"
	"jodoo-cli/cmd/profile"
	"jodoo-cli/internal/build"
	"jodoo-cli/internal/cmdutil"
	"jodoo-cli/internal/core"
	"jodoo-cli/internal/output"
	"jodoo-cli/shortcuts"
)

const rootLong = `jodoo-cli — Jodoo (api.jodoo.com) CLI tool.

USAGE:
    jodoo-cli <command> [subcommand] [method] [options]
    jodoo-cli api <path> [--data <json>] [--dry-run]
    jodoo-cli jodoo +<shortcut> [flags]

EXAMPLES:
    # List apps reachable by your API key
    jodoo-cli jodoo +app-list

    # Inspect form fields
    jodoo-cli jodoo +widget-list --app-id <a> --entry-id <e>

    # Filter records
    jodoo-cli jodoo +data-list --app-id <a> --entry-id <e> \
        --filter '{"rel":"and","cond":[{"field":"_widget_xxx","type":"text","method":"eq","value":["foo"]}]}'

    # Raw API escape hatch (every Jodoo endpoint is a POST)
    jodoo-cli api /v5/app/list --data '{"limit":100}'

FLAGS:
    --profile <name>     config profile (overrides JODOO_PROFILE)
    --data <json>        request body JSON (raw api command)
    --format <fmt>       output format: json (default) | pretty | table | ndjson | csv
    --jq <expr>          jq expression to filter JSON output
    -q <expr>            shorthand for --jq
    --dry-run            print request without executing

AI AGENT SKILLS:
    jodoo-cli ships with AI agent skills (Claude Code, etc.):
        jodoo-cli, jodoo-shared

DOCS:
    https://api.jodoo.com (account required)

More help: jodoo-cli <command> --help`

// Execute runs the root command and returns the process exit code.
func Execute() int {
	inv, err := BootstrapInvocationContext(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	f := cmdutil.NewDefault(inv)

	globals := &GlobalOptions{Profile: inv.Profile}
	rootCmd := &cobra.Command{
		Use:     "jodoo-cli",
		Short:   "Jodoo CLI — apps, forms, records, files, workflows, contacts",
		Long:    rootLong,
		Version: build.Version,
	}
	installTipsHelpFunc(rootCmd)
	rootCmd.SilenceErrors = true

	RegisterGlobalFlags(rootCmd.PersistentFlags(), globals)
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
	}

	rootCmd.AddCommand(cmdconfig.NewCmdConfig(f))
	rootCmd.AddCommand(auth.NewCmdAuth(f))
	rootCmd.AddCommand(profile.NewCmdProfile(f))
	rootCmd.AddCommand(doctor.NewCmdDoctor(f))
	rootCmd.AddCommand(api.NewCmdApi(f))
	rootCmd.AddCommand(jodoomcp.NewCmdMcp(f))
	shortcuts.RegisterShortcuts(rootCmd, f)

	if err := rootCmd.Execute(); err != nil {
		return handleRootError(f, err)
	}
	return 0
}

func handleRootError(f *cmdutil.Factory, err error) int {
	errOut := f.IOStreams.ErrOut

	if exitErr := asExitError(err); exitErr != nil {
		output.WriteErrorEnvelope(errOut, exitErr)
		return exitErr.Code
	}
	fmt.Fprintln(errOut, "Error:", err)
	return 1
}

func asExitError(err error) *output.ExitError {
	var cfgErr *core.ConfigError
	if errors.As(err, &cfgErr) {
		return output.ErrWithHint(cfgErr.Code, cfgErr.Type, cfgErr.Message, cfgErr.Hint)
	}
	var exitErr *output.ExitError
	if errors.As(err, &exitErr) {
		return exitErr
	}
	return nil
}

func installTipsHelpFunc(root *cobra.Command) {
	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelp(cmd, args)
		tips := cmdutil.GetTips(cmd)
		if len(tips) == 0 {
			return
		}
		out := cmd.OutOrStdout()
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Tips:")
		for _, tip := range tips {
			fmt.Fprintf(out, "    • %s\n", tip)
		}
	})
}
