// Command mcpx is a one-command, multi-client MCP-server installer. It detects
// the coding-agent clients installed on this machine, backs up each client's
// config, idempotently merges one MCP server entry into every client, and runs
// a stdio handshake smoke test — reporting per-client OK / FAIL.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/SuperMarioYL/mcpx/internal/catalog"
	"github.com/SuperMarioYL/mcpx/internal/clients"
	"github.com/SuperMarioYL/mcpx/internal/config"
	"github.com/SuperMarioYL/mcpx/internal/handshake"
)

// version is stamped at build time via -ldflags; defaults to the VERSION file value.
var version = "0.5.0"

// global flags
var (
	flagClients []string
	flagDryRun  bool
	flagYes     bool
)

var (
	cGreen  = color.New(color.FgGreen).SprintFunc()
	cRed    = color.New(color.FgRed).SprintFunc()
	cYellow = color.New(color.FgYellow).SprintFunc()
	cDim    = color.New(color.Faint).SprintFunc()
	cBold   = color.New(color.Bold).SprintFunc()
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, cRed("error:"), err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "mcpx",
		Short: "One command to install an MCP server into every client you have",
		Long: "mcpx detects the MCP-capable clients installed on this machine " +
			"(Claude Code, Claude Desktop, Codex, Cursor), backs up each config, " +
			"idempotently writes one MCP server entry into every client, and runs a " +
			"handshake smoke test — no account, no cloud.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringSliceVar(&flagClients, "clients", nil,
		"restrict to a subset of clients (comma-separated ids: "+strings.Join(clients.IDs(), ",")+")")
	root.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "show what would change, write nothing")
	root.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "skip confirmation prompt")

	root.AddCommand(newAddCmd(), newListCmd(), newRemoveCmd(), newCatalogCmd())
	return root
}

// ---- add ------------------------------------------------------------------

func newAddCmd() *cobra.Command {
	var (
		command string
		argsStr string
		envKV   []string
	)
	cmd := &cobra.Command{
		Use:   "add <server>",
		Short: "Install an MCP server into every detected client",
		Args:  cobra.ExactArgs(1),
		Example: strings.Join([]string{
			"  mcpx add filesystem",
			"  mcpx add mytool --command npx --args '-y my-mcp-server' --env TOKEN=abc",
			"  mcpx add filesystem --clients claude-code,codex --dry-run",
		}, "\n"),
		RunE: func(cmd *cobra.Command, cargs []string) error {
			name := cargs[0]
			env, err := parseEnv(envKV)
			if err != nil {
				return err
			}
			spec, err := catalog.Resolve(name, command, splitArgs(argsStr), env)
			if err != nil {
				return err
			}
			return runAdd(cmd, spec)
		},
	}
	cmd.Flags().StringVar(&command, "command", "", "executable to launch (for a server not in the built-in catalog)")
	cmd.Flags().StringVar(&argsStr, "args", "", "server arguments, space-separated (e.g. '-y @scope/pkg .')")
	cmd.Flags().StringArrayVar(&envKV, "env", nil, "environment variable K=V (repeatable)")
	return cmd
}

