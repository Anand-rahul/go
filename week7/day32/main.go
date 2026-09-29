// Day 32: Fibonacci exercise
// HOW TO RUN: go run week7/day32/main.go

package main

import "fmt"

// Fibonacci returns the nth Fibonacci number (0-indexed).
func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func main() {
	for i := 0; i < 15; i++ {
		fmt.Printf("Fibonacci(%d) = %d\n", i, Fibonacci(i))
	}
}

// ============================================================
// EXERCISES
// ============================================================
//
// Exercise 1: Memoized Fibonacci
//   Write FibonacciMemo(n int, cache map[int]int) int that caches
//   previously computed values to avoid recomputation.
//
// Exercise 2: Fibonacci sequence generator
//   Write a function that returns a []int of the first n Fibonacci numbers.
//
// Exercise 3: Concurrent Fibonacci
//   Using goroutines and channels, compute Fibonacci(30) through Fibonacci(40)
//   concurrently and print results as they complete.
