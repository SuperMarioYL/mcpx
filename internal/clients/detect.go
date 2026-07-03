package clients

import (
	"os"
	"sort"
)

// DetectedClient is a client whose config path was resolved on this machine,
// together with whether its config file currently exists.
type DetectedClient struct {
	Adapter    Client
	ConfigPath string
	Exists     bool // config file is present on disk
	// Skipped is true when the adapter could not resolve a config path on
	// this OS (unverified / unsupported). Such clients are reported but
	// never written to — mcpx must never guess-write.
	Skipped bool
	// SkipReason explains a Skipped client for user-facing output.
	SkipReason string
}

// Detect resolves and probes every known client adapter. A client is reported
// as detected (Exists=true) iff its config file is present on disk. Clients
// whose config path cannot be resolved on this OS are returned with
// Skipped=true and are never candidates for writing.
//
// Detect is conservative: a missing config file means "client not detected"
// (Exists=false), never a signal to create-and-guess.
func Detect() []DetectedClient {
	return detect(All())
}

// DetectSubset behaves like Detect but restricts to the given client IDs. An
// unknown ID is ignored. Passing an empty slice returns all clients.
func DetectSubset(ids []string) []DetectedClient {
	if len(ids) == 0 {
		return Detect()
	}
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	var sel []Client
	for _, c := range All() {
		if want[c.ID()] {
			sel = append(sel, c)
		}
	}
	return detect(sel)
}

func detect(adapters []Client) []DetectedClient {
	out := make([]DetectedClient, 0, len(adapters))
	for _, a := range adapters {
		path, ok, err := a.ConfigPath()
		if !ok || err != nil {
			reason := "config path could not be resolved on this machine"
			if err != nil {
				reason = err.Error()
			}
			out = append(out, DetectedClient{
				Adapter:    a,
				Skipped:    true,
				SkipReason: reason,
			})
			continue
		}
		out = append(out, DetectedClient{
			Adapter:    a,
			ConfigPath: path,
			Exists:     fileExists(path),
		})
	}
	// Stable, deterministic order for output.
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Adapter.ID() < out[j].Adapter.ID()
	})
	return out
}

// Present returns only the clients whose config file exists on disk.
func Present(all []DetectedClient) []DetectedClient {
	var p []DetectedClient
	for _, dc := range all {
		if dc.Exists && !dc.Skipped {
			p = append(p, dc)
		}
	}
	return p
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
