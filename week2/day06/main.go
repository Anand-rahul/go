// Day 6: Functions — Multiple Returns, Variadic, Closures, First-Class
// HOW TO RUN: go run week2/day06/main.go
//
// Java dev key shifts:
//   - Functions can return multiple values — idiomatic for (result, error)
//   - Named return values — can be used but use sparingly
//   - Variadic: func f(args ...int)  →  like Java's f(int... args)
//   - Functions are first-class values — assign to variables, pass as args
//   - Closures capture variables by reference (same as Java lambdas)
//   - No method overloading — use different names or variadic

package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

// === BASIC FUNCTION ===
// Java: public static int add(int a, int b) { return a + b; }
func add(a, b int) int { // consecutive same-type params share type declaration
	return a + b
}

// === MULTIPLE RETURN VALUES ===
// This is how Go handles errors instead of try/catch
// Java: throws Exception — caller can ignore. Go: caller MUST handle (or explicitly ignore)
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// === NAMED RETURN VALUES ===
// Useful for documentation, but avoid naked returns in long functions
func minMax(nums []int) (min, max int) {
	min, max = nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return // naked return — returns min and max by name
}

// === VARIADIC FUNCTIONS ===
// Java: public static int sum(int... nums)
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// === FUNCTIONS AS VALUES ===
// Java: Function<Integer, Integer> f = x -> x * 2;
// Go:
type transformer func(int) int

func apply(nums []int, fn transformer) []int {
	result := make([]int, len(nums))
	for i, v := range nums {
		result[i] = fn(v)
	}
	return result
}

// === CLOSURES ===
// A function that captures variables from its surrounding scope
// Java: lambdas capture effectively-final variables. Go: captures by reference.
func makeCounter() func() int {
	count := 0
	return func() int {
		count++ // captures 'count' from makeCounter's scope
		return count
	}
}

// makeAdder — classic closure example
func makeAdder(n int) func(int) int {
	return func(x int) int {
		return x + n // captures 'n'
	}
}

// === FUNCTION THAT RETURNS FUNCTION (higher-order) ===
func memoize(fn func(int) int) func(int) int {
	cache := make(map[int]int)
	return func(n int) int {
		if v, ok := cache[n]; ok {
			return v
		}
		result := fn(n)
		cache[n] = result
		return result
	}
}

