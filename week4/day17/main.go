// Day 17: Testing — Table-Driven Tests, Subtests, Benchmarks, Test Helpers
// HOW TO RUN: go test ./week4/day17/...  (not go run — this file has no main logic)
// HOW TO RUN: go test -v ./week4/day17/...
// HOW TO RUN: go test -run TestDivide ./week4/day17/...
// HOW TO RUN: go test -bench=. ./week4/day17/...
//
// Java dev key shifts:
//   - Test file name MUST end in _test.go
//   - Test functions MUST start with Test (capital T)
//   - Receive *testing.T parameter — not annotations like @Test
//   - No assertEquals — use t.Errorf or t.Fatalf directly
//   - Table-driven tests are idiomatic Go (no @ParameterizedTest)
//   - Subtests: t.Run("name", func) — like JUnit's @Nested
//   - Benchmarks: func BenchmarkXxx(b *testing.B) — built in, no JMH needed
//   - go test compiles and runs automatically — no test runner setup

package main

import "fmt"

// === CODE TO TEST ===
// In real projects this would be in a separate file (not _test.go)

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func isPalindrome(s string) bool {
	return s == reverseString(s)
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
	// This file is about tests — run with 'go test' not 'go run'
	// The actual test code is in day17_test.go
	fmt.Println("Run: go test -v ./week4/day17/...")
	fmt.Println("Run: go test -bench=. ./week4/day17/...")
	fmt.Println("Run: go test -race ./week4/day17/...")
}
