// Package clients defines the cross-client MCP configuration write-adapter layer.
//
// The core primitive decouples "how an MCP server is installed" (a ServerSpec,
// client-agnostic) from "where each client keeps its config and what shape it
// wants" (a Client adapter). A single ServerSpec can therefore be written into
// every detected client, each in that client's own schema.
package clients

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

// Format is the on-disk config format a client uses.
type Format int

const (
	// FormatJSON is a JSON config file with a top-level mcpServers object.
	FormatJSON Format = iota
	// FormatTOML is a TOML config file with [mcp_servers.<name>] tables.
	FormatTOML
)

func (f Format) String() string {
	switch f {
	case FormatTOML:
		return "toml"
	default:
		return "json"
	}
}

// ServerSpec is the client-agnostic intent to install one MCP server.
type ServerSpec struct {
	Name    string            // logical entry name, e.g. "filesystem"
	Command string            // executable, e.g. "npx"
	Args    []string          // arguments passed to Command
	Env     map[string]string // environment variables for the server process
}

// Client is a per-client config adapter: it resolves the client's config path
// on this machine and marshals a ServerSpec into that client's schema shape.
type Client interface {
	// ID is the stable client identifier, e.g. "claude-code".
	ID() string
	// DisplayName is the human-facing name, e.g. "Claude Code".
	DisplayName() string
	// Format reports whether the config is JSON or TOML.
	Format() Format
	// ConfigPath resolves the absolute config-file path for this machine.
	// It returns an error (with ok=false) only when the path genuinely
	// cannot be determined on this OS — callers treat that as "skip,
	// unverified", never as a reason to guess-write.
	ConfigPath() (path string, ok bool, err error)
	// EntryValue renders the JSON value for one server entry as raw JSON
	// bytes (for JSON clients). It is unused by TOML clients.
	EntryValue(spec ServerSpec) ([]byte, error)
}

// jsonEntry is the marshaled JSON shape shared by the JSON clients, minus the
// optional "type" field which some clients include and some omit.
type jsonEntry struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

func marshalJSONEntry(spec ServerSpec, includeType bool) ([]byte, error) {
	e := jsonEntry{
		Command: spec.Command,
		Args:    spec.Args,
		Env:     spec.Env,
	}
	if e.Args == nil {
		e.Args = []string{}
	}
	if e.Env == nil {
		e.Env = map[string]string{}
	}
	if includeType {
		e.Type = "stdio"
	}
	return json.Marshal(e)
}

func homeDir() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil || h == "" {
		return "", fmt.Errorf("cannot resolve home directory: %w", err)
	}
	return h, nil
}

// ---- claude-code ----------------------------------------------------------

type claudeCode struct{}

func (claudeCode) ID() string          { return "claude-code" }
func (claudeCode) DisplayName() string { return "Claude Code" }
func (claudeCode) Format() Format      { return FormatJSON }

func (claudeCode) ConfigPath() (string, bool, error) {
	h, err := homeDir()
	if err != nil {
		return "", false, err
	}
	return filepath.Join(h, ".claude.json"), true, nil
}

// EntryValue includes "type":"stdio" (verified on-machine shape).
func (claudeCode) EntryValue(spec ServerSpec) ([]byte, error) {
	return marshalJSONEntry(spec, true)
}

// ---- claude-desktop -------------------------------------------------------

type claudeDesktop struct{}

func (claudeDesktop) ID() string          { return "claude-desktop" }
func (claudeDesktop) DisplayName() string { return "Claude Desktop" }
func (claudeDesktop) Format() Format      { return FormatJSON }

func (claudeDesktop) ConfigPath() (string, bool, error) {
	h, err := homeDir()
	if err != nil {
		return "", false, err
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(h, "Library", "Application Support", "Claude", "claude_desktop_config.json"), true, nil
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "Claude", "claude_desktop_config.json"), true, nil
		}
		return filepath.Join(h, "AppData", "Roaming", "Claude", "claude_desktop_config.json"), true, nil
	case "linux":
		return filepath.Join(h, ".config", "Claude", "claude_desktop_config.json"), true, nil
	default:
		return "", false, fmt.Errorf("claude-desktop: unsupported OS %q", runtime.GOOS)
	}
}

// EntryValue omits "type" (verified on-machine shape: {command,args,env}).
func (claudeDesktop) EntryValue(spec ServerSpec) ([]byte, error) {
	return marshalJSONEntry(spec, false)
}

// ---- cursor (UNVERIFIED — fail-soft on absent config) ---------------------

type cursor struct{}

func (cursor) ID() string          { return "cursor" }
func (cursor) DisplayName() string { return "Cursor" }
func (cursor) Format() Format      { return FormatJSON }

func (cursor) ConfigPath() (string, bool, error) {
	h, err := homeDir()
	if err != nil {
		return "", false, err
	}
	return filepath.Join(h, ".cursor", "mcp.json"), true, nil
}

// EntryValue omits "type" (per plan: {command,args,env}).
func (cursor) EntryValue(spec ServerSpec) ([]byte, error) {
	return marshalJSONEntry(spec, false)
}

// ---- codex (TOML) ---------------------------------------------------------

type codex struct{}

func (codex) ID() string          { return "codex" }
func (codex) DisplayName() string { return "Codex" }
func (codex) Format() Format      { return FormatTOML }

func (codex) ConfigPath() (string, bool, error) {
	h, err := homeDir()
	if err != nil {
		return "", false, err
	}
	return filepath.Join(h, ".codex", "config.toml"), true, nil
}

// EntryValue is unused for TOML clients; the TOML writer builds the table
// surgically. Returning an error guards against accidental JSON-path use.
func (codex) EntryValue(ServerSpec) ([]byte, error) {
	return nil, fmt.Errorf("codex is a TOML client; use the TOML writer")
}

// registry holds every known client adapter, keyed by ID, in a stable order.
var registry = []Client{
	claudeCode{},
	claudeDesktop{},
	codex{},
	cursor{},
}

// All returns every known client adapter in stable registration order.
func All() []Client {
	out := make([]Client, len(registry))
	copy(out, registry)
	return out
}

// ByID returns the adapter with the given ID, or (nil, false).
func ByID(id string) (Client, bool) {
	for _, c := range registry {
		if c.ID() == id {
			return c, true
		}
	}
	return nil, false
}

// IDs returns the sorted list of known client IDs (for flag help / validation).
func IDs() []string {
	ids := make([]string, 0, len(registry))
	for _, c := range registry {
		ids = append(ids, c.ID())
	}
	sort.Strings(ids)
	return ids
}
