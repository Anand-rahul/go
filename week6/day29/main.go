// Day 29: os/exec, regexp, crypto/sha256, type aliases, encoding/hex
// HOW TO RUN: go run week6/day29/main.go
//
// Where you see this in graph-harness:
//   internal/extract/orchestrator.go    — os/exec for LSP probing (is gopls installed?)
//   internal/change_process/pipeline.go — regexp for diff hunk matching
//   internal/code_framework/types.go    — crypto/sha256 + hex for ContentID hashing
//   internal/code_framework/types.go    — type ContentID = string (type alias)
//
// Java dev key shifts:
//   - os/exec: like Java's ProcessBuilder — run external commands, capture output
//   - regexp: compiled regexps are cached — like java.util.regex.Pattern.compile()
//   - crypto/sha256: hash.New() → Write → Sum — like Java's MessageDigest
//   - type alias (type X = Y): X and Y are THE SAME TYPE — unlike type definition

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// ============================================================
// PART 1: os/exec — running external commands
// ============================================================
//
// graph-harness uses this in extract/orchestrator.go to probe whether
// language servers are installed before trying to start them:
//   cmd := exec.Command("gopls", "version")
//   if err := cmd.Run(); err != nil { /* gopls not found */ }
//
// Also used to start the actual LSP server process and talk to it via stdin/stdout.

func execDemo() {
	fmt.Println("=== os/exec demo ===")

	// === Simple: run and capture combined output ===
	// Java: ProcessBuilder pb = new ProcessBuilder("go", "version"); pb.start()...
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		fmt.Println("go version failed:", err)
	} else {
		fmt.Printf("go version: %s", out)
	}

	// === Check if a tool exists ===
	// exec.LookPath finds the binary in PATH — like `which gopls`
	path, err := exec.LookPath("gopls")
	if err != nil {
		fmt.Println("gopls NOT found in PATH — LSP disabled")
	} else {
		fmt.Printf("gopls found at: %s\n", path)
	}

	// === Capture stdout and stderr separately ===
	cmd := exec.Command("go", "list", "./week6/...")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("command failed: %v\nstderr: %s\n", err, stderr.String())
	} else {
		lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
		fmt.Printf("go list found %d packages\n", len(lines))
	}

	// === Check exit code ===
	// exec.ExitError carries the exit code
	cmd2 := exec.Command("go", "build", "./nonexistent/...")
	if err := cmd2.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("exit code: %d\n", exitErr.ExitCode())
		}
	}

	// === Long-running process: Start() + Wait() ===
	// Use when you need to write to stdin or read from stdout while running.
	// graph-harness does this for the LSP server process.
	//
	// cmd := exec.Command("gopls", "-mode=stdio")
	// cmd.Stdin = ...   // pipe for JSON-RPC requests
	// cmd.Stdout = ...  // pipe for JSON-RPC responses
	// cmd.Start()
	// defer cmd.Wait()
	fmt.Println("(Start/Wait pattern shown in comments — used for LSP server process)")
}

// ============================================================
// PART 2: regexp — compiled patterns
// ============================================================
//
// graph-harness uses regexp in change_process/pipeline.go to match:
//   - diff hunk headers:  @@ -10,5 +12,8 @@
//   - function boundaries in source files
//
// Java: Pattern.compile("...").matcher(input).find()
// Go:   regexp.MustCompile("...").MatchString(input)
//
// ALWAYS compile at package level (once). Never inside a hot loop.

