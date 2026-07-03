// Command stubserver is a minimal stdio JSON-RPC MCP server used only by
// handshake tests. Its behavior is selected by the MCPX_STUB_MODE env var:
//
//	ok    (default) — answers initialize + tools/list correctly
//	error           — answers initialize, then returns a JSON-RPC error to tools/list
//	silent          — reads requests but never replies (drives the timeout/WARN path)
//	crash           — exits immediately without replying
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      *int   `json:"id"`
	Method  string `json:"method"`
}

func main() {
	mode := os.Getenv("MCPX_STUB_MODE")
	if mode == "crash" {
		os.Exit(1)
	}

	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	for {
		line, err := in.ReadBytes('\n')
		if len(line) > 0 {
			var req request
			if json.Unmarshal(line, &req) == nil {
				handle(out, req, mode)
			}
		}
		if err != nil {
			return
		}
	}
}

func handle(out *bufio.Writer, req request, mode string) {
	if req.ID == nil { // notification — no reply
		return
	}
	if mode == "silent" {
		return
	}
	switch req.Method {
	case "initialize":
		reply(out, *req.ID, `{"protocolVersion":"2025-06-18","capabilities":{"tools":{}},"serverInfo":{"name":"stub","version":"0.0.0"}}`)
	case "tools/list":
		if mode == "error" {
			replyError(out, *req.ID, -32000, "stub tools/list failure")
			return
		}
		reply(out, *req.ID, `{"tools":[{"name":"echo","description":"echoes input"},{"name":"add","description":"adds numbers"}]}`)
	default:
		replyError(out, *req.ID, -32601, "method not found")
	}
}

func reply(out *bufio.Writer, id int, resultJSON string) {
	fmt.Fprintf(out, `{"jsonrpc":"2.0","id":%d,"result":%s}`+"\n", id, resultJSON)
	out.Flush()
}

func replyError(out *bufio.Writer, id, code int, msg string) {
	b, _ := json.Marshal(msg)
	fmt.Fprintf(out, `{"jsonrpc":"2.0","id":%d,"error":{"code":%d,"message":%s}}`+"\n", id, code, string(b))
	out.Flush()
}
