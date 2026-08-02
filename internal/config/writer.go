// Package config performs backup-before-write, idempotent merge, listing, and
// removal of a single MCP server entry inside a client config file — without
// clobbering unrelated keys. JSON clients use sjson/gjson (key-path merge);
// the Codex TOML client uses go-toml/v2 for a surgical table merge.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/SuperMarioYL/mcpx/internal/clients"
)

// jsonServersKey is the top-level object holding MCP server entries in every
// JSON client (Claude Code, Claude Desktop, Cursor).
const jsonServersKey = "mcpServers"

// tomlServersKey is the top-level TOML table holding MCP server entries in
// Codex (github.com/pelletier/go-toml/v2 dotted path root).
const tomlServersKey = "mcp_servers"

// ServerInfo is one installed MCP server as read back from a config file.
type ServerInfo struct {
	Name    string
	Command string
	Args    []string
}

// Result reports the outcome of a single Merge/Remove operation.
type Result struct {
	ConfigPath string // config file that was operated on
	BackupPath string // backup created before mutation ("" if none / dry-run)
	Changed    bool   // true if the file content actually changed
	Created    bool   // true if the config file was created fresh
}

// timestamp used for backup suffixes; overridable in tests.
var nowFunc = time.Now

// backupName returns a sibling backup path for path that does not yet exist.
// The timestamp carries millisecond resolution so two ops in the same
// wall-clock second don't collide; if a collision still occurs (same
// millisecond), a -1, -2, ... suffix is appended until an unused name is found,
// so scripted back-to-back ops never clobber the prior backup.
func backupName(path string) string {
	base := fmt.Sprintf("%s.mcpx.bak.%s", path, nowFunc().Format("20060102-150405.000"))
	bp := base
	for i := 1; ; i++ {
		if _, err := os.Stat(bp); err != nil {
			return bp
		}
		bp = fmt.Sprintf("%s-%d", base, i)
	}
}

// backup copies path to a timestamped sibling and returns the backup path. It
// is a no-op (returns "") when the file does not yet exist.
func backup(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read for backup: %w", err)
	}
	bp := backupName(path)
	if err := os.WriteFile(bp, data, resolvePerm(path)); err != nil {
		return "", fmt.Errorf("write backup: %w", err)
	}
	return bp, nil
}

// Merge idempotently writes one MCP server entry into a client's config file,
// backing the file up first. Writing the same server twice is a byte-stable
// no-op (Changed=false, no backup). Unrelated top-level keys and other server
// entries are preserved. When dryRun is true, nothing is written and no backup
// is made, but Changed reflects whether a write WOULD have happened.
func Merge(c clients.Client, spec clients.ServerSpec, dryRun bool) (Result, error) {
	path, ok, err := c.ConfigPath()
	if err != nil || !ok {
		return Result{}, fmt.Errorf("resolve config path for %s: %w", c.ID(), err)
	}
	switch c.Format() {
	case clients.FormatTOML:
		return mergeTOML(c, path, spec, dryRun)
	default:
		return mergeJSON(c, path, spec, dryRun)
	}
}

// Remove deletes the named server entry from a client's config, backing up
// first. It leaves all other entries and unrelated keys intact. Removing an
// absent entry is a no-op.
func Remove(c clients.Client, name string, dryRun bool) (Result, error) {
	path, ok, err := c.ConfigPath()
	if err != nil || !ok {
		return Result{}, fmt.Errorf("resolve config path for %s: %w", c.ID(), err)
	}
	switch c.Format() {
	case clients.FormatTOML:
		return removeTOML(path, name, dryRun)
	default:
		return removeJSON(path, name, dryRun)
	}
}

// List reads the MCP servers currently present in a client's config.
func List(c clients.Client) ([]ServerInfo, error) {
	path, ok, err := c.ConfigPath()
	if err != nil || !ok {
		return nil, fmt.Errorf("resolve config path for %s: %w", c.ID(), err)
	}
	switch c.Format() {
	case clients.FormatTOML:
		return listTOML(path)
	default:
		return listJSON(path)
	}
}

// ---- JSON path (sjson / gjson) --------------------------------------------

func readJSON(path string) ([]byte, bool, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return []byte("{}"), true, nil
	}
	if !gjson.ValidBytes(data) {
		return nil, true, fmt.Errorf("%s is not valid JSON", path)
	}
	return data, true, nil
}

