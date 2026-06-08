// Day 9: Embedding — Composition Over Inheritance
// HOW TO RUN: go run week2/day09/main.go
//
// Java dev key shifts:
//   - Go has NO inheritance (no extends, no super)
//   - Instead: embedding — include one type inside another
//   - Embedded type's methods get "promoted" to the outer type
//   - This IS NOT inheritance — it's delegation/composition
//   - You can override promoted methods by defining your own
//   - Interface embedding: compose larger interfaces from smaller ones
//   - The Go proverb: "Prefer composition over inheritance" — Go enforces this

package main

import "fmt"

// === STRUCT EMBEDDING ===
// Java: class Animal { ... }; class Dog extends Animal { ... }
// Go: embed Animal inside Dog

type Animal struct {
	Name string
}

func (a Animal) Breathe() {
	fmt.Printf("%s breathes\n", a.Name)
}

func (a Animal) Eat(food string) {
	fmt.Printf("%s eats %s\n", a.Name, food)
}

// Dog embeds Animal — NOT inherits
type Dog struct {
	Animal // no field name = anonymous embedding
	Breed  string
}

func (d Dog) Bark() {
	fmt.Printf("%s barks!\n", d.Name) // d.Name promoted from Animal
}

// Override promoted method
func (d Dog) Eat(food string) {
	fmt.Printf("%s scarfs down %s!\n", d.Name, food) // shadows Animal.Eat
}

// Cat also embeds Animal
type Cat struct {
	Animal
	Indoor bool
}

func (c Cat) Purr() {
	fmt.Printf("%s purrs\n", c.Name)
}

// === EMBEDDING MULTIPLE TYPES ===
type Logger struct{ Prefix string }
type Timer struct{ Duration int }

func (l Logger) Log(msg string) {
	fmt.Printf("[%s] %s\n", l.Prefix, msg)
}

func (t Timer) Elapsed() string {
	return fmt.Sprintf("%dms", t.Duration)
}

type Worker struct {
	Logger
	Timer
	JobName string
}

// === INTERFACE EMBEDDING ===
// Build larger interfaces from smaller ones
type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

// ReadWriter embeds both — any type implementing Read + Write satisfies ReadWriter
type ReadWriter interface {
	Reader
	Writer
}

// Closer interface
type Closer interface {
	Close() error
}

// ReadWriteCloser embeds three interfaces
type ReadWriteCloser interface {
	ReadWriter
	Closer
}

// === EMBEDDING POINTERS ===
type Base struct {
	ID int
}

func (b *Base) SetID(id int) {
	b.ID = id
}

func (b Base) GetID() int {
	return b.ID
}

type Derived struct {
	*Base // pointer embedding — shared Base
	Value string
}

// === PRACTICAL EXAMPLE ===
// A common Go pattern: embed a mutex for safe concurrent access
// (preview of Day 13's sync.Mutex)
type SafeMap struct {
	// In real code: sync.RWMutex (we'll cover this Day 13)
	data map[string]int
}

func NewSafeMap() *SafeMap {
	return &SafeMap{data: make(map[string]int)}
}

func (m *SafeMap) Set(key string, val int) { m.data[key] = val }
func (m *SafeMap) Get(key string) (int, bool) {
	v, ok := m.data[key]
	return v, ok
}

// HTTP middleware pattern via embedding
type ResponseWriter struct {
	status int
	body   string
}

func (rw *ResponseWriter) WriteHeader(status int) { rw.status = status }
func (rw *ResponseWriter) Write(body string)      { rw.body = body }

type LoggingResponseWriter struct {
	*ResponseWriter              // embed to get all methods
	Logger                       // embed logger
}

func (lrw *LoggingResponseWriter) WriteHeader(status int) {
	lrw.Log(fmt.Sprintf("Status: %d", status))
	lrw.ResponseWriter.WriteHeader(status) // call embedded method explicitly
}

func main() {
	// === BASIC EMBEDDING ===
	d := Dog{
		Animal: Animal{Name: "Rex"},
		Breed:  "Labrador",
	}

	// Promoted methods — accessed as if they were Dog's own methods
	d.Breathe() // Animal.Breathe promoted
	d.Bark()

	// Overridden method
	d.Eat("kibble")         // calls Dog.Eat (overrides Animal.Eat)
	d.Animal.Eat("kibble")  // explicitly call Animal.Eat

	c := Cat{Animal: Animal{Name: "Whiskers"}, Indoor: true}
	c.Breathe()
	c.Purr()

	fmt.Println()

	// === MULTIPLE EMBEDDING ===
	w := Worker{
		Logger:  Logger{Prefix: "JOB"},
		Timer:   Timer{Duration: 150},
		JobName: "data-sync",
	}
	w.Log("started " + w.JobName)
	fmt.Println("elapsed:", w.Elapsed())

	fmt.Println()

	// === POINTER EMBEDDING ===
	base := &Base{ID: 1}
	derived := Derived{Base: base, Value: "hello"}
	fmt.Println("derived ID:", derived.GetID())

	derived.SetID(99)
	fmt.Println("base ID after SetID on derived:", base.ID) // 99 — shared!

	fmt.Println()

	// === EMBEDDING VS INHERITANCE KEY DIFFERENCES ===
	// In Java, Dog IS-A Animal. In Go, Dog HAS-A Animal.
	// You cannot pass a Dog where Animal is expected:
	// var a Animal = d  // COMPILE ERROR — Dog is not Animal
	// But you CAN pass d.Animal:
	var a Animal = d.Animal
	a.Eat("food")

	// Embedding doesn't satisfy the parent interface either
	// (unless you add explicit forwarding methods)

	// === LOGGING RESPONSE WRITER ===
	rw := &ResponseWriter{}
	lrw := &LoggingResponseWriter{
		ResponseWriter: rw,
		Logger:         Logger{Prefix: "HTTP"},
	}
	lrw.WriteHeader(200) // calls overriding method (logs + delegates)
	lrw.Write("hello")   // calls promoted ResponseWriter.Write
	fmt.Printf("status=%d body=%q\n", rw.status, rw.body)
}

// === EXERCISES ===
// 1. Create: type Vehicle{Make, Model string, Year int} with method Info() string.
//    Create Car embeds Vehicle + Doors int.
//    Create Truck embeds Vehicle + Payload float64.
//    Print info for both.
//
// 2. Create a named-type StringSet using embedding:
//    type StringSet struct { m map[string]struct{} }
//    Add methods: Add, Remove, Contains, Size, ToSlice.
//    Embed StringSet inside TaggedStringSet which adds Tags []string.
//
// 3. In Java, you can use Animal variable to hold a Dog.
//    In Go, define an interface Breather { Breathe() } and show that
//    both Dog and Cat satisfy it (without any explicit declaration).
//
// 4. Can you embed an interface inside a struct?
//    type Base struct { Stringer } — what does this do?
//    Try it and explain when this pattern is useful.
//
// 5. What happens with method conflicts?
//    Embed two types that both have a method String() string.
//    Try calling it and observe the compiler error.