// Package-level compiled regexps — zero cost on subsequent calls
var (
	hunkHeaderRe    = regexp.MustCompile(`^@@ -(\d+),(\d+) \+(\d+),(\d+) @@`)
	funcDefRe       = regexp.MustCompile(`^func (?:\([^)]+\) )?(\w+)\(`)
	semverRe        = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9.]+))?$`)
	contentIDHashRe = regexp.MustCompile(`^[0-9a-f]{16}$`)
)

func regexpDemo() {
	fmt.Println("\n=== regexp demo ===")

	// MatchString — does it match?
	lines := []string{
		"@@ -10,5 +12,8 @@ func processEvent",
		"func (d *Dispatcher) Start() bool {",
		"// just a comment",
		"func handleError(err error) {",
	}

	for _, line := range lines {
		switch {
		case hunkHeaderRe.MatchString(line):
			// FindStringSubmatch returns full match + captured groups
			m := hunkHeaderRe.FindStringSubmatch(line)
			fmt.Printf("  HUNK: old line %s (+%s lines), new line %s (+%s lines)\n",
				m[1], m[2], m[3], m[4])
		case funcDefRe.MatchString(line):
			m := funcDefRe.FindStringSubmatch(line)
			fmt.Printf("  FUNC: %s\n", m[1])
		default:
			end := 30
			if end > len(line) {
				end = len(line)
			}
			fmt.Printf("  skip: %q\n", line[:end])
		}
	}

	// FindAllString — find all occurrences
	code := `func Foo() {} func Bar() {} func (r *Rect) Area() float64 {}`
	allFuncs := funcDefRe.FindAllStringSubmatch(code, -1)
	fmt.Printf("\nAll functions: ")
	for _, m := range allFuncs {
		fmt.Printf("%s ", m[1])
	}
	fmt.Println()

	// ReplaceAllString — transform
	redacted := regexp.MustCompile(`\b\d{4}\b`).ReplaceAllString(
		"Card ending in 4242 expires 12/26", "****",
	)
	fmt.Println("Redacted:", redacted)
}

// ============================================================
// PART 3: crypto/sha256 + encoding/hex
// ============================================================
//
// graph-harness uses this in code_framework/types.go to compute ContentID:
// a stable 16-char hex string identifying the content of a code entity.
//
//   h := sha256.New()
//   h.Write([]byte(input))
//   full := hex.EncodeToString(h.Sum(nil))
//   contentID := full[:16]  // first 16 hex chars = 8 bytes = 64-bit hash
//
// Java: MessageDigest.getInstance("SHA-256").digest(input.getBytes())

// ContentID is a stable 16-char hex string for content-addressable lookup.
// type alias: ContentID IS string — they are the exact same type.
// Contrast with type definition (type ContentID string) which would be a NEW type.
type ContentID = string // alias: interchangeable with string everywhere

func hashContent(content string) ContentID {
	h := sha256.New()         // creates a new hash state
	h.Write([]byte(content))  // feed bytes
	full := h.Sum(nil)        // Sum(nil) returns the final hash as []byte
	hexStr := hex.EncodeToString(full) // convert bytes to hex string
	return hexStr[:16]        // first 16 chars (8 bytes) — stable short ID
}

// sha256.Sum256 — one-shot version when you don't need streaming
func quickHash(content string) string {
	sum := sha256.Sum256([]byte(content)) // returns [32]byte (array, not slice)
	return hex.EncodeToString(sum[:])     // sum[:] converts array to slice
}

func hashDemo() {
	fmt.Println("\n=== crypto/sha256 + hex demo ===")

	code1 := `func (d *Dispatcher) Start() bool { return d.started.CompareAndSwap(false, true) }`
	code2 := `func (d *Dispatcher) Stop() { d.started.Store(false) }`

	id1 := hashContent(code1)
	id2 := hashContent(code2)

	fmt.Printf("ContentID for code1: %s\n", id1)
	fmt.Printf("ContentID for code2: %s\n", id2)
	fmt.Printf("Same content, same ID: %v\n", hashContent(code1) == id1)
	fmt.Printf("Different content, different ID: %v\n", id1 != id2)

	// hex.DecodeString — reverse direction
	decoded, _ := hex.DecodeString(id1)
	fmt.Printf("Decoded back to %d bytes\n", len(decoded))
}

// ============================================================
// PART 4: type alias vs type definition
// ============================================================
//
// TYPE ALIAS:      type X = Y    → X and Y are THE SAME TYPE
// TYPE DEFINITION: type X Y      → X is a NEW TYPE with Y as underlying type
//
// Java has no alias — every `class Foo extends Bar` creates a new type.
// Go's aliases exist mainly for refactoring (gradually moving a type to a new package).

type ContentIDDef string     // type DEFINITION — new type, cannot use as string directly
type ContentIDAlias = string // type ALIAS — literally IS string

func typeAliasDemo() {
	fmt.Println("\n=== type alias vs type definition ===")

	var def ContentIDDef = "abc123"
	var alias ContentIDAlias = "abc123"

	// Definition: ContentIDDef is NOT string — needs explicit conversion
	// var s1 string = def  // COMPILE ERROR
	var s1 string = string(def) // explicit conversion required
	fmt.Println("definition needs conversion:", s1)

	// Alias: ContentIDAlias IS string — no conversion needed
	var s2 string = alias // works directly
	fmt.Println("alias is transparent:", s2)

	// graph-harness uses alias for ContentID:
	// type ContentID = string
	// This means ContentID can be passed anywhere string is expected,
	// and any string can be used where ContentID is expected.
	// It's documentation, not type safety.
	// The DEFINITION would provide type safety but requires conversions everywhere.
}

func main() {
	execDemo()
	regexpDemo()
	hashDemo()
	typeAliasDemo()
}

// ============================================================
// EXERCISES
// ============================================================
//
// Exercise 1: Tool prober
//   Write func probeTools(tools []string) map[string]bool
//   For each tool name, use exec.LookPath to check if it's installed.
//   Return a map: tool name → found (true/false).
//   Test with: []string{"go", "git", "docker", "gopls", "nonexistent-tool"}
//
// Exercise 2: Run with timeout
//   Write func runWithTimeout(ctx context.Context, name string, args ...string) ([]byte, error)
//   Use exec.CommandContext(ctx, name, args...) — takes context for cancellation.
//   Create a 2-second timeout context using context.WithTimeout.
//   Run "sleep 5" — it should be killed before finishing.
//   Run "echo hello" — should succeed.
//
// Exercise 3: Named capture groups
//   Diff hunk lines look like: @@ -23,8 +25,12 @@ func Foo() {
//   The function name after @@ is optional.
//   Write a regexp with NAMED captures: (?P<old_start>\d+) etc.
//   Use FindStringSubmatch + re.SubexpIndex("old_start") to extract by name.
//   Parse 3 different hunk headers and print the extracted line numbers.
//
// Exercise 4: Content-addressable store
//   Build a type ContentStore struct { mu sync.RWMutex; data map[string]string }
//   func (s *ContentStore) Put(content string) ContentID — hashes and stores
//   func (s *ContentStore) Get(id ContentID) (string, bool) — retrieves by ID
//   Two calls to Put() with the same content must return the same ContentID.
//   Verify this works correctly.
//
// Exercise 5: Type definition for safety
//   Define: type FilePath string  (definition, not alias)
//   Define: type PackagePath string
//   Write: func openFile(path FilePath) error
//   Observe: passing a PackagePath to openFile is a COMPILE ERROR.
//   This is the opposite choice from ContentID — when you WANT type safety.
//   graph-harness makes this choice deliberately: some IDs are aliases (transparent)
//   others would be definitions (opaque) if stricter type safety were needed.
