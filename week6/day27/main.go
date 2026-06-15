// Day 27: init(), Blank Imports, //go:embed, //go:build
// HOW TO RUN: go run week6/day27/main.go
//
// Where you see this in graph-harness:
//   extractors/all/all.go            — blank imports to trigger extractor init()s
//   cmd/graph-harness/main.go:18     — `_ "github.com/.../extractors/all"`
//   internal/kernel/embedded.go      — `//go:embed embedded_manifests/*.yaml`
//   internal/source_live/scip/...    — `//go:build genscipfixture` build tag
//
// Java dev key shifts:
//   - init() is like Java's static {} initializer, but per file
//   - Multiple init() per package: all run in source order, then dependency order
//   - Blank import `_ "pkg"` runs pkg's init() for side effects (no exported names)
//   - This is Go's plugin/registry pattern: "just import it and it self-registers"
//   - //go:embed bundles files into the binary at compile time (like Java resources)
//   - //go:build replaces the old // +build syntax (Go 1.17+)

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

// ============================================================
// PART 1: init() — AUTOMATIC SETUP
// ============================================================

// init() runs automatically before main(). You cannot call it manually.
// It has no arguments and no return values.
// Multiple init() functions in the same file/package are allowed — all run.

var registeredHandlers []string // package-level variable, set before init() runs

func init() {
	// This runs before main(), after package-level vars are initialized
	registeredHandlers = append(registeredHandlers, "handler-from-init-1")
	fmt.Println("[init] first init() ran")
}

func init() {
	// Second init in the same file — also runs, after the first
	registeredHandlers = append(registeredHandlers, "handler-from-init-2")
	fmt.Println("[init] second init() ran")
}

// ============================================================
// PART 2: REGISTRY PATTERN — how graph-harness registers extractors
// ============================================================
//
// The pattern:
//   1. A central registry package exposes Register(name, constructor)
//   2. Each plugin package calls Register() inside its own init()
//   3. The main program does:  import _ "plugins/all"
//   4. "plugins/all" does:     import _ "plugins/kafka"
//                               import _ "plugins/amqp"
//   5. Each plugin's init() fires automatically → self-registers
//   6. main() asks registry.All() for the list of registered plugins
//
// This is Go's idiom for extensible systems without a service locator.
// Java equivalent: ServiceLoader + META-INF/services, or Spring @Component scan.

type ExtractorConstructor func() string // simplified: returns a description

var extractorRegistry = map[string]ExtractorConstructor{}

// Register is called by each extractor's init() — simulating graph-harness's
// code_framework.Register(name, constructor)
func Register(name string, ctor ExtractorConstructor) {
	extractorRegistry[name] = ctor
	fmt.Printf("[registry] registered extractor: %s\n", name)
}

// All returns registered extractors — called in main after all inits ran
func All() map[string]ExtractorConstructor {
	return extractorRegistry
}

// Simulating an extractor package's init() — in real graph-harness,
// each extractors/kafka/go/extractor.go has something like this:
func init() {
	Register("kafka-go", func() string { return "Kafka Go extractor" })
}

func init() {
	Register("amqp-go", func() string { return "AMQP Go extractor" })
}

// ============================================================
// PART 3: //go:embed — bundle files into binary
// ============================================================
//
// Syntax: place //go:embed directive directly above a variable declaration.
// The variable must be of type string, []byte, or embed.FS
//
// embed.FS is most useful when you need multiple files or directory trees.
// graph-harness uses this for YAML manifests in internal/kernel/embedded.go:
//   //go:embed embedded_manifests/*.yaml
//   var embeddedManifests embed.FS

//go:embed testdata/config.yaml
var configBytes []byte // embed a single file as []byte

//go:embed testdata/config.yaml
var configString string // or as string

//go:embed testdata/*
var staticFiles embed.FS // embed a whole directory tree

func demonstrateEmbed() {
	fmt.Println("\n=== //go:embed demo ===")

	// Access the single-file embed
	fmt.Printf("config.yaml (%d bytes):\n%s\n", len(configBytes), configString)

	// Walk the embedded filesystem
	fs.WalkDir(staticFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, _ := staticFiles.ReadFile(path)
		fmt.Printf("  embedded: %s (%d bytes)\n", path, len(data))
		return nil
	})

	// Read a specific file from embed.FS
	data, err := staticFiles.ReadFile("testdata/config.yaml")
	if err != nil {
		fmt.Println("read error:", err)
	} else {
		fmt.Printf("Read via embed.FS: %d bytes\n", len(data))
	}
}