func mergeJSON(c clients.Client, path string, spec clients.ServerSpec, dryRun bool) (Result, error) {
	existing, existed, err := readJSON(path)
	if err != nil {
		return Result{}, err
	}
	if !existed {
		existing = []byte("{}")
	}

	entry, err := c.EntryValue(spec)
	if err != nil {
		return Result{}, fmt.Errorf("marshal entry for %s: %w", c.ID(), err)
	}

	keyPath := jsonServersKey + "." + escapeKey(spec.Name)
	updated, err := sjson.SetRawBytes(existing, keyPath, entry)
	if err != nil {
		return Result{}, fmt.Errorf("merge entry into %s: %w", path, err)
	}

	// Idempotency: compare canonicalized bytes so key ordering / whitespace
	// differences do not count as a change.
	changed := !jsonEqual(existing, updated)
	res := Result{ConfigPath: path, Changed: changed, Created: !existed}
	if !changed || dryRun {
		return res, nil
	}

	bp, err := backup(path)
	if err != nil {
		return Result{}, err
	}
	res.BackupPath = bp
	if err := writeFileAtomic(path, updated, resolvePerm(path)); err != nil {
		return Result{}, err
	}
	return res, nil
}

func removeJSON(path, name string, dryRun bool) (Result, error) {
	existing, existed, err := readJSON(path)
	if err != nil {
		return Result{}, err
	}
	if !existed {
		return Result{ConfigPath: path}, nil
	}
	keyPath := jsonServersKey + "." + escapeKey(name)
	if !gjson.GetBytes(existing, keyPath).Exists() {
		return Result{ConfigPath: path, Changed: false}, nil
	}
	updated, err := sjson.DeleteBytes(existing, keyPath)
	if err != nil {
		return Result{}, fmt.Errorf("delete entry from %s: %w", path, err)
	}
	res := Result{ConfigPath: path, Changed: true}
	if dryRun {
		return res, nil
	}
	bp, err := backup(path)
	if err != nil {
		return Result{}, err
	}
	res.BackupPath = bp
	if err := writeFileAtomic(path, updated, resolvePerm(path)); err != nil {
		return Result{}, err
	}
	return res, nil
}

func listJSON(path string) ([]ServerInfo, error) {
	existing, existed, err := readJSON(path)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}
	servers := gjson.GetBytes(existing, jsonServersKey)
	if !servers.Exists() || !servers.IsObject() {
		return nil, nil
	}
	var out []ServerInfo
	servers.ForEach(func(key, val gjson.Result) bool {
		si := ServerInfo{Name: key.String(), Command: val.Get("command").String()}
		val.Get("args").ForEach(func(_, a gjson.Result) bool {
			si.Args = append(si.Args, a.String())
			return true
		})
		out = append(out, si)
		return true
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// escapeKey escapes a server name for use as an sjson/gjson path segment
// (dots and wildcards would otherwise be interpreted as path syntax).
func escapeKey(name string) string {
	out := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		switch name[i] {
		case '.', '*', '?', '\\', '|', '#', '@':
			out = append(out, '\\')
		}
		out = append(out, name[i])
	}
	return string(out)
}

// jsonEqual reports whether two JSON documents are semantically equal
// (independent of key order and formatting).
func jsonEqual(a, b []byte) bool {
	var va, vb any
	if err := json.Unmarshal(a, &va); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &vb); err != nil {
		return false
	}
	na, err := json.Marshal(canonical(va))
	if err != nil {
		return false
	}
	nb, err := json.Marshal(canonical(vb))
	if err != nil {
		return false
	}
	return string(na) == string(nb)
}

// canonical recursively sorts map keys so marshaling is order-independent.
func canonical(v any) any {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		ordered := make([]kv, 0, len(t))
		for _, k := range keys {
			ordered = append(ordered, kv{k, canonical(t[k])})
		}
		return ordered
	case []any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = canonical(t[i])
		}
		return out
	default:
		return v
	}
}

type kv struct {
	K string
	V any
}

func (p kv) MarshalJSON() ([]byte, error) {
	kb, err := json.Marshal(p.K)
	if err != nil {
		return nil, err
	}
	vb, err := json.Marshal(p.V)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, len(kb)+len(vb)+2)
	buf = append(buf, '{')
	buf = append(buf, kb...)
	buf = append(buf, ':')
	buf = append(buf, vb...)
	buf = append(buf, '}')
	return buf, nil
}

// ---- TOML path (go-toml/v2) -----------------------------------------------

// tomlEntry mirrors a [mcp_servers.<name>] table.
type tomlEntry struct {
	Command string            `toml:"command"`
	Args    []string          `toml:"args"`
	Env     map[string]string `toml:"env,omitempty"`
}

