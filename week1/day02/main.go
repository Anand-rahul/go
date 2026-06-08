// Day 2: Control Flow — if / for / switch / range
// HOW TO RUN: go run week1/day02/main.go
//
// Java dev key shifts:
//   - No while loop — for does everything
//   - No parentheses around conditions: if x > 0 { not if (x > 0) {
//   - if can have an init statement: if err := do(); err != nil { ... }
//   - switch: NO fallthrough by default (Java falls through — opposite!)
//   - range iterates slices, arrays, maps, strings, channels

package main

import "fmt"

func main() {
	// === IF / ELSE ===
	x := 15
	if x > 10 {
		fmt.Println("big")
	} else if x > 5 {
		fmt.Println("medium")
	} else {
		fmt.Println("small")
	}

	// if with init statement — extremely common in Go for error handling
	// 'remainder' exists only inside the if/else block
	if remainder := x % 2; remainder == 0 {
		fmt.Println(x, "is even")
	} else {
		fmt.Printf("%d is odd (remainder: %d)\n", x, remainder)
	}

	// === FOR — the only loop keyword in Go ===

	// 1. Classic C-style (same as Java)
	for i := 0; i < 5; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// 2. While-style (Java: while (n < 100) { n *= 2; })
	n := 1
	for n < 100 {
		n *= 2
	}
	fmt.Println("first power of 2 >= 100:", n)

	// 3. Infinite loop (Java: while(true) { ... break; })
	count := 0
	for {
		count++
		if count == 3 {
			break
		}
	}
	fmt.Println("broke at:", count)

	// 4. continue — same as Java
	fmt.Print("odd: ")
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			continue
		}
		fmt.Print(i, " ")
	}
	fmt.Println()

	// === RANGE — iterate over collections ===
	// Java: for (int v : list)  or  for (Map.Entry<K,V> e : map.entrySet())
	// Go:   for index, value := range collection

	nums := []int{10, 20, 30, 40, 50}

	for i, v := range nums {
		fmt.Printf("  nums[%d] = %d\n", i, v)
	}

	// Discard index with blank identifier _
	for _, v := range nums {
		fmt.Print(v, " ")
	}
	fmt.Println()

	// range over string iterates RUNES (Unicode code points), not bytes
	// 'é' in UTF-8 is 2 bytes, so the index jumps
	word := "café"
	for i, r := range word {
		fmt.Printf("  index=%d  rune=%c\n", i, r)
	}

	// range over map — order is RANDOM (like Java's HashMap)
	scores := map[string]int{"alice": 90, "bob": 85, "carol": 92}
	for name, score := range scores {
		fmt.Printf("  %s: %d\n", name, score)
	}

	// === SWITCH ===
	// Key: NO fallthrough by default (Java falls through without break!)
	day := "Monday"
	switch day {
	case "Saturday", "Sunday": // comma = OR
		fmt.Println("Weekend")
	case "Monday", "Tuesday", "Wednesday", "Thursday", "Friday":
		fmt.Println("Weekday")
	default:
		fmt.Println("Unknown")
	}

	// Switch with no expression — cleaner if/else chain
	score2 := 75
	switch {
	case score2 >= 90:
		fmt.Println("A")
	case score2 >= 80:
		fmt.Println("B")
	case score2 >= 70:
		fmt.Println("C")
	default:
		fmt.Println("F")
	}

	// Explicit fallthrough — when you actually want Java's default behavior
	switch x {
	case 15:
		fmt.Print("fifteen ")
		fallthrough // explicitly continues to next case
	case 14:
		fmt.Print("(ran case 14 too)")
	}
	fmt.Println()

	// === LABELED BREAK — escape nested loops ===
outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if i == 1 && j == 1 {
				break outer
			}
			fmt.Printf("(%d,%d) ", i, j)
		}
	}
	fmt.Println()
}

// === EXERCISES ===
// 1. Write a loop printing the Fibonacci sequence up to 1000.
//    Use the while-style for loop.
//
// 2. Use range over this map, printing only entries where value > 50:
//    grades := map[string]int{"math": 80, "english": 45, "science": 90}
//
// 3. Write a switch that takes an HTTP status code (200, 201, 400, 404, 500)
//    and prints a human-readable message. Include a default case.
//
// 4. Use the if-init pattern: write a function divide(a, b int) (int, bool)
//    that returns (0, false) when b==0. Call it using:
//    if result, ok := divide(10, 0); ok { ... } else { ... }
//
// 5. var i int = 99; for i := 0; i < 3; i++ { fmt.Println(i) }; fmt.Println(i)
//    What does the last Println print? Why? (scoping question)
