// Day 22: defer, panic, recover
// HOW TO RUN: go run week5/day22/main.go
//
// Java dev key shifts:
//   - defer = try-finally: deferred calls run when function returns (any path)
//   - panic = throwing a RuntimeException that unwinds the stack
//   - recover = catch in a deferred function — ONLY works in a defer
//   - Go philosophy: use errors for expected failures, panic only for unrecoverable bugs
//   - defer runs LIFO (last in, first out) — like nested finally blocks
//   - defer evaluates its ARGUMENTS immediately but runs the function later

package main

import (
	"fmt"
	"os"
)

// === DEFER BASICS ===
// Java: try { ... } finally { cleanup(); }
// Go:   defer cleanup()  — declared next to the resource acquisition

func deferBasic() {
	fmt.Println("start")
	defer fmt.Println("deferred 1 (runs last)")
	defer fmt.Println("deferred 2 (runs second)")
	defer fmt.Println("deferred 3 (runs first)")
	fmt.Println("end")
	// Output order: start, end, deferred 3, deferred 2, deferred 1 (LIFO)
}

// === DEFER FOR RESOURCE CLEANUP ===
// The canonical pattern: defer close right after open
func readFileDemo(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close() // guaranteed to run even if function panics

	// ... read from f ...
	fmt.Println("file opened:", filename)
	return nil
}

// === DEFER WITH LOOPS — a gotcha ===
func deferInLoop() {
	// Wrong: all defers stack up, resources held until function returns
	// for i := 0; i < 3; i++ {
	//     f, _ := os.Open(...)
	//     defer f.Close() // ALL files stay open until function ends!
	// }

	// Right: wrap in a function to close after each iteration
	for i := 0; i < 3; i++ {
		func() {
			defer fmt.Printf("closing resource %d\n", i)
			fmt.Printf("using resource %d\n", i)
		}()
	}
}

// === DEFER ARGUMENT EVALUATION ===
// Arguments to deferred functions are evaluated IMMEDIATELY, not when defer runs
func deferEvaluation() {
	x := 1
	defer fmt.Println("deferred x:", x) // x is captured as 1 RIGHT NOW
	x = 100
	fmt.Println("current x:", x)
	// deferred prints 1, not 100!

	// But if you use a closure, it captures by reference:
	y := 1
	defer func() {
		fmt.Println("closure y:", y) // sees y=100 because closure captures reference
	}()
	y = 100
}

// === NAMED RETURN VALUES + DEFER ===
// Deferred functions CAN modify named return values (rare but useful)
func divide(a, b float64) (result float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	if b == 0 {
		panic("division by zero") // will be caught by recover above
	}
	result = a / b
	return
}

// === PANIC ===
// Like throwing a RuntimeException that isn't meant to be caught normally
// Use ONLY for truly unrecoverable situations:
//   - nil pointer dereference (happens automatically)
//   - index out of bounds (happens automatically)
//   - logic errors that should never happen (programmer errors)

func mustPositive(n int) int {
	if n <= 0 {
		panic(fmt.Sprintf("mustPositive: got %d, expected positive", n))
	}
	return n
}

// === RECOVER ===
// Must be called in a DEFERRED function to catch panics
// Like catch(RuntimeException e) — but more explicit

func safeDiv(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	result = a / b // panics if b == 0
	return
}

// Recover all panics from a goroutine
func safeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("goroutine panic recovered: %v\n", r)
			}
		}()
		fn()
	}()
}

// === CLEANUP PATTERN (Go 1.21+: slices.Sort, etc.) ===
// The canonical setup/teardown pattern without TestMain

type testDB struct{ name string }

func setupDB() *testDB {
	fmt.Println("setting up DB")
	return &testDB{"mydb"}
}

func teardownDB(db *testDB) {
	fmt.Println("tearing down DB:", db.name)
}

func withDB(fn func(*testDB)) {
	db := setupDB()
	defer teardownDB(db)
	fn(db)
}

// === PANIC IN PRACTICE — init() validation ===
// Acceptable panics: startup, init(), programmer contract violations
var config = mustLoadConfig()

func mustLoadConfig() string {
	val := "production"
	if val == "" {
		panic("config: required env var ENVIRONMENT is not set")
	}
	return val
}

func main() {
	fmt.Println("=== DEFER BASICS (LIFO) ===")
	deferBasic()

	fmt.Println("\n=== DEFER IN LOOP ===")
	deferInLoop()

	fmt.Println("\n=== DEFER ARGUMENT EVALUATION ===")
	deferEvaluation()

	fmt.Println("\n=== RECOVER IN DEFERRED FUNCTION ===")
	result, err := divide(10, 0)
	fmt.Printf("divide(10,0): result=%v err=%v\n", result, err)
	result, err = divide(10, 2)
	fmt.Printf("divide(10,2): result=%v err=%v\n", result, err)

	fmt.Println("\n=== SAFE INTEGER DIVISION ===")
	r, e := safeDiv(10, 2)
	fmt.Printf("safeDiv(10,2): %d %v\n", r, e)
	r, e = safeDiv(10, 0)
	fmt.Printf("safeDiv(10,0): %d %v\n", r, e)

	fmt.Println("\n=== SAFE GOROUTINE ===")
	safeGo(func() {
		panic("goroutine panicked!")
	})
	fmt.Println("main continues after goroutine panic (because safeGo recovered it)")

	fmt.Println("\n=== WITH PATTERN ===")
	withDB(func(db *testDB) {
		fmt.Printf("using DB: %s\n", db.name)
		// DB is automatically torn down when this function returns
	})

	fmt.Println("\n=== PANIC PROPAGATION ===")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("outer recover:", r)
			}
		}()

		func() {
			// No recover here — panic bubbles up
			panic("panic from inner function")
		}()

		fmt.Println("this never runs")
	}()

	fmt.Println("config loaded:", config)
}

// === EXERCISES ===
// 1. Write a function withTransaction(db *DB, fn func(*DB) error) error that:
//    - Opens a transaction (simulate with print)
//    - Defers a rollback
//    - If fn() succeeds, commits instead of rolling back
//    Hint: use named return to change defer behavior
//
// 2. Write a safe wrapper for slice indexing:
//    func safeGet[T any](slice []T, i int) (val T, err error)
//    Use recover to catch index-out-of-bounds panics.
//
// 3. Write a middleware that recovers from panics in HTTP handlers:
//    func recoverMiddleware(next http.Handler) http.Handler
//    Log the panic, return 500, then continue.
//
// 4. The defer order gotcha:
//    for i := 0; i < 3; i++ {
//        defer fmt.Println(i)
//    }
//    What does this print? Why? Fix it to print 0, 1, 2 in order.
//
// 5. When is it appropriate to NOT recover from a panic?
//    Give 3 examples of panics you should let propagate vs recover.
