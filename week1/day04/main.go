// Day 4: Maps and Structs
// HOW TO RUN: go run week1/day04/main.go
//
// Java dev key shifts:
//   - map[K]V is Go's HashMap<K,V> — but no methods like .getOrDefault()
//   - Maps must be initialized (make or literal) before writing — else nil panic
//   - Two-value lookup: v, ok := m[key] — no need for .containsKey() + .get()
//   - Structs replace classes — no inheritance, just data + methods
//   - Struct fields are exported (public) if capitalized, unexported if lowercase
//   - No constructors — use plain functions returning struct or *struct

package main

import "fmt"

// === STRUCTS ===
// Java: class Person { String name; int age; ... }
// Go: struct with exported fields (capital = public, lowercase = package-private)
type Person struct {
	Name string // exported (capital N)
	Age  int
	City string
}

// Struct with unexported fields
type BankAccount struct {
	owner   string // unexported — only this package can access
	balance float64
}

// Nested struct
type Address struct {
	Street string
	City   string
	Pin    string
}

type Employee struct {
	Person  // embedded (not inherited) — covered fully in Day 9
	Address Address
	Role    string
	Salary  float64
}

func main() {
	// === MAP BASICS ===
	// Java: Map<String, Integer> m = new HashMap<>();
	// Go:

	// 1. Make (empty map)
	scores := make(map[string]int)
	scores["alice"] = 92
	scores["bob"] = 85
	scores["carol"] = 78
	fmt.Println("scores:", scores)

	// 2. Map literal
	config := map[string]string{
		"host": "localhost",
		"port": "8080",
		"env":  "dev",
	}
	fmt.Println("host:", config["host"])

	// 3. Two-value lookup — replaces Java's containsKey() + get()
	// Java: if (map.containsKey("host")) { String v = map.get("host"); }
	if host, ok := config["host"]; ok {
		fmt.Println("found host:", host)
	}

	// Key that doesn't exist returns zero value (no KeyError, no null)
	missing := config["database"] // returns "" (zero value for string)
	fmt.Printf("missing key: '%s'\n", missing)

	// 4. Delete
	delete(config, "env")
	fmt.Println("after delete:", config)

	// 5. Iterate — Java: for (Map.Entry<K,V> e : map.entrySet())
	for key, value := range scores {
		fmt.Printf("  %s → %d\n", key, value)
	}

	// Map with slice values — like Java: Map<String, List<String>>
	groupedStudents := map[string][]string{
		"math":    {"alice", "bob"},
		"science": {"carol", "dave"},
	}
	groupedStudents["math"] = append(groupedStudents["math"], "eve")
	fmt.Println("math students:", groupedStudents["math"])

	// Set pattern — Go has no Set type, use map[T]struct{}
	// struct{} takes zero bytes (empty struct)
	seen := make(map[string]struct{})
	words := []string{"go", "java", "go", "python", "java", "go"}
	for _, w := range words {
		seen[w] = struct{}{}
	}
	fmt.Println("unique words count:", len(seen))
	_, inSet := seen["go"]
	fmt.Println("go in set:", inSet)

	// === STRUCTS ===

	// 1. Struct literal — positional (fragile, avoid)
	p1 := Person{"Rahul", 28, "Bangalore"}
	fmt.Println("p1:", p1)

	// 2. Named fields (preferred)
	p2 := Person{
		Name: "Priya",
		Age:  25,
		// City: "" — zero value if omitted
	}
	fmt.Println("p2:", p2)

	// 3. var — zero value struct
	var p3 Person
	p3.Name = "Dev"
	p3.Age = 30
	fmt.Println("p3:", p3)

	// 4. Pointer to struct — like Java objects (Java ALWAYS uses heap objects)
	p4 := &Person{Name: "Kiran", Age: 22, City: "Mumbai"}
	fmt.Println("p4:", p4)
	fmt.Println("p4.Name:", p4.Name) // auto-deref — same as (*p4).Name

	// Structs are VALUE types — assignment copies
	// Java: Person p2 = p1 — both point to same object
	// Go:   p5 := p1     — p5 is a COPY
	p5 := p1
	p5.Name = "Copy"
	fmt.Println("p1 after copy mutation:", p1.Name) // Rahul — unchanged
	fmt.Println("p5:", p5.Name)                      // Copy

	// 5. Nested struct
	emp := Employee{
		Person: Person{Name: "Anita", Age: 27, City: "Delhi"},
		Address: Address{
			Street: "10 Main St",
			City:   "Delhi",
			Pin:    "110001",
		},
		Role:   "Engineer",
		Salary: 120000,
	}
	fmt.Println("Employee:", emp.Name, emp.Role) // Name promoted from embedded Person
	fmt.Println("Address:", emp.Address.City)

	// 6. Anonymous struct — one-off use
	point := struct{ X, Y int }{X: 3, Y: 4}
	fmt.Println("point:", point)
}

// No constructors in Go — use regular functions
// Convention: NewTypeName(args) *TypeName
func NewBankAccount(owner string, initial float64) *BankAccount {
	return &BankAccount{owner: owner, balance: initial}
}

func (a *BankAccount) Deposit(amount float64) {
	a.balance += amount
}

func (a *BankAccount) Balance() float64 {
	return a.balance
}

// === EXERCISES ===
// 1. Write a word-frequency counter:
//    Input string: "the cat sat on the mat the cat"
//    Output: map of word → count
//    Print in sorted order (hint: import "sort", sort.Strings(keys))
//
// 2. Create a struct Student{Name string, Grades []int} with a method
//    Average() float64. Create 3 students and print their averages.
//
// 3. Build a phonebook: map[string]string (name → phone).
//    Write functions: Add, Get, Delete, List.
//
// 4. Two maps a and b — what happens when you do c := a and then modify c?
//    Write code to verify. (Maps are reference types — like Java maps!)
//
// 5. Use the set pattern (map[string]struct{}) to find words that appear
//    in both slice1 = ["go","java","rust"] and slice2 = ["python","go","java"]
