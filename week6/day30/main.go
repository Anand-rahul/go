// Day 30: Cobra CLI, YAML/TOML config parsing, fuzz testing
// HOW TO RUN: go run week6/day30/main.go [command] [flags]
//   go run week6/day30/main.go serve --port 9090 --debug
//   go run week6/day30/main.go watch --workspace /tmp
//   go run week6/day30/main.go config show
// FUZZ:       go test -fuzz=FuzzParseManifest ./week6/day30/...
// REQUIRES:   go get github.com/spf13/cobra gopkg.in/yaml.v3 github.com/BurntSushi/toml
//
// Where you see this in graph-harness:
//   internal/cli/  — all CLI commands built with cobra
//   internal/kernel/embedded.go — YAML manifest parsing
//   internal/kernel/provenance_fuzz_test.go — fuzz testing
//
// Java dev key shifts:
//   - Cobra = Spring Shell or Picocli — declarative CLI with subcommands
//   - YAML/TOML: no stdlib support — need external packages (unlike JSON)
//   - Fuzz testing: go test -fuzz — automated input generation (JQF in Java)

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ============================================================
// PART 1: Cobra CLI — subcommands, flags, persistent flags
// ============================================================
//
// Cobra pattern:
//   root command (graph-harness)
//     ├─ serve              ← subcommand
//     ├─ watch              ← subcommand
//     └─ config             ← subcommand group
//         └─ show           ← nested subcommand
//
// Each command is a cobra.Command with:
//   Use:   command name + arg description
//   Short: one-line description (shown in help)
//   RunE:  function to run (returns error — cobra handles printing it)

// Global flags (set on rootCmd, available to all subcommands)
var (
	cfgFile string
	verbose bool
)

func buildCLI() *cobra.Command {
	// ROOT COMMAND
	rootCmd := &cobra.Command{
		Use:   "harness",
		Short: "graph-harness: code structure analysis tool",
		// No RunE on root — it just shows help. Subcommands do the work.
	}

	// PersistentFlags: inherited by ALL subcommands (like global options)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// ── SERVE subcommand ──
	var port int
	var debug bool

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the graph-harness daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Starting server on port %d (debug=%v, verbose=%v)\n", port, debug, verbose)
			if cfgFile != "" {
				fmt.Printf("Using config: %s\n", cfgFile)
			}
			return nil
		},
	}
	// Local flags: only available on this subcommand
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "port to listen on")
	serveCmd.Flags().BoolVar(&debug, "debug", false, "enable debug mode")

	// ── WATCH subcommand ──
	var workspace string
	var excludes []string

	watchCmd := &cobra.Command{
		Use:   "watch [flags]",
		Short: "Watch a workspace for changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Watching workspace: %s\n", workspace)
			if len(excludes) > 0 {
				fmt.Printf("Excluding: %s\n", strings.Join(excludes, ", "))
			}
			return nil
		},
	}
	watchCmd.Flags().StringVarP(&workspace, "workspace", "w", ".", "workspace directory")
	watchCmd.Flags().StringSliceVar(&excludes, "exclude", nil, "patterns to exclude")

	// ── CONFIG subcommand GROUP ──
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Config file:", cfgFile)
			fmt.Println("Verbose:", verbose)
			return nil
		},
	}

	// Nested subcommand: config show
	configCmd.AddCommand(configShowCmd)

	rootCmd.AddCommand(serveCmd, watchCmd, configCmd)
	return rootCmd
}

// ============================================================
// PART 2: YAML parsing — graph-harness manifest format
// ============================================================
//
// graph-harness YAML manifests describe entity kinds, event kinds, etc.
// They're embedded at compile time (//go:embed) and parsed at startup.
//
// YAML struct tags use `yaml:"fieldname"` — same idea as `json:"..."`
// yaml.v3 supports:
//   - struct tags for mapping
//   - inline embeds
//   - custom Unmarshaler interface
//
// Java: Jackson with yaml-dataformat, or SnakeYAML

type ManifestConfig struct {
	APIVersion string    `yaml:"apiVersion"`
	Kind       string    `yaml:"kind"`
	Metadata   Metadata  `yaml:"metadata"`
	Spec       HarnessSpec `yaml:"spec"`
}

type Metadata struct {
	Name    string            `yaml:"name"`
	Labels  map[string]string `yaml:"labels,omitempty"`
}

type HarnessSpec struct {
	EntityKinds []EntityKind `yaml:"entityKinds"`
	EventKinds  []EventKind  `yaml:"eventKinds"`
}

type EntityKind struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type EventKind struct {
	Name    string `yaml:"name"`
	Trigger string `yaml:"trigger"`
}

const sampleManifest = `
apiVersion: graph-harness/v1
kind: Harness
metadata:
  name: code-framework
  labels:
    component: core
spec:
  entityKinds:
    - name: route
      description: HTTP route definition
    - name: event
      description: Domain event type
  eventKinds:
    - name: FileChanged
      trigger: on-file-write
    - name: SymbolResolved
      trigger: on-lsp-response
`

func yamlDemo() {
	fmt.Println("\n=== YAML parsing demo ===")

	var cfg ManifestConfig
	if err := yaml.Unmarshal([]byte(sampleManifest), &cfg); err != nil {
		fmt.Println("YAML parse error:", err)
		return
	}

	fmt.Printf("Manifest: %s/%s (name: %s)\n", cfg.APIVersion, cfg.Kind, cfg.Metadata.Name)
	fmt.Printf("Entity kinds: %d\n", len(cfg.Spec.EntityKinds))
	for _, ek := range cfg.Spec.EntityKinds {
		fmt.Printf("  - %s: %s\n", ek.Name, ek.Description)
	}
	fmt.Printf("Event kinds: %d\n", len(cfg.Spec.EventKinds))

	// Marshal back to YAML
	out, _ := yaml.Marshal(&cfg)
	fmt.Printf("Round-trip marshal (%d bytes)\n", len(out))
}

