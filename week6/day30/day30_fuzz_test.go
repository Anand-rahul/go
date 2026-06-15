// Fuzz tests for Day 30 — run with: go test -fuzz=Fuzz ./week6/day30/
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// FuzzParseManifest: graph-harness equivalent of provenance_fuzz_test.go
// Goal: parsing arbitrary YAML must never PANIC — errors are OK, panics are bugs.
// go test -fuzz=FuzzParseManifest -fuzztime=30s ./week6/day30/
func FuzzParseManifest(f *testing.F) {
	// Seed corpus: known-good inputs that Go starts mutating from
	f.Add(sampleManifest)
	f.Add(`apiVersion: v1
kind: Harness
metadata:
  name: test`)
	f.Add(`{}`)
	f.Add(``)
	f.Add(`not: yaml: at: all: [[[`)

	f.Fuzz(func(t *testing.T, input string) {
		var cfg ManifestConfig
		// Must never panic — errors are fine
		_ = yaml.Unmarshal([]byte(input), &cfg)
		// If it parsed without error, round-trip should also not panic
		if cfg.Kind != "" {
			_, _ = yaml.Marshal(&cfg)
		}
	})
}

// FuzzHashContent: verifies that content hashing is stable and never panics
// Property: same input → same output (deterministic)
// go test -fuzz=FuzzHashContent -fuzztime=30s ./week6/day30/
func FuzzHashContent(f *testing.F) {
	f.Add("func main() {}")
	f.Add("")
	f.Add("hello world")
	f.Add(strings.Repeat("a", 10000))

	f.Fuzz(func(t *testing.T, input string) {
		h1 := sha256.Sum256([]byte(input))
		h2 := sha256.Sum256([]byte(input))

		// Must be deterministic — same input always produces same hash
		if h1 != h2 {
			t.Fatalf("non-deterministic hash for input %q", input[:min(50, len(input))])
		}

		// hex encoding must always produce valid hex
		s := hex.EncodeToString(h1[:])
		if len(s) != 64 {
			t.Fatalf("expected 64 hex chars, got %d", len(s))
		}
	})
}
