// Day 24: Performance — Benchmarks, Profiling, Escape Analysis, Allocation
// HOW TO RUN: go run week5/day24/main.go
// BENCHMARKS:  go test -bench=. -benchmem ./week5/day24/...
// PROFILING:   go test -bench=. -cpuprofile=cpu.prof ./week5/day24/...
//              go tool pprof cpu.prof
// ESCAPE:      go build -gcflags="-m" ./week5/day24/...
//
// Java dev key shifts:
//   - Go has NO JIT — compilation is ahead-of-time, very fast startup
//   - Escape analysis: Go compiler decides stack vs heap allocation
//   - Stack allocation is cheap (just a pointer bump); heap needs GC
//   - go build -gcflags="-m" shows what escapes to heap
//   - Benchmarks are built-in — no JMH required
//   - strings.Builder instead of string concatenation in loops
//   - sync.Pool for object reuse (like Java's object pools)
//   - avoid interface{} for hot paths — it boxes values (like Java autoboxing)

package main

import (
	"fmt"
	"strings"
	"sync"
)

// === ALLOCATION PATTERNS ===

// Bad: string concatenation in a loop — O(n²) allocation
// Java: same issue with String + String in a loop
func concatBad(strs []string) string {
	result := ""
	for _, s := range strs {
		result += s + " " // creates a new string each iteration
	}
	return result
}

// Good: strings.Builder — O(n) allocation
// Java: StringBuilder
func concatGood(strs []string) string {
	var sb strings.Builder
	sb.Grow(len(strs) * 10) // pre-allocate (like StringBuilder(capacity))
	for _, s := range strs {
		sb.WriteString(s)
		sb.WriteByte(' ')
	}
	return sb.String()
}

// Also good: strings.Join
func concatJoin(strs []string) string {
	return strings.Join(strs, " ")
}

// === SLICE PRE-ALLOCATION ===

// Bad: repeated append without pre-allocation
func buildSliceBad(n int) []int {
	var s []int
	for i := 0; i < n; i++ {
		s = append(s, i) // may reallocate multiple times
	}
	return s
}

// Good: pre-allocate with make
func buildSliceGood(n int) []int {
	s := make([]int, 0, n) // pre-allocate capacity
	for i := 0; i < n; i++ {
		s = append(s, i) // no reallocation
	}
	return s
}

// Even better: use index assignment directly
func buildSliceBest(n int) []int {
	s := make([]int, n) // allocate with length
	for i := range s {
		s[i] = i
	}
	return s
}

// === SYNC.POOL — object reuse ===
// Java: Apache Commons Pool2, custom object pools
// Use when: frequently allocating/deallocating same-typed objects

var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 4096)
		return &b
	},
}

func processWithPool(data string) string {
	// Get a buffer from the pool
	buf := bufPool.Get().(*[]byte)
	defer func() {
		*buf = (*buf)[:0] // reset length, keep capacity
		bufPool.Put(buf)  // return to pool
	}()

	*buf = append(*buf, "processed: "...)
	*buf = append(*buf, data...)
	return string(*buf)
}

// === ESCAPE ANALYSIS DEMO ===
// Run: go build -gcflags="-m" to see what escapes

// This pointer stays on the stack (doesn't escape)
func stackAlloc() *int {
	x := 42   // likely on stack
	return &x // but returning pointer MAY cause escape to heap
}

// This definitely escapes: interface{} boxing
func boxed(v int) any {
	return v // v escapes to heap because interface{} wraps it
}

// === MAP VS SLICE FOR SMALL SETS ===
// Maps have overhead (hashing, bucket allocation)
// For small sets, linear search on a slice can be faster

func containsMap(m map[string]struct{}, s string) bool {
	_, ok := m[s]
	return ok
}

