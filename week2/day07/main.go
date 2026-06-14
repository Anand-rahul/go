// Day 7: Methods and Receivers
// HOW TO RUN: go run week2/day07/main.go
//
// Java dev key shifts:
//   - Methods are defined outside the struct (not inside like Java classes)
//   - Receiver is the first "parameter": func (r Rect) Area() float64
//   - Value receiver (r Rect) — method gets a copy, like Java's final parameter
//   - Pointer receiver (r *Rect) — method can modify the struct (most methods use this)
//   - Rule: if ANY method needs a pointer receiver, use pointer receiver for ALL
//   - You can define methods on ANY type in your package, not just structs
//   - No method overloading — use different names

package main

import (
	"fmt"
	"errors"
	"math"
)

// === VALUE RECEIVER ===
// Use when: reading only, struct is small, value semantics make sense

type Circle struct {
	Radius float64
}

// Value receiver — Circle is copied into this method
// Java: public double area() { return Math.PI * radius * radius; }
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// Value receiver CAN'T modify the original
func (c Circle) SetRadius(r float64) {
	c.Radius = r // modifies the copy — original unchanged!
}

// === POINTER RECEIVER ===
// Use when: modifying the struct, struct is large, consistency

type Rectangle struct {
	Width, Height float64
}

func (r *Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor  // modifies original
	r.Height *= factor
}

func (r *Rectangle) String() string { // fmt.Stringer interface (Day 8)
	return fmt.Sprintf("Rectangle(%.1f × %.1f)", r.Width, r.Height)
}

// === METHODS ON NON-STRUCT TYPES ===
// Java: you can't add methods to int or String — you'd create a wrapper
// Go: define methods on any named type in your package

type Celsius float64
type Fahrenheit float64

func (c Celsius) ToFahrenheit() Fahrenheit {
	return Fahrenheit(c*9/5 + 32)
}

func (f Fahrenheit) ToCelsius() Celsius {
	return Celsius((f - 32) * 5 / 9)
}

func (c Celsius) String() string {
	return fmt.Sprintf("%.1f°C", float64(c))
}

// === METHOD SET RULES ===
// T  value:   can call value receivers only
// *T pointer: can call BOTH value and pointer receivers
// (Go auto-deref: if you have *T you can call T's methods)

type Stack struct {
	items []int
}

func (s *Stack) Push(v int) {
	s.items = append(s.items, v)
}

func (s *Stack) Pop() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	last := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return last, true
}

func (s *Stack) Peek() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	return s.items[len(s.items)-1], true
}

func (s Stack) Len() int { // value receiver — just reading
	return len(s.items)
}