func runAdd(cmd *cobra.Command, spec clients.ServerSpec) error {
	detected := clients.DetectSubset(flagClients)
	present := clients.Present(detected)

	printDetection(cmd, detected)
	if len(present) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), cYellow("No MCP-capable clients detected. Nothing to do."))
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nInstalling %s (%s %s) into %d client(s)%s\n",
		cBold(spec.Name), spec.Command, strings.Join(spec.Args, " "), len(present), dryRunTag())

	if !flagDryRun && !flagYes && !confirm(cmd, fmt.Sprintf("Write %q into %d client(s)?", spec.Name, len(present))) {
		fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
		return nil
	}

	ok, fail := 0, 0
	for _, dc := range present {
		res, err := config.Merge(dc.Adapter, spec, flagDryRun)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-16s %s %v\n", dc.Adapter.DisplayName(), cRed("write FAIL"), err)
			fail++
			continue
		}
		writeLine := "backup → write → merged"
		switch {
		case flagDryRun && res.Changed:
			writeLine = "would backup → write → merge"
		case flagDryRun && !res.Changed:
			writeLine = "already up to date (no change)"
		case !res.Changed:
			writeLine = "already up to date (idempotent no-op)"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %-16s %s\n", dc.Adapter.DisplayName(), cDim(writeLine))
		if res.BackupPath != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-16s %s\n", "", cDim("backup: "+res.BackupPath))
		}

		// Handshake (skipped in dry-run — nothing was written to test against).
		if flagDryRun {
			continue
		}
		hs := handshake.Test(spec)
		switch hs.Status {
		case handshake.StatusOK:
			fmt.Fprintf(cmd.OutOrStdout(), "  %-16s %s %s\n", "",
				cGreen("✓ handshake OK"), cDim(fmt.Sprintf("(initialize + tools/list, %d tools)", hs.ToolCount)))
			ok++
		case handshake.StatusWarn:
			fmt.Fprintf(cmd.OutOrStdout(), "  %-16s %s %s\n", "",
				cYellow("! handshake WARN"), cDim(hs.Reason+" — entry kept, not rolled back"))
			ok++
		default:
			fmt.Fprintf(cmd.OutOrStdout(), "  %-16s %s %s\n", "",
				cRed("✗ handshake FAIL"), cDim(hs.Reason))
			fail++
		}
	}

	if flagDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "\n"+cDim("dry-run: no files were written."))
		return nil
	}
	total := ok + fail
	summary := fmt.Sprintf("%d/%d clients ready", ok, total)
	if fail == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "\n%s %s\n", cGreen("✓"), cGreen(summary))
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "\n%s %s\n", cYellow("!"), cYellow(summary))
	}
	fmt.Fprintln(cmd.OutOrStdout(), cDim("Restart (or reload) the affected clients to pick up the new server."))
	return nil
}

// ---- list -----------------------------------------------------------------

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show detected clients and the MCP servers each has installed",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			detected := clients.DetectSubset(flagClients)
			present := clients.Present(detected)
			printDetection(cmd, detected)
			if len(present) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), cYellow("No MCP-capable clients detected."))
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout())
			for _, dc := range present {
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", cBold(dc.Adapter.DisplayName()), cDim("("+dc.ConfigPath+")"))
				servers, err := config.List(dc.Adapter)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s %v\n", cRed("read FAIL"), err)
					continue
				}
				if len(servers) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "  "+cDim("(no MCP servers)"))
					continue
				}
				for _, s := range servers {
					cmdline := strings.TrimSpace(s.Command + " " + strings.Join(s.Args, " "))
					fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", s.Name, cDim(cmdline))
				}
			}
			return nil
		},
	}
}

// ---- remove ---------------------------------------------------------------

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <server>",
		Aliases: []string{"rm"},
		Short:   "Remove an MCP server entry from every client (backup preserved)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, cargs []string) error {
			name := cargs[0]
			detected := clients.DetectSubset(flagClients)
			present := clients.Present(detected)
			printDetection(cmd, detected)
			if len(present) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), cYellow("No MCP-capable clients detected."))
				return nil
			}
			if !flagDryRun && !flagYes && !confirm(cmd, fmt.Sprintf("Remove %q from %d client(s)?", name, len(present))) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}
			removed := 0
			fmt.Fprintf(cmd.OutOrStdout(), "\nRemoving %s%s\n", cBold(name), dryRunTag())
			for _, dc := range present {
				res, err := config.Remove(dc.Adapter, name, flagDryRun)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  %-16s %s %v\n", dc.Adapter.DisplayName(), cRed("FAIL"), err)
					continue
				}
				switch {
				case !res.Changed:
					fmt.Fprintf(cmd.OutOrStdout(), "  %-16s %s\n", dc.Adapter.DisplayName(), cDim("not present (skipped)"))
				case flagDryRun:
					fmt.Fprintf(cmd.OutOrStdout(), "  %-16s %s\n", dc.Adapter.DisplayName(), cYellow("would remove"))
					removed++
				default:
					fmt.Fprintf(cmd.OutOrStdout(), "  %-16s %s\n", dc.Adapter.DisplayName(), cGreen("removed"))
					if res.BackupPath != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "  %-16s %s\n", "", cDim("backup: "+res.BackupPath))
					}
					removed++
				}
			}
			if flagDryRun {
				fmt.Fprintln(cmd.OutOrStdout(), "\n"+cDim("dry-run: no files were written."))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "\n%s removed from %d client(s)\n", cGreen("✓"), removed)
			}
			return nil
		},
	}
	return cmd
}