// ============================================================
// PART 3: TOML parsing — graph-harness .graph-harness.toml
// ============================================================
//
// graph-harness uses TOML for the workspace config file (.graph-harness.toml).
// TOML is preferred for human-edited config files (cleaner than YAML for nested tables).
//
// Java: no stdlib TOML — need jackson-dataformat-toml or similar.

type WorkspaceConfig struct {
	Workspace WorkspaceSection `toml:"workspace"`
	Extractors []string        `toml:"extractors"`
}

type WorkspaceSection struct {
	Root    string   `toml:"root"`
	Exclude []string `toml:"exclude"`
	MaxSize int      `toml:"max_size_mb"`
}

const sampleTOML = `
extractors = ["kafka-go", "amqp-go", "graphql-ts"]

[workspace]
root = "/home/rahul/myproject"
exclude = ["vendor", "node_modules", ".git"]
max_size_mb = 500
`

func tomlDemo() {
	fmt.Println("\n=== TOML parsing demo ===")

	var cfg WorkspaceConfig
	if _, err := toml.Decode(sampleTOML, &cfg); err != nil {
		fmt.Println("TOML parse error:", err)
		return
	}

	fmt.Printf("Workspace root: %s\n", cfg.Workspace.Root)
	fmt.Printf("Max size: %d MB\n", cfg.Workspace.MaxSize)
	fmt.Printf("Extractors: %v\n", cfg.Extractors)
	fmt.Printf("Excluded: %v\n", cfg.Workspace.Exclude)

	// toml.NewEncoder for marshal — symmetrical with Decode
	// var buf strings.Builder
	// toml.NewEncoder(&buf).Encode(cfg)
	fmt.Println("TOML parsed OK")
}

// ============================================================
// PART 4: Fuzz testing
// ============================================================
//
// Fuzz testing = automated randomized input generation to find crashes/panics.
// graph-harness has: internal/kernel/provenance_fuzz_test.go
//   func FuzzProvenanceRoundTrip(f *testing.F) {
//       f.Add("seed input")   // seed corpus
//       f.Fuzz(func(t *testing.T, input string) {
//           // must not panic for any input
//       })
//   }
//
// Run with: go test -fuzz=FuzzFunctionName ./pkg/...
// Go mutates inputs and tries to find crashes. Saved to testdata/fuzz/ corpus.
//
// Java equivalent: JQF (junit-quickcheck), but Go's is built-in since 1.18.
//
// *** The actual fuzz function must be in a _test.go file — see day30_fuzz_test.go ***

func fuzzInfo() {
	fmt.Println("\n=== Fuzz testing info ===")
	fmt.Println("Fuzz functions live in *_test.go files.")
	fmt.Println("See: week6/day30/day30_fuzz_test.go")
	fmt.Println("Run: go test -fuzz=FuzzParseManifest ./week6/day30/")
	fmt.Println("     go test -fuzz=FuzzHashContent ./week6/day30/")
	fmt.Println("")
	fmt.Println("Graph-harness: internal/kernel/provenance_fuzz_test.go")
	fmt.Println("  → verifies provenance round-trip never panics on any input")
}

func main() {
	// Run the CLI — os.Args drives it
	// For demo purposes, if no args given, show help
	if len(os.Args) == 1 {
		// Demo mode: show what the CLI produces
		fmt.Println("=== Cobra CLI demo ===")
		fmt.Println("Try: go run week6/day30/main.go serve --port 9090 --debug")
		fmt.Println("     go run week6/day30/main.go watch --workspace /tmp")
		fmt.Println("     go run week6/day30/main.go config show --verbose")
		fmt.Println("     go run week6/day30/main.go --help")
		fmt.Println()
		yamlDemo()
		tomlDemo()
		fuzzInfo()
		return
	}

	rootCmd := buildCLI()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// ============================================================
// EXERCISES
// ============================================================
//
// Exercise 1: Add a `query` subcommand
//   Add: harness query --kind route --path "internal/api/*"
//   Flags: --kind (string), --path (string glob), --limit (int, default 20)
//   RunE should print: "querying kind=route path=internal/api/* limit=20"
//   Use PersistentPreRunE on root to print "Config: <cfgFile>" before any command runs.
//
// Exercise 2: YAML config with defaults
//   Extend ManifestConfig to have a Settings section:
//     settings:
//       timeout_ms: 5000
//       max_retries: 3
//       log_level: "info"
//   Write func loadOrDefault(path string) ManifestConfig:
//     If path == "", return a hardcoded default config.
//     If path exists, parse it.
//     If parse fails, return the default config and log the error.
//
// Exercise 3: TOML round-trip
//   Create a WorkspaceConfig programmatically (not from string).
//   Marshal it to TOML using toml.NewEncoder into a strings.Builder.
//   Print the TOML string.
//   Then unmarshal it back and compare — should be identical.
//
// Exercise 4: Required flags
//   In the watch command, mark --workspace as Required:
//     watchCmd.MarkFlagRequired("workspace")
//   Run without --workspace — Cobra should reject it with a helpful error.
//   Run with --workspace /tmp — should work.
//
// Exercise 5: Cobra command validation
//   The `query` subcommand should accept exactly 1 positional argument (the entity name).
//   Add: Args: cobra.ExactArgs(1)
//   RunE receives args []string — use args[0] as the entity name.
//   Test: run without arg (error), run with one arg (works), run with two (error).
