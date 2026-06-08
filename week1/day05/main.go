// Day 5: Pointers
// HOW TO RUN: go run week1/day05/main.go
//
// Java dev key shifts:
//   - Java hides pointers — every object IS a reference under the hood
//   - Go makes pointers explicit with & and *
//   - & = "address of" (gives you a pointer)
//   - * = "dereference" (get the value a pointer points to)
//   - new(T) allocates and returns *T (like Java's new, but returns a pointer)
//   - No pointer arithmetic (unlike C) — safe
//   - nil is the zero value for pointers (like Java's null for objects)
//   - Go auto-dereferences struct fields: p.Name == (*p).Name

package main

import "fmt"

func main() {
	// === BASICS ===
	x := 42
	p := &x // p is *int — a pointer to x
	fmt.Println("x:", x)
	fmt.Println("&x (address):", p)  // some memory address like 0xc000...
	fmt.Println("*p (dereference):", *p) // 42

	// Modify through pointer
	*p = 100
	fmt.Println("x after *p = 100:", x) // 100 — x was changed!

	// === WHY POINTERS MATTER ===
	// In Java: void increment(int x) { x++; } — doesn't change caller's x
	// In Go: same! Pass by value means the function gets a COPY.

	a := 10
	incrementByValue(a)
	fmt.Println("a after incrementByValue:", a) // still 10

	incrementByPointer(&a)
	fmt.Println("a after incrementByPointer:", a) // 11

	// === POINTER TO STRUCT ===
	// This is where Go and Java feel similar — Java objects are already pointers
	type Point struct{ X, Y int }

	p1 := Point{1, 2}        // value — on stack (probably)
	p2 := &Point{3, 4}       // pointer to Point — on heap

	// Auto-dereference: Go lets you write p2.X instead of (*p2).X
	fmt.Println("p1.X:", p1.X)
	fmt.Println("p2.X:", p2.X) // same as (*p2).X

	// Mutating through pointer
	p2.X = 99
	fmt.Println("p2 after mutation:", *p2)

	// === new() ===
	// new(T) allocates zeroed T and returns *T
	// Java: new Integer(0) — rarely useful there; in Go it's occasionally handy
	n := new(int)   // *int pointing to a zero int
	*n = 7
	fmt.Println("new int:", *n)

	// === POINTER VS VALUE — when to use which? ===
	// Rule of thumb (covered more in Day 7 with methods):
	//   Use *T when:
	//     1. The function needs to modify the value
	//     2. The struct is large (avoid copying)
	//     3. You need to represent "optional" (nil = absent)
	//   Use T when:
	//     1. The value is small (int, bool, small struct)
	//     2. The function just reads the value
	//     3. You want immutability guarantees

	// === NIL POINTER ===
	var ptr *int // nil pointer — zero value
	fmt.Println("nil pointer:", ptr)
	// fmt.Println(*ptr) // would PANIC — like Java's NullPointerException

	// Always check nil before dereferencing
	if ptr != nil {
		fmt.Println("value:", *ptr)
	} else {
		fmt.Println("pointer is nil")
	}

	// === POINTER IN PRACTICE — returning modified data ===
	// Java: all objects are already pointers so you just return the object
	// Go: return a pointer when the receiver needs to be modified or is large
	rect := newRectangle(5, 3)
	fmt.Printf("Rectangle: %+v\n", *rect) // %+v prints field names too
	rect.scale(2)
	fmt.Printf("Scaled: %+v\n", *rect)

	// === COMPARING POINTERS ===
	a2 := 42
	b2 := 42
	pa := &a2
	pb := &b2
	pc := pa // pc points to same variable as pa

	fmt.Println("pa == pb:", pa == pb) // false — different addresses
	fmt.Println("pa == pc:", pa == pc) // true — same address
	fmt.Println("*pa == *pb:", *pa == *pb) // true — same value

	// === THE KEY MENTAL MODEL ===
	// Java:  MyObj obj = new MyObj() — obj is always a reference (pointer)
	// Go:    obj := MyObj{}          — obj is a VALUE (copy-on-assign)
	//        obj := &MyObj{}         — obj is a pointer (reference-like)
	// This is why Go methods on pointers matter (Day 7)
}

type Rectangle struct {
	Width, Height float64
}

func newRectangle(w, h float64) *Rectangle {
	return &Rectangle{Width: w, Height: h}
}

func (r *Rectangle) scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

func incrementByValue(x int) {
	x++ // modifies local copy only
}

func incrementByPointer(x *int) {
	*x++ // modifies the actual variable
}

// === EXERCISES ===
// 1. Write swap(a, b *int) that swaps two ints in-place using pointers.
//    Call it and verify both variables changed.
//
// 2. Write a function sumSlice(nums []int, result *int) that stores
//    the sum into result. (Note: normally you'd just return the sum —
//    this is for pointer practice.)
//
// 3. Create a Config struct with a pointer field: Parent *Config.
//    Build a chain of 3 configs and traverse the chain.
//
// 4. When does this panic? Write the code and add a nil check to fix it:
//    var s *string
//    fmt.Println(*s)  // what happens?
//
// 5. Java question: In Java, if you pass a String to a method and modify it,
//    the caller's reference doesn't change (strings are immutable).
//    In Go, if you pass a *string and reassign the pointer inside the function,
//    does the caller's pointer change? Write code to find out.
