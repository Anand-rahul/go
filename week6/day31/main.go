// Day 31: JSON-RPC 2.0, net/rpc, Protocol Buffers (protobuf), gRPC concepts
// HOW TO RUN: go run week6/day31/main.go
// PROTOBUF:   requires protoc + go install google.golang.org/protobuf/cmd/protoc-gen-go
//             (protobuf generation is shown conceptually — no generated files needed here)
//
// Where you see this in graph-harness:
//   internal/jsonrpc/server.go        — JSON-RPC 2.0 server
//   internal/source_live/lsp/jsonrpc.go — LSP uses JSON-RPC 2.0 as transport
//   internal/source_live/scip/        — SCIP uses protobuf binary format
//   go.mod: github.com/sourcegraph/jsonrpc2 + google.golang.org/protobuf
//
// Java dev key shifts:
//   - JSON-RPC 2.0: HTTP-like but over any transport (TCP, stdio, websocket)
//   - RPC = Remote Procedure Call — call a function on another process/machine
//   - Protobuf = language-neutral binary serialization (faster + smaller than JSON)
//   - gRPC = Google's RPC framework using HTTP/2 + protobuf
//   - LSP (Language Server Protocol) is JSON-RPC 2.0 over stdio — that's it

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

// ============================================================
// PART 1: Go's built-in net/rpc — simple RPC over TCP
// ============================================================
//
// net/rpc is Go's built-in RPC framework.
// Methods exported on a registered type become callable remotely.
// Transport: TCP, HTTP, Unix socket.
// Encoding: gob (Go's binary format) by default. JSON-RPC via rpc/jsonrpc.
//
// Rules for net/rpc methods:
//   1. Exported method (capital letter)
//   2. Exactly 2 args: (args *T, reply *R)
//   3. Returns error
//   4. Both T and R must be exportable (gob-encodable)
//
// Java equivalent: Java RMI, or exposing a bean via Spring Remoting.

type AnalyzeArgs struct {
	FilePath string
	Content  string
}

type AnalyzeReply struct {
	EntityCount int
	Entities    []string
	Duration    time.Duration
}

// EntityAnalyzer — the service type registered with net/rpc
type EntityAnalyzer struct{}

// Analyze satisfies net/rpc method signature
func (a *EntityAnalyzer) Analyze(args *AnalyzeArgs, reply *AnalyzeReply) error {
	start := time.Now()

	// Simulate entity extraction
	var entities []string
	for _, line := range strings.Split(args.Content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "func ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				entities = append(entities, "func:"+strings.TrimSuffix(parts[1], "("))
			}
		}
		if strings.HasPrefix(line, "type ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				entities = append(entities, "type:"+parts[1])
			}
		}
	}

	reply.Entities = entities
	reply.EntityCount = len(entities)
	reply.Duration = time.Since(start)
	return nil
}

func netRPCDemo() {
	fmt.Println("=== net/rpc demo ===")

	// ── SERVER ──
	rpc.Register(&EntityAnalyzer{})

	listener, err := net.Listen("tcp", "127.0.0.1:0") // :0 = OS picks a free port
	if err != nil {
		fmt.Println("listen error:", err)
		return
	}
	defer listener.Close()

	go func() {
		// Accept one connection for this demo
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		rpc.ServeConn(conn) // handles all RPC calls on this connection
	}()

	// ── CLIENT ──
	client, err := rpc.Dial("tcp", listener.Addr().String())
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	defer client.Close()

	args := &AnalyzeArgs{
		FilePath: "dispatcher.go",
		Content: `
package main

type Dispatcher struct{}

func (d *Dispatcher) Start() bool { return true }
func (d *Dispatcher) Stop() {}
func (d *Dispatcher) Dispatch(event string) {}
`,
	}
	reply := &AnalyzeReply{}

	// Synchronous call — blocks until result arrives
	if err := client.Call("EntityAnalyzer.Analyze", args, reply); err != nil {
		fmt.Println("RPC error:", err)
		return
	}

	fmt.Printf("Found %d entities in %v:\n", reply.EntityCount, reply.Duration)
	for _, e := range reply.Entities {
		fmt.Printf("  %s\n", e)
	}

	// Async call — returns immediately, result delivered via channel
	replyAsync := &AnalyzeReply{}
	call := client.Go("EntityAnalyzer.Analyze", args, replyAsync, nil)
	select {
	case <-call.Done:
		fmt.Printf("Async call done: %d entities\n", replyAsync.EntityCount)
	case <-time.After(2 * time.Second):
		fmt.Println("Async call timed out")
	}
}