func main() {
	// Basic
	fmt.Println("add(3,4):", add(3, 4))

	// Multiple returns
	result, err := divide(10, 3)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Printf("10/3 = %.4f\n", result)
	}

	_, err = divide(5, 0)
	if err != nil {
		fmt.Println("error:", err)
	}

	// Named returns
	nums := []int{3, 1, 4, 1, 5, 9, 2, 6}
	min, max := minMax(nums)
	fmt.Printf("min=%d max=%d\n", min, max)

	// Variadic — call with individual args
	fmt.Println("sum:", sum(1, 2, 3, 4, 5))

	// Variadic — spread a slice with ...
	values := []int{10, 20, 30}
	fmt.Println("sum from slice:", sum(values...))

	// Functions as values
	double := func(x int) int { return x * 2 }
	square := func(x int) int { return x * x }

	data := []int{1, 2, 3, 4, 5}
	fmt.Println("doubled:", apply(data, double))
	fmt.Println("squared:", apply(data, square))

	// Inline anonymous function
	fmt.Println("abs:", apply([]int{-3, -1, 2, -5}, func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}))

	// Closures
	counter1 := makeCounter()
	counter2 := makeCounter() // independent counter
	fmt.Println(counter1(), counter1(), counter1()) // 1 2 3
	fmt.Println(counter2(), counter2())             // 1 2 — own state

	add5 := makeAdder(5)
	add10 := makeAdder(10)
	fmt.Println("add5(3):", add5(3))   // 8
	fmt.Println("add10(3):", add10(3)) // 13

	// Memoization via closure
	slowSqrt := func(n int) int {
		return int(math.Sqrt(float64(n)))
	}
	fastSqrt := memoize(slowSqrt)
	fmt.Println(fastSqrt(16), fastSqrt(25), fastSqrt(16)) // 16 is cached

	// Immediately invoked function (like Java's anonymous blocks)
	result2 := func(x, y int) int {
		return x * y
	}(6, 7)
	fmt.Println("6*7:", result2)

	// defer — runs at function exit (covered deeply in Day 22)
	defer fmt.Println("this runs last (deferred)")
	fmt.Println("this runs first")

	fmt.Println("Ex1")
	tests := []string{"25", "abc", "-5"}

	for _, t := range tests {
		age, err := parseAge(t)
		if err != nil {
			fmt.Printf("parseAge(%q) failed: %v\n", t, err)
			continue
		}

		fmt.Printf("parseAge(%q) = %d\n", t, age)
	}
	// 2. Write: func filter(nums []int, pred func(int) bool) []int
	//    Use it to filter even numbers from [1,2,3,4,5,6,7,8].

	fmt.Println("Ex2")
	newNums := []int{1, 2, 3, 4, 5, 6, 7, 8}

	evens:= filter(newNums, func(n int) bool {
		return n%2 == 0
	})

	fmt.Println(evens)

	fmt.Println("Ex3")
	// 3. Write: func pipeline(fns ...func(int) int) func(int) int
	//    that applies each function in sequence to an input.
	//    pipeline(double, square)(3) → square(double(3)) → square(6) → 36
	p := pipeline(double, square)
	fmt.Println(p(3))

	fmt.Println(square(double(3)))
	fmt.Println(square(6))


	fmt.Println("Ex4")
	// 4. Write a closure-based rate limiter:
	//    func makeRateLimiter(limit int) func() bool
	//    Returns true up to 'limit' times, then always false.
	limiter := makeRateLimiter(3)

	fmt.Println(limiter())
	fmt.Println(limiter())
	fmt.Println(limiter())
	fmt.Println(limiter())
	fmt.Println(limiter())

	fmt.Println("Ex5")
	// 5. Java gotcha: lambdas capture effectively-final vars. Go captures by reference.
	//    What does this print?
	//    funcs := make([]func(), 3)
	//    for i := 0; i < 3; i++ { funcs[i] = func() { fmt.Println(i) } }
	//    for _, f := range funcs { f() }
	//    Fix it so each func prints its own i.
	funcs := make([]func(), 3)

	for i := 0; i < 3; i++ {
		funcs[i] = func() {
			fmt.Println(i)
		}
	}

	for _, f := range funcs {
		f()
	}
}

func makeRateLimiter(limit int) func() bool {
	count := 0

	return func() bool {
		if count >= limit {
			return false
		}

		count++
		return true
	}
}

func pipeline(fns ...func(int) int) func(int) int{
	return func(x int) int {
		result := x

		for _, fn := range fns {
			result = fn(result)
		}

		return result
	}
}
func double(x int) int {
	return x * 2
}

func square(x int) int {
	return x * x
}
func parseAge(s string) (int, error) {
	return strconv.Atoi(s)
}

func filter(nums[] int , pred func(int) bool)[]int{
	result:=make([]int,0)

	for _,num := range nums{
		if pred(num){
			result=append(result,num)
		}
	}
	return result
}

// === EXERCISES ===
// 1. Write: func parseAge(s string) (int, error)
//    Use strconv.Atoi. Return the error from Atoi directly.
//    Call it with "25", "abc", and "-5".
//
// 2. Write: func filter(nums []int, pred func(int) bool) []int
//    Use it to filter even numbers from [1,2,3,4,5,6,7,8].
//
// 3. Write: func pipeline(fns ...func(int) int) func(int) int
//    that applies each function in sequence to an input.
//    pipeline(double, square)(3) → square(double(3)) → square(6) → 36
//
// 4. Write a closure-based rate limiter:
//    func makeRateLimiter(limit int) func() bool
//    Returns true up to 'limit' times, then always false.
//
// 5. Java gotcha: lambdas capture effectively-final vars. Go captures by reference.
//    What does this print?
//    funcs := make([]func(), 3)
//    for i := 0; i < 3; i++ { funcs[i] = func() { fmt.Println(i) } }
//    for _, f := range funcs { f() }
//    Fix it so each func prints its own i.
