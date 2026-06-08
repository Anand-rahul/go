// Day 3: Arrays and Slices
// HOW TO RUN: go run week1/day03/main.go
//
// Java dev key shifts:
//   - Arrays are fixed-size value types — rarely used directly in Go
//   - Slices are the workhorse — think ArrayList backed by an array
//   - Slices have three parts: pointer, length, capacity
//   - append() may allocate a NEW backing array — always capture the return value
//   - make([]T, len, cap) pre-allocates like new ArrayList(initialCapacity)
//   - Passing a slice to a function shares the underlying array (like Java objects)

package main

import "fmt"

func main() {
	// === ARRAYS — fixed size, VALUE type ===
	// Java: int[] arr = new int[5];
	var arr [5]int // zero-valued: [0 0 0 0 0]
	arr[0] = 10
	arr[2] = 50
	fmt.Println("array:", arr, "len:", len(arr))

	// Array literal — [...]int{} lets compiler count
	primes := [5]int{2, 3, 5, 7, 11}
	days := [...]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	fmt.Println(primes, days)

	// Arrays are VALUE types in Go — assignment makes a copy
	// Java: int[] b = a  →  b is a reference to the same array
	// Go:   b := a       →  b is a full COPY
	a := [3]int{1, 2, 3}
	b := a
	b[0] = 99
	fmt.Println("a:", a) // [1 2 3] — unchanged
	fmt.Println("b:", b) // [99 2 3]

	// === SLICES — dynamic, the real workhorse ===
	// Slice type: []int  (no size in brackets — that's what makes it a slice not array)

	// 1. Slice literal
	langs := []string{"Go", "Java", "Python", "JS"}
	fmt.Printf("langs: %v  len=%d  cap=%d\n", langs, len(langs), cap(langs))

	// 2. Slice from array — a VIEW into the array (no copy!)
	//    syntax: array[low:high]  →  elements [low, high)
	nums := [6]int{10, 20, 30, 40, 50, 60}
	s := nums[1:4] // [20 30 40]
	fmt.Println("slice:", s)

	// Mutating the slice also mutates the underlying array!
	s[0] = 99
	fmt.Println("nums after s[0]=99:", nums) // nums[1] is now 99

	// 3. make — pre-allocate
	//    make([]T, length, capacity)  — like new ArrayList(cap)
	words := make([]string, 0, 10)
	fmt.Printf("words: len=%d cap=%d\n", len(words), cap(words))

	// 4. nil slice — zero value, safe to use
	var nilSlice []int
	fmt.Println("nil?", nilSlice == nil, "len:", len(nilSlice)) // true, 0

	// === APPEND ===
	// Like ArrayList.add() — but returns a new slice header (MUST capture return)
	fruits := []string{"apple", "banana"}
	fruits = append(fruits, "cherry")             // single element
	fruits = append(fruits, "date", "elderberry") // variadic
	fmt.Println("fruits:", fruits)

	// Append slice to slice with ... (spread/unpack operator)
	more := []string{"fig", "grape"}
	fruits = append(fruits, more...)
	fmt.Println("all fruits:", fruits)

	// Watch capacity grow as append doubles the backing array
	growing := make([]int, 0)
	for i := 0; i < 9; i++ {
		growing = append(growing, i)
		fmt.Printf("  len=%-2d  cap=%-2d\n", len(growing), cap(growing))
	}

	// === COPY ===
	// Java: System.arraycopy() or Arrays.copyOf()
	src := []int{1, 2, 3, 4, 5}
	dst := make([]int, len(src))
	n := copy(dst, src)
	dst[0] = 999
	fmt.Printf("copied %d elements. src: %v  dst: %v\n", n, src, dst)

	// === SLICE TRICKS ===
	data := []int{1, 2, 3, 4, 5, 6, 7, 8}
	fmt.Println("first 3:", data[:3])              // [1 2 3]
	fmt.Println("last 3:", data[len(data)-3:])     // [6 7 8]
	fmt.Println("middle:", data[2:5])              // [3 4 5]

	// Delete element at index i — no built-in remove
	i := 2
	data = append(data[:i], data[i+1:]...)
	fmt.Println("after delete index 2:", data)

	// === 2D SLICES ===
	// Java: int[][] matrix = new int[3][3];
	matrix := make([][]int, 3)
	for i := range matrix {
		matrix[i] = make([]int, 3)
		for j := range matrix[i] {
			matrix[i][j] = i*3 + j
		}
	}
	for _, row := range matrix {
		fmt.Println(row)
	}

	// === SLICE IS A REFERENCE — critical mental model ===
	// Passing a slice to a function shares the underlying array
	original := []int{1, 2, 3}
	mutateFirst(original)
	fmt.Println("after mutateFirst:", original) // [99 2 3]

	fmt.Println("Ex1")
	nNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for _, v := range nNums {
		if v%2 == 0 {
			fmt.Println(v)
		}
	}

	fmt.Println("Ex2")
	intVal := []int{1, 2, 2, 3, 1, 4, 5, 4}
	fmt.Println(removeDuplicates(intVal))

	fmt.Println("Ex3")

	testInt:= make([]int, 0, 3)
	for i := 0; i < 7; i++ {
		fmt.Printf("len=%d cap=%d slice=%v\n",
			len(testInt),
			cap(testInt),
			testInt)

		testInt = append(testInt, i)
	}

	fmt.Println("Ex4")
	ums := []int{1, 2, 3, 4, 5}

	reversed := reverseSlice(ums[1:3])

	fmt.Println("Original:", ums)
	fmt.Println("Reversed:", reversed)

	fmt.Println("Ex5")
	//fixed Capcity
	s1 := []int{1, 2, 3}
	s2 := s1

	s2 = append(s2, 4)
	fmt.Println("New:", s1)
	fmt.Println("Old:", s2)

	s2[0] = 99
	fmt.Println("New:", s1)
	fmt.Println("Old:", s2)

	//Increased Capacity
	s3 := make([]int, 3, 6)
	s3[0] = 1
	s3[1] = 2
	s3[2] = 3
	s4 := s3
	s4 = append(s4, 4)
	fmt.Println("New:", s3)
	fmt.Println("Old:", s4)
	s4[0] = 99
	fmt.Println("New:", s3)
	fmt.Println("Old:", s4)

}

func mutateFirst(s []int) {
	s[0] = 99
}

func removeDuplicates(in []int) []int {
	seen := make(map[int]bool)
	result := []int{}

	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}
func reverseSlice(s []int) []int {
	result := make([]int, len(s))

	for i := 0; i < len(s); i++ {
		result[i] = s[len(s)-1-i]
	}

	return result
}

// === EXERCISES ===
// 1. Create a slice of ints 1-10. Use range to print only even values.
//
// 2. Write: func removeDuplicates(in []int) []int
//    Input: [1,2,2,3,3,3,4] → Output: [1,2,3,4]
//    Hint: use a map[int]bool to track seen values.
//
// 3. Use make([]int, 0, 3). Append 7 elements one by one.
//    Print len and cap after each append. At what element does cap change?
//
// 4. Write: func reverseSlice(s []int) []int — returns reversed copy.
//
// 5. s1 := []int{1,2,3}; s2 := s1; s2 = append(s2, 4)
//    Is s1 affected? Now try: s2[0] = 99 — is s1 affected?
//    Why the difference? (This is the most important slice gotcha.)