// ============================================================
// PART 2: JSON-RPC 2.0 — what graph-harness actually uses
// ============================================================
//
// JSON-RPC 2.0 is the protocol used by:
//   - LSP (Language Server Protocol) — all editor/LSP communication
//   - graph-harness internal bus (jsonrpc/server.go)
//
// Format: newline-delimited JSON messages with Content-Length header (LSP) or raw JSON
//
// Request:  {"jsonrpc":"2.0","id":1,"method":"textDocument/hover","params":{...}}
// Response: {"jsonrpc":"2.0","id":1,"result":{...}}
// Error:    {"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"Method not found"}}
// Notify:   {"jsonrpc":"2.0","method":"window/showMessage","params":{...}} (no "id")
//
// The key difference from net/rpc: JSON-RPC is transport-agnostic.
// LSP uses it over stdio (the language server reads from stdin, writes to stdout).
// graph-harness uses sourcegraph/jsonrpc2 library for this.

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"` // nil = notification (no response)
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"` // raw JSON for any params type
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Standard JSON-RPC 2.0 error codes
const (
	ErrParseError     = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternalError  = -32603
)

// SimpleJSONRPCServer handles JSON-RPC 2.0 messages from a reader/writer pair.
// graph-harness's real server is more sophisticated (uses sourcegraph/jsonrpc2)
// but this shows the core loop.
type SimpleJSONRPCServer struct {
	handlers map[string]func(params json.RawMessage) (any, error)
	nextID   atomic.Int64
}

func NewJSONRPCServer() *SimpleJSONRPCServer {
	return &SimpleJSONRPCServer{
		handlers: make(map[string]func(params json.RawMessage) (any, error)),
	}
}

func (s *SimpleJSONRPCServer) Handle(method string, fn func(params json.RawMessage) (any, error)) {
	s.handlers[method] = fn
}

func (s *SimpleJSONRPCServer) ServeReader(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			writeError(w, nil, ErrParseError, "parse error")
			continue
		}

		handler, ok := s.handlers[req.Method]
		if !ok {
			if req.ID != nil { // notifications have no ID, no response needed
				writeError(w, req.ID, ErrMethodNotFound, "method not found: "+req.Method)
			}
			continue
		}

		result, err := handler(req.Params)
		if req.ID == nil {
			continue // notification — no response
		}

		if err != nil {
			writeError(w, req.ID, ErrInternalError, err.Error())
			continue
		}

		resultBytes, _ := json.Marshal(result)
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  resultBytes,
		}
		data, _ := json.Marshal(resp)
		fmt.Fprintf(w, "%s\n", data)
	}
}

func writeError(w io.Writer, id *int, code int, msg string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &JSONRPCError{Code: code, Message: msg},
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "%s\n", data)
}

