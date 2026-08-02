// Package handshake runs a stdio JSON-RPC smoke test against an MCP server:
// it spawns the server process, sends an `initialize` request and a
// `tools/list` request, and reports OK only if both round-trip within a
// timeout. It uses os/exec + encoding/json directly (no MCP SDK) so the binary
// stays dependency-free and is not coupled to an SDK version.
package handshake

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/SuperMarioYL/mcpx/internal/clients"
)

// DefaultTimeout bounds the whole handshake so a wedged server never hangs the CLI.
const DefaultTimeout = 8 * time.Second

// Status is the verdict of a handshake.
type Status int

const (
	// StatusOK means both initialize and tools/list round-tripped.
	StatusOK Status = iota
	// StatusFail means the handshake did not complete (spawn error, bad
	// response, or protocol error).
	StatusFail
	// StatusWarn means the write succeeded but the handshake timed out; the
	// entry is kept (not rolled back) per the fail-soft contract.
	StatusWarn
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusWarn:
		return "WARN"
	default:
		return "FAIL"
	}
}

// Result reports a handshake outcome.
type Result struct {
	Status    Status
	Reason    string        // failure/warn explanation ("" on OK)
	ToolCount int           // number of tools reported by tools/list
	Elapsed   time.Duration // wall-clock time taken
}

// jsonrpcRequest is a JSON-RPC 2.0 request frame.
type jsonrpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// jsonrpcResponse is the subset of a JSON-RPC 2.0 response we care about.
type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *jsonrpcError   `json:"error"`
	Method  string          `json:"method"` // set on notifications (ignored)
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Test runs the handshake against the server described by spec, using
// DefaultTimeout. It never blocks longer than the timeout.
func Test(spec clients.ServerSpec) Result {
	return TestContext(context.Background(), spec, DefaultTimeout)
}

// TestContext runs the handshake with an explicit parent context and timeout.
func TestContext(parent context.Context, spec clients.ServerSpec, timeout time.Duration) Result {
	start := time.Now()
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	if strings.TrimSpace(spec.Command) == "" {
		return Result{Status: StatusFail, Reason: "no command to launch", Elapsed: time.Since(start)}
	}

	cmd := exec.CommandContext(ctx, spec.Command, spec.Args...)
	cmd.Env = envSlice(spec.Env)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return Result{Status: StatusFail, Reason: "stdin pipe: " + err.Error(), Elapsed: time.Since(start)}
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{Status: StatusFail, Reason: "stdout pipe: " + err.Error(), Elapsed: time.Since(start)}
	}
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return Result{Status: StatusFail, Reason: "spawn failed: " + err.Error(), Elapsed: time.Since(start)}
	}
	// Ensure the process is reaped and killed on any exit path.
	defer func() {
		_ = stdin.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	}()

	// One persistent reader goroutine feeds a shared channel for the whole
	// session, so a response frame read while waiting for the initialize
	// reply is never lost before the tools/list wait.
	frames := make(chan frame, 8)
	go readLoop(bufio.NewReader(stdout), frames)

	result, err := run(ctx, stdin, frames)
	result.Elapsed = time.Since(start)
	if err != nil {
		// Distinguish a timeout (WARN, fail-soft) from a hard protocol error.
		if ctx.Err() == context.DeadlineExceeded {
			result.Status = StatusWarn
			result.Reason = "handshake timed out after " + timeout.String()
			return result
		}
		result.Status = StatusFail
		if result.Reason == "" {
			result.Reason = err.Error()
		}
		return result
	}
	return result
}