// ---- catalog --------------------------------------------------------------

func newCatalogCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "catalog",
		Short: "List the built-in MCP servers you can add by name",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), cBold("Built-in servers")+cDim(" (mcpx add <name>):"))
			for _, e := range catalog.List() {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", e.Name, cDim(e.Description))
				fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", "", cDim("→ "+e.Command+" "+strings.Join(e.Args, " ")))
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\n"+cDim("Anything else: mcpx add <name> --command <bin> --args '<a b c>' --env K=V"))
			return nil
		},
	}
}

// ---- helpers --------------------------------------------------------------

func printDetection(cmd *cobra.Command, detected []clients.DetectedClient) {
	var names []string
	for _, dc := range detected {
		if dc.Exists && !dc.Skipped {
			names = append(names, dc.Adapter.DisplayName())
		}
	}
	sort.Strings(names)
	if len(names) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Detected %d client(s): %s\n", len(names), strings.Join(names, ", "))
	}
	for _, dc := range detected {
		if dc.Skipped {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s %s %s\n", cDim("·"), dc.Adapter.DisplayName(),
				cDim("skipped ("+dc.SkipReason+")"))
		} else if !dc.Exists {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s %s %s\n", cDim("·"), dc.Adapter.DisplayName(),
				cDim("not detected (no config at "+dc.ConfigPath+")"))
		}
	}
}

func confirm(cmd *cobra.Command, prompt string) bool {
	fmt.Fprintf(cmd.OutOrStdout(), "%s [y/N] ", prompt)
	var resp string
	_, _ = fmt.Fscanln(cmd.InOrStdin(), &resp)
	resp = strings.ToLower(strings.TrimSpace(resp))
	return resp == "y" || resp == "yes"
}

func dryRunTag() string {
	if flagDryRun {
		return " " + cYellow("(dry-run)")
	}
	return ""
}

// splitArgs splits a space-separated args string, honoring simple quoting so
// an argument containing spaces can be given as 'a "b c" d'. A comma is an
// ordinary character: a single token that contains a comma but no whitespace
// (e.g. a URL query http://x/api?a=1,b=2) stays one arg instead of being
// mis-split at the comma. Quote the token if it also needs to contain spaces.
func splitArgs(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return fields(s)
}

// fields splits on whitespace while respecting single/double quotes.
func fields(s string) []string {
	var out []string
	var cur strings.Builder
	var quote rune
	inField := false
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
			inField = true
		case r == '\'' || r == '"':
			quote = r
			inField = true
		case r == ' ' || r == '\t':
			if inField {
				out = append(out, cur.String())
				cur.Reset()
				inField = false
			}
		default:
			cur.WriteRune(r)
			inField = true
		}
	}
	if inField {
		out = append(out, cur.String())
	}
	return out
}

func parseEnv(kv []string) (map[string]string, error) {
	if len(kv) == 0 {
		return nil, nil
	}
	env := make(map[string]string, len(kv))
	for _, pair := range kv {
		i := strings.IndexByte(pair, '=')
		if i <= 0 {
			return nil, fmt.Errorf("invalid --env %q: expected K=V", pair)
		}
		env[pair[:i]] = pair[i+1:]
	}
	return env, nil
}