func readTOML(path string) (map[string]any, bool, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]any{}, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read %s: %w", path, err)
	}
	root := map[string]any{}
	if len(data) > 0 {
		if err := toml.Unmarshal(data, &root); err != nil {
			return nil, true, fmt.Errorf("%s is not valid TOML: %w", path, err)
		}
	}
	return root, true, nil
}

func mergeTOML(c clients.Client, path string, spec clients.ServerSpec, dryRun bool) (Result, error) {
	root, existed, err := readTOML(path)
	if err != nil {
		return Result{}, err
	}

	servers, _ := root[tomlServersKey].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}

	newEntry := map[string]any{
		"command": spec.Command,
		"args":    stringSlice(spec.Args),
	}
	if len(spec.Env) > 0 {
		newEntry["env"] = stringMap(spec.Env)
	}

	// Idempotency check against the current table.
	if cur, ok := servers[spec.Name].(map[string]any); ok && tomlEntryEqual(cur, newEntry) {
		return Result{ConfigPath: path, Changed: false, Created: !existed}, nil
	}

	servers[spec.Name] = newEntry
	root[tomlServersKey] = servers

	out, err := toml.Marshal(root)
	if err != nil {
		return Result{}, fmt.Errorf("marshal TOML for %s: %w", path, err)
	}
	res := Result{ConfigPath: path, Changed: true, Created: !existed}
	if dryRun {
		return res, nil
	}
	bp, err := backup(path)
	if err != nil {
		return Result{}, err
	}
	res.BackupPath = bp
	if err := writeFileAtomic(path, out, resolvePerm(path)); err != nil {
		return Result{}, err
	}
	return res, nil
}

func removeTOML(path, name string, dryRun bool) (Result, error) {
	root, existed, err := readTOML(path)
	if err != nil {
		return Result{}, err
	}
	if !existed {
		return Result{ConfigPath: path}, nil
	}
	servers, _ := root[tomlServersKey].(map[string]any)
	if servers == nil {
		return Result{ConfigPath: path, Changed: false}, nil
	}
	if _, ok := servers[name]; !ok {
		return Result{ConfigPath: path, Changed: false}, nil
	}
	delete(servers, name)
	if len(servers) == 0 {
		delete(root, tomlServersKey)
	} else {
		root[tomlServersKey] = servers
	}
	out, err := toml.Marshal(root)
	if err != nil {
		return Result{}, fmt.Errorf("marshal TOML for %s: %w", path, err)
	}
	res := Result{ConfigPath: path, Changed: true}
	if dryRun {
		return res, nil
	}
	bp, err := backup(path)
	if err != nil {
		return Result{}, err
	}
	res.BackupPath = bp
	if err := writeFileAtomic(path, out, resolvePerm(path)); err != nil {
		return Result{}, err
	}
	return res, nil
}

func listTOML(path string) ([]ServerInfo, error) {
	root, existed, err := readTOML(path)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}
	servers, _ := root[tomlServersKey].(map[string]any)
	if servers == nil {
		return nil, nil
	}
	var out []ServerInfo
	for name, raw := range servers {
		entry, _ := raw.(map[string]any)
		si := ServerInfo{Name: name}
		if entry != nil {
			si.Command, _ = entry["command"].(string)
			if args, ok := entry["args"].([]any); ok {
				for _, a := range args {
					if s, ok := a.(string); ok {
						si.Args = append(si.Args, s)
					}
				}
			}
		}
		out = append(out, si)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func tomlEntryEqual(a, b map[string]any) bool {
	ba, err := toml.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := toml.Marshal(b)
	if err != nil {
		return false
	}
	return string(ba) == string(bb)
}

func stringSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func stringMap(m map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// resolvePerm returns the permission bits of the existing file at path, or
// 0o600 when the file is new or stat fails. Existing config files keep their
// current mode so a hardened 0600 ~/.claude.json (which holds MCP env secrets)
// is never silently widened to 0644 by an add/remove; newly-created configs
// default to 0600 rather than the world-readable 0644.
func resolvePerm(path string) os.FileMode {
	if fi, err := os.Stat(path); err == nil {
		return fi.Mode().Perm()
	}
	return 0o600
}

// writeFileAtomic writes data to a temp file in the same dir then renames it
// over path, so a crash mid-write never leaves a truncated config.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".mcpx-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return fmt.Errorf("chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp over %s: %w", path, err)
	}
	return nil
}
