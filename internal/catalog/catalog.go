// Package catalog resolves a server name (or an explicit --command/--args
// from-spec) into a client-agnostic ServerSpec.
package catalog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/SuperMarioYL/mcpx/internal/clients"
)

// Entry is a catalog record for a well-known MCP server.
type Entry struct {
	Name        string
	Description string
	Command     string
	Args        []string
}

// builtin is a small curated catalog of common MCP servers. These are launch
// commands only — mcpx does NOT install the underlying binaries (npx/uvx pull
// them on first run, per the out-of-scope contract).
var builtin = map[string]Entry{
	"filesystem": {
		Name:        "filesystem",
		Description: "Local filesystem access (read/write files under given roots)",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
	},
	"git": {
		Name:        "git",
		Description: "Git repository operations (log, diff, status, blame)",
		Command:     "uvx",
		Args:        []string{"mcp-server-git"},
	},
	"fetch": {
		Name:        "fetch",
		Description: "Fetch and convert web pages to markdown",
		Command:     "uvx",
		Args:        []string{"mcp-server-fetch"},
	},
	"sqlite": {
		Name:        "sqlite",
		Description: "Query a local SQLite database",
		Command:     "uvx",
		Args:        []string{"mcp-server-sqlite", "--db-path", "./data.db"},
	},
	"memory": {
		Name:        "memory",
		Description: "Knowledge-graph memory store",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-memory"},
	},
	"time": {
		Name:        "time",
		Description: "Time and timezone conversion utilities",
		Command:     "uvx",
		Args:        []string{"mcp-server-time"},
	},
	"sequential-thinking": {
		Name:        "sequential-thinking",
		Description: "Structured step-by-step reasoning scratchpad",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
	},
	"everything": {
		Name:        "everything",
		Description: "Reference server exercising all MCP features (for testing)",
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-everything"},
	},
}

// FromSpec builds a ServerSpec from an explicit command/args/env. Used by the
// generic `mcpx add <name> --command <bin> --args '<a b c>' --env K=V` path.
// name is the logical entry name; command must be non-empty.
func FromSpec(name, command string, args []string, env map[string]string) (clients.ServerSpec, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return clients.ServerSpec{}, fmt.Errorf("server name is required")
	}
	if strings.TrimSpace(command) == "" {
		return clients.ServerSpec{}, fmt.Errorf("--command is required when %q is not a built-in catalog server", name)
	}
	return clients.ServerSpec{
		Name:    name,
		Command: command,
		Args:    args,
		Env:     env,
	}, nil
}

// Resolve turns a server name into a ServerSpec. If command is non-empty the
// from-spec path is used (overriding any catalog entry of the same name);
// otherwise the built-in catalog is consulted. env is merged on top of any
// catalog defaults (currently the catalog carries no env).
func Resolve(name, command string, args []string, env map[string]string) (clients.ServerSpec, error) {
	if strings.TrimSpace(command) != "" {
		return FromSpec(name, command, args, env)
	}
	e, ok := builtin[name]
	if !ok {
		return clients.ServerSpec{}, fmt.Errorf(
			"unknown server %q: pass --command/--args to install it from spec, or choose a built-in (%s)",
			name, strings.Join(Names(), ", "))
	}
	spec := clients.ServerSpec{
		Name:    e.Name,
		Command: e.Command,
		Args:    append([]string(nil), e.Args...),
		Env:     map[string]string{},
	}
	for k, v := range env {
		spec.Env[k] = v
	}
	return spec, nil
}

// Names returns the sorted list of built-in catalog server names.
func Names() []string {
	names := make([]string, 0, len(builtin))
	for n := range builtin {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// List returns the built-in catalog entries sorted by name.
func List() []Entry {
	out := make([]Entry, 0, len(builtin))
	for _, n := range Names() {
		out = append(out, builtin[n])
	}
	return out
}

// Lookup returns the built-in entry for name, if present.
func Lookup(name string) (Entry, bool) {
	e, ok := builtin[name]
	return e, ok
}