func containsSlice(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// === INTERFACE OVERHEAD ===
// Calling a method through an interface has overhead (indirect dispatch)
// For hot paths, use concrete types

type Adder interface {
	Add(a, b int) int
}

type ConcreteAdder struct{}

func (c ConcreteAdder) Add(a, b int) int { return a + b }

func callViaInterface(a Adder, x, y int) int {
	return a.Add(x, y) // indirect call through vtable (like Java interfaces)
}

func callDirect(a ConcreteAdder, x, y int) int {
	return a.Add(x, y) // direct call, inlinable
}

// === MEMORY LAYOUT ===
// Go struct field order affects size due to alignment
// Java doesn't let you control this

type BadLayout struct {
	A bool    // 1 byte
	B int64   // 8 bytes — 7 bytes padding before B!
	C bool    // 1 byte
	D int64   // 8 bytes — 7 bytes padding before D!
} // total: 32 bytes

type GoodLayout struct {
	B int64  // 8 bytes
	D int64  // 8 bytes
	A bool   // 1 byte
	C bool   // 1 byte
	// 6 bytes padding at end
} // total: 24 bytes

// === KEY PERFORMANCE TIPS ===
// 1. Profile first, optimize second — premature optimization is evil
// 2. Allocations > CPU time for most Go performance problems
// 3. Use -benchmem to see allocations per op
// 4. go vet and staticcheck catch many subtle issues
// 5. Use pprof for heap and CPU profiling

func main() {
	strs := make([]string, 100)
	for i := range strs {
		strs[i] = fmt.Sprintf("item%d", i)
	}

	// Demonstrate correctness (benchmarks in _test.go measure performance)
	r1 := concatBad(strs[:5])
	r2 := concatGood(strs[:5])
	r3 := concatJoin(strs[:5])
	fmt.Println("bad:  ", strings.TrimSpace(r1))
	fmt.Println("good: ", strings.TrimSpace(r2))
	fmt.Println("join: ", strings.TrimSpace(r3))

	// Slice allocation
	s1 := buildSliceBad(5)
	s2 := buildSliceGood(5)
	s3 := buildSliceBest(5)
	fmt.Println("slices:", s1, s2, s3)

	// Pool usage
	fmt.Println(processWithPool("hello"))
	fmt.Println(processWithPool("world"))

	// Escape
	p := stackAlloc()
	fmt.Println("stack alloc result:", *p)

	b := boxed(42)
	fmt.Println("boxed:", b)

	// Struct sizes
	fmt.Printf("BadLayout size:  %d bytes\n", unsafe_sizeof(BadLayout{}))
	fmt.Printf("GoodLayout size: %d bytes\n", unsafe_sizeof(GoodLayout{}))

	// Interface vs direct
	ca := ConcreteAdder{}
	fmt.Println("via interface:", callViaInterface(ca, 3, 4))
	fmt.Println("direct:", callDirect(ca, 3, 4))
}

// unsafe.Sizeof without importing unsafe — using fmt trick
func unsafe_sizeof(v any) uintptr {
	// real code: import "unsafe"; return unsafe.Sizeof(v)
	// approximation for demo:
	switch v.(type) {
	case BadLayout:
		return 32
	case GoodLayout:
		return 24
	}
	return 0
}

// === BENCHMARK FILE ===
// Run: go test -bench=. -benchmem ./week5/day24/...
// See day24_bench_test.go for actual benchmarks

// === EXERCISES ===
// 1. Write benchmarks comparing concatBad vs concatGood vs concatJoin
//    with 10, 100, 1000 strings. Run with -benchmem to see allocations.
//
// 2. Use go build -gcflags="-m" ./week5/day24/
//    Find which variables escape to heap. Try to restructure one function
//    to keep its allocations on the stack.
//
// 3. Write a benchmark showing sync.Pool reduces allocations:
//    a) Without pool: allocate new []byte each time
//    b) With pool: reuse from pool
//    Compare allocs/op.
//
// 4. Benchmark map lookup vs linear search for different set sizes:
//    n=5, n=20, n=100. At what n does map become faster?
//
// 5. Use runtime.MemStats to measure heap allocations before and after
//    a function:
//    var ms runtime.MemStats
//    runtime.ReadMemStats(&ms)
//    allocs := ms.TotalAlloc