func main() {
	// === VALUE RECEIVER ===
	c := Circle{Radius: 5}
	fmt.Printf("Circle area: %.2f\n", c.Area())
	fmt.Printf("Circle perimeter: %.2f\n", c.Perimeter())

	// Value receiver doesn't modify original
	c.SetRadius(10)
	fmt.Println("Radius after SetRadius(10):", c.Radius) // still 5!

	// === POINTER RECEIVER ===
	r := &Rectangle{Width: 4, Height: 3}
	fmt.Printf("Rect area: %.1f\n", r.Area())
	fmt.Println("Before scale:", r)
	r.Scale(2)
	fmt.Println("After scale:", r) // uses String() method

	// Go auto-takes address when needed
	r2 := Rectangle{Width: 6, Height: 2}
	r2.Scale(3) // Go does (&r2).Scale(3) automatically
	fmt.Println("r2 after scale:", r2)

	// === METHODS ON NAMED TYPES ===
	boiling := Celsius(100)
	fmt.Printf("%s = %.1f°F\n", boiling, boiling.ToFahrenheit())

	body := Fahrenheit(98.6)
	fmt.Printf("%.1f°F = %s\n", body, body.ToCelsius())

	// === STACK ===
	s := &Stack{}
	s.Push(1)
	s.Push(2)
	s.Push(3)
	fmt.Println("stack len:", s.Len())

	for s.Len() > 0 {
		v, _ := s.Pop()
		fmt.Print(v, " ")
	}
	fmt.Println()

	v, ok := s.Pop()
	fmt.Printf("pop from empty: v=%d ok=%v\n", v, ok)

	// === CHAINING (builder pattern) ===
	type Builder struct{ parts []string }

	add := func(b *Builder, s string) *Builder {
		b.parts = append(b.parts, s)
		return b
	}

	b := &Builder{}
	add(add(add(b, "Go"), "is"), "fun")
	fmt.Println("built:", b.parts)


	fmt.Println("Ex1")
	// 1. Create a BankAccount struct with balance float64.
	//    Add methods: Deposit(amount float64), Withdraw(amount float64) error,
	//    Balance() float64. Withdraw should return error if insufficient funds.
	acc := NewBankAccount(1000)

	acc.Deposit(500)

	if err := acc.Withdraw(2000); err != nil {
		fmt.Println("Withdraw failed:", err)
	}

	if err := acc.Withdraw(300); err != nil {
		fmt.Println("Withdraw failed:", err)
	}

	fmt.Println("Current balance:", acc.Balance())

	fmt.Println("Ex2")
	// 2. Create type StringSlice []string with methods:
	//    Contains(s string) bool, Add(s string), Remove(s string)

	ss := StringSlice{"go", "java", "rust"}

	fmt.Println(ss.Contains("java"))
	fmt.Println(ss.Contains("python"))

	ss.Add("python")
	fmt.Println(ss)

	ss.Remove("java")
	fmt.Println(ss)

	fmt.Println("Ex3")

	var i Incrementer

	counter := Counter{}

	i = &counter

	i.Increment()
	i.Increment()

	fmt.Println(counter.count)

	fmt.Println("Ex4")
	// 4. Create a type Kilometers float64 and Miles float64 with conversion methods.
	km := Kilometers(10)
	mi := km.ToMiles()

	fmt.Printf("%.2f km = %.2f miles\n", km, mi)

	m := Miles(10)
	k := m.ToKilometers()

	fmt.Printf("%.2f miles = %.2f km\n", m, k)

	fmt.Println("Ex5")
	// 5. What's the difference between:
	//    c := Circle{5}; c.SetRadius(10)  — and —
	//    c := &Circle{5}; c.SetRadius(10)
	//    Write both and verify which one actually changes the radius.
	//
	circleNew := Circle{5}
	circleNew.SetRadius(10)

	fmt.Println(circleNew.Radius)
	circleNew.SetRadiusPointer(10)
	fmt.Println(circleNew.Radius)


	cV := &Circle{50}
	cV.SetRadius(10)

	fmt.Println(cV.Radius)
	cV.SetRadiusPointer(10)
	fmt.Println(cV.Radius)


}

func (c *Circle) SetRadiusPointer(r float64) {
	c.Radius = r
}

type Kilometers float64
type Miles float64

func (k Kilometers) ToMiles() Miles {
	return Miles(k * 0.621371)
}

func (m Miles) ToKilometers() Kilometers {
	return Kilometers(m * 1.60934)
}

type Incrementer interface {
	Increment()
}

type Counter struct {
	count int
}

func (c *Counter) Increment() {
	c.count++
}

type StringSlice []string

func (ss StringSlice) Contains(s string) bool {
	for _, str := range ss {
		if str == s {
			return true
		}
	}
	return false
}

func (ss *StringSlice) Add(s string) {
	*ss = append(*ss, s)
}

func (ss *StringSlice) Remove(s string) {
	result := make([]string, 0, len(*ss))
	for _, str := range *ss {
		if str != s {
			result = append(result, str)
		}
	}
	*ss = result
}


type BankAccount struct {
	balance float64
}

func NewBankAccount(balance float64) *BankAccount {
	return &BankAccount{
		balance: balance,
	}
}

func (b *BankAccount) Deposit(amount float64) {
	b.balance += amount
}

func (b *BankAccount) Withdraw(amount float64) error {
	if amount > b.balance {
		return errors.New("insufficient funds")
	}

	b.balance -= amount
	return nil
}

func (b *BankAccount) Balance() float64 {
	return b.balance
}


// === EXERCISES ===
// 1. Create a BankAccount struct with balance float64.
//    Add methods: Deposit(amount float64), Withdraw(amount float64) error,
//    Balance() float64. Withdraw should return error if insufficient funds.
//
// 2. Create type StringSlice []string with methods:
//    Contains(s string) bool, Add(s string), Remove(s string)
//
// 3. Why should you use pointer receivers consistently?
//    Create struct Counter with a value-receiver Increment() and try to
//    use it through an interface — you'll see a compile error.
//    Fix it with a pointer receiver. (Preview of Day 8)
//
// 4. Create a type Kilometers float64 and Miles float64 with conversion methods.
//
// 5. What's the difference between:
//    c := Circle{5}; c.SetRadius(10)  — and —
//    c := &Circle{5}; c.SetRadius(10)
//    Write both and verify which one actually changes the radius.