func jsonRPCDemo() {
	fmt.Println("\n=== JSON-RPC 2.0 demo ===")

	server := NewJSONRPCServer()

	// Register method handlers
	server.Handle("analyze/file", func(params json.RawMessage) (any, error) {
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(params, &args); err != nil {
			return nil, err
		}
		return map[string]any{
			"path":     args.Path,
			"entities": 42,
			"ok":       true,
		}, nil
	})

	server.Handle("status", func(params json.RawMessage) (any, error) {
		return map[string]string{"status": "ok", "version": "1.0.0"}, nil
	})

	// Simulate client sending JSON-RPC messages over stdio
	id1, id2 := 1, 2
	requests := []JSONRPCRequest{
		{JSONRPC: "2.0", ID: &id1, Method: "analyze/file",
			Params: json.RawMessage(`{"path":"internal/api/users.go"}`)},
		{JSONRPC: "2.0", ID: &id2, Method: "status"},
		{JSONRPC: "2.0", Method: "workspace/didChange",
			Params: json.RawMessage(`{"uri":"file:///foo.go"}`)}, // notification — no ID
		{JSONRPC: "2.0", ID: &id2, Method: "nonexistent/method"},
	}

	// Build input pipe
	var input strings.Builder
	for _, req := range requests {
		data, _ := json.Marshal(req)
		input.Write(data)
		input.WriteByte('\n')
	}

	fmt.Println("Server responses:")
	server.ServeReader(strings.NewReader(input.String()), os.Stdout)
}

// ============================================================
// PART 3: Protocol Buffers — what they are and how Go uses them
// ============================================================
//
// graph-harness uses protobuf for SCIP (Source Code Intelligence Protocol):
//   internal/source_live/scip/proto/scip.proto → scip.pb.go (generated)
//
// Protobuf is:
//   - Language-neutral binary serialization (like Avro or Thrift)
//   - MUCH smaller/faster than JSON for large data (SCIP index files can be GBs)
//   - Schema-first: define in .proto file, generate Go code with protoc
//   - Forward/backward compatible with field numbers
//
// Java equivalent: Protocol Buffers (same thing — Google invented it,
//   has first-class Java support via maven plugin + protoc-java-api)
//
// .proto file looks like:
//   syntax = "proto3";
//   message Document {
//     string uri = 1;
//     repeated Occurrence occurrences = 2;
//   }
//   message Occurrence {
//     repeated int32 range = 1;
//     int32 symbol_roles = 4;
//   }
//
// Generated Go code:
//   type Document struct {
//     Uri         string      `protobuf:"bytes,1,opt,name=uri" json:"uri,omitempty"`
//     Occurrences []*Occurrence `protobuf:"bytes,2,rep,name=occurrences" json:"occurrences,omitempty"`
//   }
//
// Usage in Go:
//   import "google.golang.org/protobuf/proto"
//
//   // Marshal to binary
//   data, err := proto.Marshal(document)
//
//   // Unmarshal from binary
//   var doc scip.Document
//   err := proto.Unmarshal(data, &doc)
//
// graph-harness reads .scip index files like this:
//   data, _ := os.ReadFile("index.scip")
//   var index scip.Index
//   proto.Unmarshal(data, &index)
//   for _, doc := range index.Documents { ... }

func protobufInfo() {
	fmt.Println("\n=== Protocol Buffers (conceptual) ===")

	fmt.Println(`
.proto definition:
  syntax = "proto3";
  message Document {
    string uri = 1;                    // field number 1 — never change this!
    repeated Occurrence occurrences = 2;
  }

Generated Go struct (via protoc-gen-go):
  type Document struct {
    Uri         string
    Occurrences []*Occurrence
  }

Usage:
  // Marshal to binary (much smaller than JSON)
  data, err := proto.Marshal(document)

  // Unmarshal
  var doc scip.Document
  proto.Unmarshal(data, &doc)

Why graph-harness uses protobuf (not JSON):
  - SCIP index for a large repo can be 100MB+ of symbol data
  - Protobuf binary is ~5-10x smaller and ~10x faster to parse than JSON
  - Field numbers mean old code can read new protobufs (forward compat)

Key difference from JSON struct tags:
  - json:"name"   → field name in JSON text
  - protobuf:"bytes,1,..." → field NUMBER in binary (numbers are stable IDs)
`)
}

