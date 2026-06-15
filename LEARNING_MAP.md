# Go Learning Map — Java Dev → Go Mid-Level

> **You**: 4 years Java. **Goal**: Go mid-level proficiency.
> **Approach**: One file per day. Read it, run it, do the exercises, report back.

---

## How to Use This

1. **Each morning**: `cd` into the day's file directory
2. **Read** the file top to bottom (comments explain Java → Go differences)
3. **Run it**: `go run dayXX_topic.go`
4. **Do the exercises** at the bottom of each file
5. **Report back** using `progress/YYYY-MM-DD.md` (copy from `progress/template.md`)

### Setup (do this first)
```bash
# Install Go: https://go.dev/dl/ (or: sudo pacman -S go on your Arch)
go version          # should be 1.21+

# Initialize a module in this directory
cd /home/rahul/codedump/go
go mod init golearn

# Run any day file
go run week1/day01_hello_types.go
```

---

## The Map

| Week | Theme | Days | Key Java → Go shift |
|------|-------|------|---------------------|
| 1 | **Foundations** | 1–5 | Syntax, types, slices, maps, pointers |
| 2 | **OOP without OOP** | 6–10 | Functions, methods, interfaces, errors |
| 3 | **Concurrency** | 11–15 | Goroutines, channels, sync, context |
| 4 | **Go Ecosystem** | 16–20 | Modules, testing, stdlib, HTTP, JSON |
| 5 | **Mid-level** | 21–25 | Generics, defer/panic, patterns, perf, project |

---

## Week 1 — Foundations

| Day | File | What You Learn |
|-----|------|----------------|
| 1 | `week1/day01_hello_types.go` | Hello world, var/const, basic types vs Java primitives |
| 2 | `week1/day02_control_flow.go` | if/for/switch — no while, no do-while, range loops |
| 3 | `week1/day03_slices.go` | Arrays vs Slices vs Java arrays/ArrayList |
| 4 | `week1/day04_maps_structs.go` | Maps (HashMap), Structs (no classes), struct literals |
| 5 | `week1/day05_pointers.go` | Explicit pointers — what Java hides from you |

## Week 2 — OOP Without OOP

| Day | File | What You Learn |
|-----|------|----------------|
| 6 | `week2/day06_functions.go` | Multiple returns, variadic, closures, first-class funcs |
| 7 | `week2/day07_methods.go` | Value vs pointer receivers — Go's version of methods |
| 8 | `week2/day08_interfaces.go` | Implicit interfaces — no `implements`, duck typing |
| 9 | `week2/day09_embedding.go` | Embedding instead of inheritance |
| 10 | `week2/day10_errors.go` | Errors as values — Go's answer to try/catch |

## Week 3 — Concurrency (Go's Superpower)

| Day | File | What You Learn |
|-----|------|----------------|
| 11 | `week3/day11_goroutines.go` | Goroutines vs Java threads — 10k goroutines is normal |
| 12 | `week3/day12_channels.go` | Channels — typed pipes between goroutines |
| 13 | `week3/day13_sync.go` | select, WaitGroup, Mutex — coordination primitives |
| 14 | `week3/day14_context.go` | context.Context — cancellation and deadlines |
| 15 | `week3/day15_patterns.go` | Fan-out/fan-in, pipelines, worker pools |

## Week 4 — Go Ecosystem

| Day | File | What You Learn |
|-----|------|----------------|
| 16 | `week4/day16_modules.go` | go.mod, imports, init() — vs Maven/Gradle |
| 17 | `week4/day17_testing.go` | Table-driven tests, subtests, benchmarks |
| 18 | `week4/day18_io.go` | os, io, bufio, filepath — file I/O |
| 19 | `week4/day19_http.go` | net/http — simple server and client |
| 20 | `week4/day20_json.go` | encoding/json — Marshal/Unmarshal, struct tags |

## Week 5 — Mid-level

| Day | File | What You Learn |
|-----|------|----------------|
| 21 | `week5/day21_generics.go` | Generics (1.18+) — type params, constraints |
| 22 | `week5/day22_defer_panic.go` | defer stack, panic/recover — Go's exception equivalent |
| 23 | `week5/day23_patterns.go` | Functional options, middleware, builder patterns |
| 24 | `week5/day24_benchmarks.go` | Benchmarks, pprof intro, escape analysis |
| 25 | `week5/day25_project.go` | Mini project — CLI task manager, ties everything together |

## Week 6 — Reading Real Go (graph-harness)

> Goal: understand every pattern you'll encounter in the graph-harness codebase.
> Each day maps directly to files in `/home/rahul/codedump/graph-harness/graph-harness/`.

| Day | File | What You Learn | Where in graph-harness |
|-----|------|----------------|------------------------|
| 26 | `week6/day26/main.go` | `database/sql` + SQLite — queries, transactions, prepared statements | `internal/code_core/adjacency.go`, `internal/facts/facts.go` |
| 27 | `week6/day27/main.go` | `init()`, blank imports, plugin registry, `//go:embed`, `//go:build` | `extractors/all/all.go`, `internal/kernel/embedded.go` |
| 28 | `week6/day28/main.go` | `atomic` deep dive, `singleflight`, `sync.Cond`, compile-time interface assertions | `internal/code_framework/dispatch.go`, `entityref_cache.go` |
| 29 | `week6/day29/main.go` | `os/exec`, `regexp`, `crypto/sha256` + `encoding/hex`, type alias vs definition | `internal/extract/orchestrator.go`, `code_framework/types.go` |
| 30 | `week6/day30/main.go` | Cobra CLI, YAML + TOML parsing, fuzz testing | `internal/cli/`, `internal/kernel/`, `provenance_fuzz_test.go` |
| 31 | `week6/day31/main.go` | `net/rpc`, JSON-RPC 2.0, protobuf concepts, gRPC overview | `internal/jsonrpc/server.go`, `source_live/lsp/`, `source_live/scip/` |

---

## Java Mental Model Shifts (Read This First)

| Java | Go | Why |
|------|-----|-----|
| Classes | Structs + methods | No inheritance hierarchy |
| Interfaces with `implements` | Implicit interfaces | If it has the methods, it IS the interface |
| try/catch/finally | `error` return values | Errors are data, not exceptions |
| `null` | zero values | Every type has a zero value, no NPE |
| Threads (expensive) | Goroutines (cheap, ~2KB) | You CAN spawn 100k of them |
| `synchronized` | channels or sync.Mutex | Prefer channels for communication |
| Generics (verbose) | Generics (1.18, simpler) | Type inference is good |
| `Optional<T>` | pointer or (T, bool) | Explicit nil or second return |
| `throws IOException` | `(result, error)` return | Caller must handle errors |
| Maven/Gradle | `go mod` | Built-in, much simpler |

---

## Daily Report Format

After each day, copy `progress/template.md` → `progress/YYYY-MM-DD.md` and fill it out.
Report it back to me and I'll adjust the next day's focus.
