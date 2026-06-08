// Day 21: Generics — Type Parameters, Constraints, Type Sets
// HOW TO RUN: go run week5/day21/main.go
// Requires Go 1.18+
//
// Java dev key shifts:
//   - Go generics arrived in 1.18 — similar to Java generics but with key differences
//   - Syntax: func Fn[T constraint](arg T) T
//   - Constraints are interfaces that specify what operations T supports
//   - comparable = types that support == and != (like Java's Comparable but for equality)
//   - any = interface{} — accepts all types but you can do nothing with T besides store it
//   - No wildcards (no ? extends T or ? super T) — use type unions instead
//   - Type inference works well — you often don't need to specify type params

package main

import (
	"cmp"
	"fmt"
)

// === BASIC GENERIC FUNCTION ===
// Java: public static <T> T identity(T value) { return value; }
// Go:
func identity[T any](v T) T {
	return v
}

// Map: apply a function to each element, return new slice
// Java: list.stream().map(fn).collect(toList())
func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter: keep elements matching predicate
func Filter[T any](slice []T, pred func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if pred(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce: fold a slice into a single value
// Java: stream.reduce(identity, accumulator)
func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	acc := initial
	for _, v := range slice {
		acc = fn(acc, v)
	}
	return acc
}

// === COMPARABLE CONSTRAINT ===
// comparable: types that support == and !=
// Use for maps, sets, dedup, search

func Contains[T comparable](slice []T, target T) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	var result []T
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// === cmp.Ordered CONSTRAINT ===
// cmp.Ordered: types that support <, >, <=, >= (int, float64, string, etc.)
// Java: Comparable<T>

func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func MinInSlice[T cmp.Ordered](slice []T) T {
	m := slice[0]
	for _, v := range slice[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

// === CUSTOM CONSTRAINT ===
// A constraint is just an interface with a type set
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

// ~ means "this type or any type whose underlying type is T"
// This allows user-defined types like: type MyInt int

func Sum[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

// === GENERIC STRUCT ===
// Java: class Stack<T> { }
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	last := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return last, true
}

func (s *Stack[T]) Len() int { return len(s.items) }

// Generic Pair
type Pair[A, B any] struct {
	First  A
	Second B
}

func NewPair[A, B any](a A, b B) Pair[A, B] {
	return Pair[A, B]{First: a, Second: b}
}

// Zip two slices into pairs
func Zip[A, B any](as []A, bs []B) []Pair[A, B] {
	n := len(as)
	if len(bs) < n {
		n = len(bs)
	}
	result := make([]Pair[A, B], n)
	for i := 0; i < n; i++ {
		result[i] = NewPair(as[i], bs[i])
	}
	return result
}

// === TYPE CONSTRAINTS WITH METHODS ===
type Stringer interface {
	String() string
}

func PrintAll[T Stringer](items []T) {
	for _, item := range items {
		fmt.Println(item.String())
	}
}

// Custom type that satisfies Number constraint
type Celsius float64

func (c Celsius) String() string { return fmt.Sprintf("%.1f°C", float64(c)) }

func main() {
	// Basic
	fmt.Println(identity("hello"))
	fmt.Println(identity(42))
	fmt.Println(identity(3.14))

	// Map
	nums := []int{1, 2, 3, 4, 5}
	doubled := Map(nums, func(n int) int { return n * 2 })
	strs := Map(nums, func(n int) string { return fmt.Sprintf("n=%d", n) })
	fmt.Println("doubled:", doubled)
	fmt.Println("as strings:", strs)

	// Filter
	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Println("evens:", evens)

	// Reduce
	sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	product := Reduce(nums, 1, func(acc, n int) int { return acc * n })
	fmt.Println("sum:", sum, "product:", product)

	// Contains
	fmt.Println("contains 3:", Contains(nums, 3))
	fmt.Println("contains 'go':", Contains([]string{"go", "java", "rust"}, "go"))

	// Unique
	dupes := []int{1, 2, 2, 3, 3, 3, 4}
	fmt.Println("unique:", Unique(dupes))
	fmt.Println("unique strings:", Unique([]string{"a", "b", "a", "c", "b"}))

	// Min/Max
	fmt.Println("min(3,7):", Min(3, 7))
	fmt.Println("max(3,7):", Max(3, 7))
	fmt.Println("min string:", Min("banana", "apple"))
	fmt.Println("min in slice:", MinInSlice([]float64{3.14, 2.71, 1.41}))

	// Sum with custom Number type
	fmt.Println("sum ints:", Sum([]int{1, 2, 3, 4, 5}))
	fmt.Println("sum float64:", Sum([]float64{1.1, 2.2, 3.3}))

	// ~ operator: Celsius is float64 underneath, satisfies Number
	temps := []Celsius{100, 37, -40, 20}
	fmt.Println("sum temps:", Sum(temps))

	// Generic Stack
	s := &Stack[string]{}
	s.Push("go")
	s.Push("is")
	s.Push("great")
	for s.Len() > 0 {
		v, _ := s.Pop()
		fmt.Print(v, " ")
	}
	fmt.Println()

	// Generic Pair
	p := NewPair("age", 28)
	fmt.Printf("pair: %v → %v\n", p.First, p.Second)

	// Zip
	names := []string{"Alice", "Bob", "Charlie"}
	scores := []int{90, 85, 92}
	zipped := Zip(names, scores)
	for _, pair := range zipped {
		fmt.Printf("  %s: %d\n", pair.First, pair.Second)
	}
}

// === EXERCISES ===
// 1. Write generic func GroupBy[K comparable, V any](slice []V, key func(V) K) map[K][]V
//    Use it to group []string by first letter.
//
// 2. Write a generic OrderedMap[K cmp.Ordered, V any] backed by two slices:
//    keys []K and values []V (kept in sorted key order).
//    Methods: Set, Get, Keys, Values.
//
// 3. Write func First[T any](slice []T, pred func(T) bool) (T, bool)
//    Returns the first element matching the predicate, or zero value + false.
//
// 4. Write func Must[T any](v T, err error) T
//    Panics if err != nil, otherwise returns v.
//    Usage: data := Must(os.ReadFile("file.txt"))
//
// 5. Compare Go generics to Java generics:
//    a) Java has type erasure at runtime. Does Go?
//    b) Java has wildcards (? extends T). What's the Go equivalent?
//    c) Can you instantiate T in Go? In Java?
