// Day 8: Interfaces
// HOW TO RUN: go run week2/day08/main.go
//
// Java dev key shifts:
//   - Interfaces are IMPLICIT — no "implements" keyword needed
//   - If your type has the methods, it satisfies the interface automatically
//   - This is called structural typing or duck typing
//   - interface{} (or any in Go 1.18+) = Java's Object — holds any type
//   - Type assertion: v.(Type)  →  like Java's cast (String) obj
//   - Type switch: switch v := x.(type) { case string: ... }
//   - Interfaces are small in Go — often 1-2 methods (prefer narrow interfaces)

package main

import (
	"fmt"
	"math"
)

// === DEFINING AN INTERFACE ===
// Java: public interface Shape { double area(); double perimeter(); }
// Go:
type Shape interface {
	Area() float64
	Perimeter() float64
}

// === IMPLEMENTING THE INTERFACE (implicitly) ===
// No "implements Shape" needed — just have the methods

type Circle struct{ Radius float64 }
type Rectangle struct{ Width, Height float64 }
type Triangle struct{ A, B, C float64 }

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

func (t Triangle) Area() float64 {
	s := (t.A + t.B + t.C) / 2
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}
func (t Triangle) Perimeter() float64 { return t.A + t.B + t.C }

// Function accepting the interface — works with ANY Shape
// Java: public static void printInfo(Shape s) { ... }
func printShapeInfo(s Shape) {
	fmt.Printf("%T: area=%.2f perimeter=%.2f\n", s, s.Area(), s.Perimeter())
}

// === COMPOSING INTERFACES ===
// Java: interface D extends A, B { }
type Stringer interface {
	String() string
}

type Describable interface {
	Shape
	Stringer // interface embedding
}

// === THE EMPTY INTERFACE ===
// any = interface{} — holds a value of any type
// Java equivalent: Object
func printAny(v any) {
	fmt.Printf("value: %v  type: %T\n", v, v)
}

// === TYPE ASSERTION ===
// Java: if (obj instanceof String str) { ... }
// Go:
func describeType(v any) {
	// Single assertion — panics if wrong type
	// s := v.(string)  // panics if v is not a string

	// Safe assertion with ok check (use this!)
	if s, ok := v.(string); ok {
		fmt.Printf("string: %q (len=%d)\n", s, len(s))
		return
	}
	if n, ok := v.(int); ok {
		fmt.Printf("int: %d\n", n)
		return
	}
	fmt.Printf("unknown: %T\n", v)
}

// === TYPE SWITCH ===
// The idiomatic way to handle multiple types
// Java: if/else instanceof chain
func typeSwitch(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("string of length %d", len(val))
	case int:
		return fmt.Sprintf("int: %d", val)
	case float64:
		return fmt.Sprintf("float64: %.2f", val)
	case bool:
		return fmt.Sprintf("bool: %v", val)
	case []int:
		return fmt.Sprintf("[]int of length %d", len(val))
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("other: %T", val)
	}
}

// === INTERFACE VALUES HOLD (type, value) PAIRS ===
// This is a subtle but important detail

// === PRACTICAL EXAMPLE: Sorter interface ===
type Sortable interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
}

type IntSlice []int

func (s IntSlice) Len() int           { return len(s) }
func (s IntSlice) Less(i, j int) bool { return s[i] < s[j] }
func (s IntSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func bubbleSort(s Sortable) {
	n := s.Len()
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if s.Less(j+1, j) {
				s.Swap(j, j+1)
			}
		}
	}
}

func main() {
	// Polymorphism through interfaces — no inheritance needed
	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 4, Height: 3},
		Triangle{A: 3, B: 4, C: 5},
	}
	for _, s := range shapes {
		printShapeInfo(s)
	}

	// Empty interface
	printAny(42)
	printAny("hello")
	printAny([]int{1, 2, 3})
	printAny(nil)

	// Type assertion
	describeType("go is cool")
	describeType(42)
	describeType(3.14)

	// Type switch
	values := []any{42, "hello", 3.14, true, []int{1, 2}, nil}
	for _, v := range values {
		fmt.Println(typeSwitch(v))
	}

	// Sortable interface
	nums := IntSlice{5, 3, 1, 4, 2}
	bubbleSort(nums)
	fmt.Println("sorted:", []int(nums))

	// Interface nil trap — important!
	// A nil interface value vs an interface holding a nil pointer
	var s Shape // nil interface — both type and value are nil
	fmt.Println("nil interface:", s == nil) // true

	var c *Circle = nil
	var s2 Shape = c    // interface holds (*Circle, nil) — NOT nil!
	fmt.Println("interface with nil pointer:", s2 == nil) // false — GOTCHA!
	// This is a common source of bugs when returning errors
}

// === EXERCISES ===
// 1. Create a Logger interface with Log(msg string) and a FileLogger and
//    ConsoleLogger that implement it. Write a function that accepts Logger.
//
// 2. Create a PaymentProcessor interface with Charge(amount float64) error.
//    Implement MockProcessor (always succeeds) and FailingProcessor (always fails).
//    Write a Checkout function that uses PaymentProcessor.
//
// 3. The io.Writer interface is just: Write(p []byte) (n int, err error)
//    Create a LineCounter that implements io.Writer and counts newlines.
//    Use fmt.Fprintln(yourLineCounter, "hello\nworld") to test it.
//
// 4. Explain the nil interface trap from the demo above.
//    When returning errors, why should you return nil directly
//    instead of returning a typed nil (*MyError)(nil)?
//
// 5. Write a function that takes []any and returns counts by type:
//    map[string]int{"string": 2, "int": 3, "float64": 1}