func run(ctx context.Context, stdin io.Writer, frames <-chan frame) (Result, error) {
	// 1) initialize
	initParams := map[string]any{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "mcpx", "version": "0.1.0"},
	}
	if err := writeRequest(stdin, jsonrpcRequest{JSONRPC: "2.0", ID: 1, Method: "initialize", Params: initParams}); err != nil {
		return Result{}, fmt.Errorf("send initialize: %w", err)
	}
	if _, err := readResponse(ctx, frames, 1); err != nil {
		return Result{}, fmt.Errorf("initialize: %w", err)
	}

	// initialized notification (best-effort; some servers require it)
	_ = writeRaw(stdin, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})

	// 2) tools/list
	if err := writeRequest(stdin, jsonrpcRequest{JSONRPC: "2.0", ID: 2, Method: "tools/list", Params: map[string]any{}}); err != nil {
		return Result{}, fmt.Errorf("send tools/list: %w", err)
	}
	resp, err := readResponse(ctx, frames, 2)
	if err != nil {
		return Result{}, fmt.Errorf("tools/list: %w", err)
	}

	count := 0
	if len(resp.Result) > 0 {
		var tl struct {
			Tools []json.RawMessage `json:"tools"`
		}
		if err := json.Unmarshal(resp.Result, &tl); err == nil {
			count = len(tl.Tools)
		}
	}
	return Result{Status: StatusOK, ToolCount: count}, nil
}

func writeRequest(w io.Writer, req jsonrpcRequest) error {
	return writeRaw(w, req)
}

func writeRaw(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

// frame is one line read from the server's stdout.
type frame struct {
	line []byte
	err  error
}

// readLoop reads newline-delimited frames from the server's stdout and feeds
// them to frames until an error/EOF, which it forwards and then stops.
func readLoop(reader *bufio.Reader, frames chan<- frame) {
	for {
		line, err := reader.ReadBytes('\n')
		frames <- frame{line: line, err: err}
		if err != nil {
			return
		}
	}
}

// readResponse consumes frames until it finds a JSON-RPC response matching
// wantID, skipping notifications and mismatched IDs. It respects ctx
// cancellation (timeout) so it never blocks past the deadline.
func readResponse(ctx context.Context, frames <-chan frame, wantID int) (jsonrpcResponse, error) {
	for {
		select {
		case <-ctx.Done():
			return jsonrpcResponse{}, ctx.Err()
		case fr, ok := <-frames:
			if !ok {
				return jsonrpcResponse{}, fmt.Errorf("server closed stdout before responding to id %d", wantID)
			}
			if len(fr.line) > 0 {
				trimmed := strings.TrimSpace(string(fr.line))
				if trimmed != "" {
					var resp jsonrpcResponse
					if err := json.Unmarshal([]byte(trimmed), &resp); err == nil {
						if resp.ID != nil && *resp.ID == wantID {
							if resp.Error != nil {
								return resp, fmt.Errorf("server error %d: %s", resp.Error.Code, resp.Error.Message)
							}
							return resp, nil
						}
						// notification or a different id — keep reading
					}
				}
			}
			if fr.err != nil {
				if fr.err == io.EOF {
					return jsonrpcResponse{}, fmt.Errorf("server closed stdout before responding to id %d", wantID)
				}
				return jsonrpcResponse{}, fr.err
			}
		}
	}
}

// envSlice builds the child process environment by layering the user-supplied
// vars over the parent environment (os.Environ). Go's os/exec replaces (rather
// than extends) the parent env when cmd.Env is non-nil, so without this the
// spawned server would receive ONLY the --env vars — losing PATH/HOME, which
// means npx/uvx catalog servers can't locate node/python and exit immediately,
// readLoop then sees EOF and the handshake reports a false FAIL even though the
// config write succeeded. When env is empty we return nil so cmd.Env stays nil
// and the parent env is inherited untouched (catalog-only adds are unaffected).
func envSlice(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	parent := os.Environ()
	out := make([]string, 0, len(parent)+len(env))
	applied := make(map[string]struct{}, len(env))
	for _, kv := range parent {
		if key, _, found := strings.Cut(kv, "="); found {
			if override, hasUser := env[key]; hasUser {
				out = append(out, key+"="+override)
				applied[key] = struct{}{}
				continue
			}
		}
		out = append(out, kv)
	}
	for k, v := range env {
		if _, ok := applied[k]; !ok {
			out = append(out, k+"="+v)
		}
	}
	return out
}
