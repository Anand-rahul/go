// Day 1: Hello Go — Variables, Constants, Basic Types
// HOW TO RUN: go run week1/day01/main.go
//
// Java dev key shifts:
//   - No semicolons, braces MUST be on the same line (gofmt enforces this)
//   - Type comes AFTER the variable name: var x int, NOT int x
//   - := is short declaration (type inferred) — use inside functions only
//   - Every declared variable must be used — compile error if unused
//   - Zero values: int=0, bool=false, string="", pointer=nil (no NPE surprises)
//   - No wrapper types: just int (not Integer), bool (not Boolean)

package main

import "fmt"

// === CONSTANTS ===
// Java: static final int MAX = 100;
const MaxRetries = 3

const (
	StatusOK    = 200
	StatusError = 500
)

// iota — auto-incrementing enum-like constants (no separate enum keyword)
type Direction int

const (
	North Direction = iota // 0
	East                   // 1
	South                  // 2
	West                   // 3
)

const (
	MON =iota +1
	TUE
	WED
	THU
	FRI
	SAT
	SUN
)

func (d Direction) String() string {
	return [...]string{"North", "East", "South", "West"}[d]
}

func main() {
	// === VAR DECLARATIONS ===
	var count int   // zero value = 0
	var name string // zero value = ""
	var active bool // zero value = false
	fmt.Println("zero values:", count, "|"+name+"|", active)

	// Short declaration — most common inside functions
	// Java: int score = 42; String city = "Bangalore";
	score := 42
	city := "Bangalore"
	pi := 3.14159
	fmt.Println(score, city, pi)

	// Multiple assignment on one line
	x, y := 10, 20
	fmt.Println("x:", x, "y:", y)

	// Swap — no temp variable needed
	x, y = y, x
	fmt.Println("after swap — x:", x, "y:", y)

	// === BASIC TYPES ===
	// Java: int, long, float, double, char, byte, boolean
	// Go:   int, int8/16/32/64, uint, uint8/16/32/64, float32/64, bool, string
	//       byte = uint8     (8-bit, useful for raw data)
	//       rune = int32     (Unicode code point, Go's char equivalent)
	var age int32 = 25
	var salary float64 = 75_000.50 // underscore as digit separator (Go 1.13+)
	var initial rune = 'R'         // single quotes = rune literal
	var letter byte = 'A'
	fmt.Println(age, salary, initial, letter)
	fmt.Printf("initial as char: %c\n", initial)

	// === STRINGS ===
	// Java: String is an object. Go: string is a value type (immutable byte slice)
	greeting := "Hello, Go!"
	fmt.Println(greeting)
	fmt.Println("byte length:", len(greeting)) // bytes, not characters!

	// Raw string literal with backticks (like Java text blocks)
	json := `{
"name": "rahul",
"lang": "go"
}`
	fmt.Println(json)

	// Printf format verbs — similar to Java's String.format()
	first, last := "Rahul", "Anand"
	fmt.Printf("Name: %s %s\n", first, last)
	fmt.Printf("Score: %d, Pi: %.3f\n", score, pi)
	fmt.Printf("Type of score: %T\n", score) // %T prints the Go type
	fmt.Printf("Value+type: %v %T\n", city, city)

	// === TYPE CONVERSION ===
	// Java does implicit widening: int → long. Go does NOT. Always explicit.
	a := 42
	b := float64(a) // must cast explicitly
	c := int(b)
	fmt.Println(a, b, c)

	// === IOTA DEMO ===
	fmt.Println("Directions:", North, East, South, West)

	//ex 1
	pName := "Rahul"
	cAge := 24
	cYoe := 4
	fmt.Printf("name :%s , age :%d ,yoe :%d\n",pName,cAge, cYoe)

	//ex2
	fmt.Println("MON =", MON)
	fmt.Println("TUE =", TUE)
	fmt.Println("WED =", WED)
	fmt.Println("THU =", THU)
	fmt.Println("FRI =", FRI)
	fmt.Println("SAT =", SAT)
	fmt.Println("SUN =", SUN)

	//ex3
	unusedVar := 5
	_=unusedVar

	// var i int = int(float32(3.14))
	// fmt.Println(i)
	f := float32(3.14)
	i := int(f)

	fmt.Println(i)
}

// === EXERCISES ===
// 1. Declare variables for: your name, age, years of Java experience.
//    Print them with fmt.Printf using %s, %d, %f.
//
// 2. Create a const block for weekdays MON=1..SUN=7 using iota+1.
//    Print them.
//
// 3. Try this: declare a variable but don't use it. See the compile error.
//    Fix it using the blank identifier: _ = unusedVar
//
// 4. What is the zero value of each type? Verify:
//    var i int, var f float64, var s string, var b bool — print and confirm.
//
// 5. Java lets you do: int i = (int) 3.14  — silent truncation.
//    In Go, try: var i int = 3.14 — what happens? Why is Go stricter?