// ============================================================
// PART 4: //go:build — conditional compilation
// ============================================================
//
// //go:build tags control whether a file is included in the build.
// They must be the FIRST non-empty line in the file (before package).
//
// graph-harness example (scip fixture generator):
//   //go:build genscipfixture
//   package main
//
// This file is ONLY compiled when: go build -tags genscipfixture ./...
// Without the tag, the file is completely ignored.
//
// Common patterns:
//   //go:build linux           — OS-specific code
//   //go:build !windows        — exclude Windows
//   //go:build integration     — integration tests (run with: go test -tags integration)
//   //go:build ignore          — permanently exclude a file (draft/scratch)
//
// Java equivalent: Maven profiles, but Go's is checked at compile time by the
// compiler itself — no build tool involvement.
//
// NOTE: The current file has NO build tag, so it always compiles.
// Look at the buildtag_demo.go file below for an example.

func showBuildTagInfo() {
	fmt.Println("\n=== Build tag examples ===")
	fmt.Println("//go:build linux       → only on Linux")
	fmt.Println("//go:build !cgo        → only when CGO disabled")
	fmt.Println("//go:build integration → go test -tags integration ./...")
	fmt.Println("//go:build ignore      → permanently excluded from all builds")
}

func main() {
	fmt.Println("=== init() and registry ===")
	fmt.Println("Handlers registered via init():", registeredHandlers)

	fmt.Println("\n=== extractor registry (populated by init()) ===")
	for name, ctor := range All() {
		fmt.Printf("  %s → %s\n", name, ctor())
	}

	demonstrateEmbed()
	showBuildTagInfo()

	fmt.Println("\n=== Key insight ===")
	fmt.Println("In graph-harness, you never see:")
	fmt.Println("  extractors := []Extractor{NewKafka(), NewAMQP(), ...}")
	fmt.Println("Instead, each extractor self-registers via init().")
	fmt.Println("main.go just imports _ \"extractors/all\" and asks the registry.")
	fmt.Println("")
	fmt.Println("This is the Go equivalent of Spring @Component scan.")
	fmt.Println("New extractor = new package + Register() in init(). Zero changes to main.")

	// Blank import side effect demo — in this file we simulate it inline.
	// In real code it would look like:
	//   import _ "github.com/myapp/plugins/kafka"
	// which would make kafka's init() run without needing any kafka identifier.
	_ = strings.ToUpper // just to use the import
}

// ============================================================
// EXERCISES
// ============================================================
//
// Exercise 1: init() ordering
//   Create a new file week6/day27/extra.go in the same package.
//   Add a package-level var: var extraMsg = "set at package level"
//   Add an init() that appends to registeredHandlers: "handler-from-extra"
//   Run and observe: init() in extra.go runs BEFORE main() but you can't
//   control which file's init() runs first. Go only guarantees:
//     1. All package-level vars initialized
//     2. All init()s run
//     3. Then main()
//
// Exercise 2: Self-registering types
//   Create a new type: type Plugin struct { Name string; Version string }
//   Create a package-level var: var plugins []Plugin
//   Create a RegisterPlugin(p Plugin) function.
//   Add 3 different init() calls that register different plugins.
//   In main(), print all registered plugins.
//   This is exactly how graph-harness's extractor system works.
//
// Exercise 3: embed.FS as config source
//   Create testdata/settings.json with some JSON content.
//   Embed it with //go:embed testdata/settings.json into a []byte.
//   Unmarshal it using encoding/json (from Day 20).
//   Print the parsed values.
//   This is how graph-harness loads its YAML manifests at startup.
//
// Exercise 4: Build tag guard
//   Create a file week6/day27/slow_test_helpers.go
//   Add at the top: //go:build integration
//   Add a function: func heavySetup() string { return "heavy setup done" }
//   Try to call heavySetup() from main without the tag — it should fail to compile.
//   Then run: go build -tags integration ./week6/day27/
//   The function becomes available.
//
// Exercise 5: Blank import chain
//   Create week6/day27/plugins/kafka/kafka.go — calls Register("kafka-v2", ...)
//   Create week6/day27/plugins/all/all.go — blank imports kafka
//   In main.go, blank import "all" — should see kafka-v2 in the registry
//   without main.go knowing anything about kafka.
