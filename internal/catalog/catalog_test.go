package catalog

import (
	"strings"
	"testing"
)

// Regression for fix-catalog-args-silent-drop: --args with a built-in catalog
// name and no --command used to be silently dropped, so `mcpx add filesystem
// --args "... /tmp"` ran with the default catalog args. It must now error.
func TestResolveRejectsArgsWithoutCommandOnCatalogPath(t *testing.T) {
	_, err := Resolve("filesystem", "", []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"}, nil)
	if err == nil {
		t.Fatal("expected an error when --args is passed with a catalog name and no --command")
	}
	if !strings.Contains(err.Error(), "--args") || !strings.Contains(err.Error(), "--command") {
		t.Fatalf("error should point at the --command/--args remedy, got: %v", err)
	}
}

func TestResolveCatalogPathStillWorks(t *testing.T) {
	spec, err := Resolve("filesystem", "", nil, nil)
	if err != nil {
		t.Fatalf("plain catalog add must keep working: %v", err)
	}
	if spec.Command != "npx" || len(spec.Args) == 0 {
		t.Fatalf("unexpected catalog spec: %+v", spec)
	}
}

func TestResolveFromSpecPathStillWorks(t *testing.T) {
	spec, err := Resolve("mytool", "npx", []string{"-y", "@scope/pkg", "."}, nil)
	if err != nil {
		t.Fatalf("from-spec add must keep working: %v", err)
	}
	if spec.Command != "npx" || spec.Args[0] != "-y" {
		t.Fatalf("unexpected from-spec spec: %+v", spec)
	}
}

func TestResolveUnknownServerStillErrors(t *testing.T) {
	_, err := Resolve("nope", "", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "unknown server") {
		t.Fatalf("expected unknown-server error, got: %v", err)
	}
}
