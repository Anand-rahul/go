// Break Day 1: Functions — Deeper Practice
// HOW TO RUN: go run week2/break01/main.go
//
// These exercises reinforce Day 6 patterns you'll use in real Go codebases:
//   - once: run something exactly once (think: lazy init, sync.Once by hand)
//   - retry: resilience pattern — retrying flaky operations
//   - curry: transform multi-arg functions into chains of single-arg functions
//   - compose: mathematical function composition
//   - debounce (conceptual): understand how closures hold state across calls
//
// All of these are closure + first-class function patterns.
// Java devs often reach for classes/singletons for this — Go uses closures.

package main

import (
	"errors"
	"fmt"
	"strings"
)

// =====================================================================
// EXERCISE 1: once
// =====================================================================
// Write a function:
//   func once(fn func()) func()
//
// Returns a wrapper that calls fn exactly once.
// Any subsequent calls do nothing.
//
// Real use case: lazy initialization, setup that must run once.
// (Go's sync.Once does this thread-safely — this is the manual version)
//
// Expected output:
//   initialized!
//   (nothing printed for the 2nd and 3rd calls)

func once(fn func()) func() {
	// TODO: implement
	// Hint: use a bool in a closure to track if fn has been called
	return func() {}
}

// =====================================================================
// EXERCISE 2: retry
// =====================================================================
// Write a function:
//   func retry(attempts int, fn func() error) error
//
// Calls fn up to `attempts` times.
// Returns nil as soon as fn succeeds (returns nil error).
// Returns the last error if all attempts fail.
//
// Real use case: flaky network calls, transient DB errors.
//
// Expected output:
//   attempt 1 failed
//   attempt 2 failed
//   attempt 3 succeeded
//   result: nil
//   all 3 attempts failed: always fails

func retry(attempts int, fn func() error) error {
	// TODO: implement
	return errors.New("not implemented")
}

// =====================================================================
// EXERCISE 3: curry
// =====================================================================
// Write a function:
//   func curry(fn func(int, int) int) func(int) func(int) int
//
// Converts a 2-argument function into a chain of single-argument functions.
// curry(add)(3)(4) == add(3, 4) == 7
//
// Real use case: partial application — fix one argument, reuse with many.
//
// Expected output:
//   curry(add)(3)(4) = 7
//   addThree(10) = 13
//   addThree(20) = 23
//   addThree(30) = 33

func curry(fn func(int, int) int) func(int) func(int) int {
	// TODO: implement
	return func(a int) func(int) int {
		return func(b int) int {
			return 0
		}
	}
}

// =====================================================================
// EXERCISE 4: compose
// =====================================================================
// Write a function:
//   func compose(fns ...func(string) string) func(string) string
//
// Returns a new function that applies each fn RIGHT to LEFT (math convention).
// compose(f, g, h)(x) == f(g(h(x)))
//
// Real use case: building transformation pipelines (middleware chains, text processing).
//
// Expected output:
//   composed("  hello world  ") = "HELLO WORLD"
//   (trim spaces → title case → upper case, applied right to left)

func compose(fns ...func(string) string) func(string) string {
	// TODO: implement
	// Hint: iterate fns in reverse order
	return func(s string) string {
		return s
	}
}

// =====================================================================
// EXERCISE 5: memoize with cache stats
// =====================================================================
// Extend the memoize pattern from Day 6:
//   func memoizeWithStats(fn func(int) int) (func(int) int, func() (hits, misses int))
//
// Returns TWO functions:
//   1. The memoized version of fn
//   2. A stats function that returns (cacheHits, cacheMisses)
//
// Expected output:
//   fib(10) = 55
//   fib(10) = 55   (cached)
//   fib(8)  = 21   (partially cached — 8 and 9 already computed during fib(10))
//   hits: 1, misses: 11   (exact numbers depend on your fib implementation)
//
// Note: for this exercise, implement a simple iterative fibonacci, not recursive.
// func fib(n int) int — iterative, returns nth fibonacci number (0-indexed: fib(0)=0, fib(1)=1, fib(10)=55)

func memoizeWithStats(fn func(int) int) (func(int) int, func() (int, int)) {
	// TODO: implement
	// Hint: three captured variables — cache map, hits int, misses int
	memo := func(n int) int { return fn(n) }
	stats := func() (int, int) { return 0, 0 }
	return memo, stats
}

// =====================================================================
// HELPERS (provided — do not modify)
// =====================================================================

func addInts(a, b int) int { return a + b }

func fib(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func trimSpaces(s string) string  { return strings.TrimSpace(s) }
func toUpper(s string) string     { return strings.ToUpper(s) }
func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}

// =====================================================================
// MAIN — do not modify, just implement the functions above
// =====================================================================

func main() {
	fmt.Println("=== Exercise 1: once ===")
	init := once(func() { fmt.Println("initialized!") })
	init() // prints
	init() // silent
	init() // silent
	fmt.Println()

	fmt.Println("=== Exercise 2: retry ===")
	callCount := 0
	flaky := func() error {
		callCount++
		if callCount < 3 {
			fmt.Printf("  attempt %d failed\n", callCount)
			return errors.New("not ready yet")
		}
		fmt.Printf("  attempt %d succeeded\n", callCount)
		return nil
	}
	err := retry(5, flaky)
	fmt.Println("result:", err)

	callCount = 0
	alwaysFails := func() error {
		callCount++
		return errors.New("always fails")
	}
	err = retry(3, alwaysFails)
	fmt.Println("all 3 attempts failed:", err)
	fmt.Println()

	fmt.Println("=== Exercise 3: curry ===")
	curriedAdd := curry(addInts)
	fmt.Println("curry(add)(3)(4) =", curriedAdd(3)(4))
	addThree := curriedAdd(3)
	fmt.Println("addThree(10) =", addThree(10))
	fmt.Println("addThree(20) =", addThree(20))
	fmt.Println("addThree(30) =", addThree(30))
	fmt.Println()

	fmt.Println("=== Exercise 4: compose ===")
	// Applied right-to-left: trimSpaces first, then toTitleCase, then toUpper
	transform := compose(toUpper, toTitleCase, trimSpaces)
	fmt.Println(`composed("  hello world  ") =`, transform("  hello world  "))
	fmt.Println()

	fmt.Println("=== Exercise 5: memoize with stats ===")
	memoFib, stats := memoizeWithStats(fib)
	fmt.Println("fib(10) =", memoFib(10))
	fmt.Println("fib(10) =", memoFib(10)) // cache hit
	fmt.Println("fib(8)  =", memoFib(8))  // cache hit (computed indirectly? no — fib is iterative here, each n is independent)
	h, m := stats()
	fmt.Printf("hits: %d, misses: %d\n", h, m)
}