// ============================================================
// PART 4: gRPC — what it is (graph-harness doesn't use it but you'll see it)
// ============================================================
//
// gRPC = Google RPC. Built on:
//   - HTTP/2 (multiplexed, bidirectional streaming)
//   - Protocol Buffers (default serialization)
//   - Service definitions in .proto
//
// .proto service definition:
//   service EntityService {
//     rpc Analyze(AnalyzeRequest) returns (AnalyzeResponse);
//     rpc Watch(WatchRequest) returns (stream EntityEvent);  // server streaming
//   }
//
// Generated Go code creates:
//   - Client stub (EntityServiceClient) — call methods like Go functions
//   - Server interface (EntityServiceServer) — implement methods to serve
//
// Java equivalent: gRPC-Java (exact same protocol, .proto files are shared)
//
// graph-harness uses JSON-RPC (not gRPC) because:
//   - LSP protocol is defined as JSON-RPC 2.0 — no choice
//   - JSON-RPC over stdio is simpler for language server processes
//   - gRPC is better for service-to-service HTTP/2 communication

func grpcInfo() {
	fmt.Println("=== gRPC (conceptual — not used in graph-harness) ===")
	fmt.Println(`
gRPC .proto service:
  service EntityService {
    rpc Analyze(AnalyzeRequest) returns (AnalyzeResponse);
    rpc Watch(WatchRequest) returns (stream EntityEvent);
  }

vs JSON-RPC (what graph-harness uses):
  - gRPC: HTTP/2 + protobuf binary, streaming built-in
  - JSON-RPC: any transport + JSON text, simpler, no codegen needed
  - LSP MUST use JSON-RPC (the spec says so)
  - graph-harness bus uses JSON-RPC to stay compatible with LSP tooling
`)
}

func main() {
	netRPCDemo()
	jsonRPCDemo()
	protobufInfo()
	grpcInfo()
}

// ============================================================
// EXERCISES
// ============================================================
//
// Exercise 1: Bidirectional net/rpc
//   Add a second method to EntityAnalyzer:
//     func (a *EntityAnalyzer) CountLines(args *AnalyzeArgs, reply *int) error
//   It should count the non-empty lines in args.Content.
//   Call it from the client alongside Analyze.
//   Use client.Go() for both calls simultaneously — collect results from both
//   call.Done channels. This is the async fan-out pattern from net/rpc.
//
// Exercise 2: JSON-RPC notification handler
//   Add a "workspace/didChange" handler to SimpleJSONRPCServer.
//   Notifications have no ID — the server must NOT send a response.
//   Log "file changed: <uri>" when this notification arrives.
//   Verify that when you send it, nothing appears on the output (no response).
//
// Exercise 3: JSON-RPC error handling
//   Add a handler for "analyze/bad" that returns an error:
//     return nil, fmt.Errorf("feature not implemented")
//   Send a request to this method — verify the response has "error" not "result".
//   Add another handler that returns nil result — verify "result": null in response.
//
// Exercise 4: Protobuf-like manual binary encoding (educational)
//   Without any library, implement a tiny encoder for this schema:
//     type EntityRecord struct { ID int32; Name string; Kind string }
//   Encode to []byte using: 4 bytes for ID, 1 byte for len(Name), Name bytes,
//   1 byte for len(Kind), Kind bytes.
//   Write Encode(r EntityRecord) []byte and Decode(data []byte) (EntityRecord, error).
//   This shows why protobuf (with field numbers + varint encoding) is better.
//
// Exercise 5: net/rpc over Unix socket
//   Instead of TCP, use net.Listen("unix", "/tmp/harness.sock").
//   Start server in a goroutine.
//   Connect client with rpc.Dial("unix", "/tmp/harness.sock").
//   Make 3 concurrent calls using client.Go() and collect all results.
//   Unix sockets are faster than TCP for local communication — graph-harness
//   uses them for communication between the daemon and CLI.
