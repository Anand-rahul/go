// Day 16: Packages and Modules
// HOW TO RUN: go run week4/day16/main.go
//
// Java dev key shifts:
//   - go.mod = pom.xml/build.gradle (but much simpler)
//   - No classpath hell — modules are content-addressed by version
//   - Package = directory (one package per directory)
//   - Exported name = starts with capital letter (no public keyword)
//   - import "module/path/to/pkg" — uses the full module path
//   - init() runs automatically before main() — one per file, multiple allowed
//   - 'go get' adds dependencies, 'go mod tidy' cleans up unused ones
//   - Blank import: import _ "pkg" — runs init() for side effects only
//   - Alias import: import foo "some/long/pkg/name"

package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
)

// === INIT FUNCTION ===
// Java: static { } initializer block
// Go: init() — runs before main(), after all package-level vars are set
// Multiple init() functions are allowed in one file (or across files in a package)
// Order: package-level vars → init() → main()

var globalConfig = map[string]string{} // initialized before init()

func init() {
	globalConfig["env"] = "development"
	globalConfig["version"] = "1.0.0"
	fmt.Println("init() ran — config loaded")
}

// Another init in the same file — both run
func init() {
	globalConfig["feature_flags"] = "enabled"
	fmt.Println("second init() ran")
}

func main() {
	fmt.Println("main() starting")
	fmt.Println("config:", globalConfig)

	// === STDLIB PACKAGE DEMOS ===
	// Go's standard library is rich — less need for third-party deps

	// math package
	fmt.Printf("\n=== math ===\n")
	fmt.Printf("Pi: %.5f\n", math.Pi)
	fmt.Printf("Sqrt(2): %.5f\n", math.Sqrt(2))
	fmt.Printf("Ceil(1.2): %.0f\n", math.Ceil(1.2))
	fmt.Printf("Floor(1.9): %.0f\n", math.Floor(1.9))
	fmt.Printf("Pow(2,10): %.0f\n", math.Pow(2, 10))
	fmt.Printf("Log2(1024): %.0f\n", math.Log2(1024))

	// strings package — Java: String methods + StringUtils
	fmt.Printf("\n=== strings ===\n")
	s := "  Hello, Go World!  "
	fmt.Println(strings.TrimSpace(s))
	fmt.Println(strings.ToUpper(s))
	fmt.Println(strings.ToLower(s))
	fmt.Println(strings.Contains(s, "Go"))
	fmt.Println(strings.HasPrefix(strings.TrimSpace(s), "Hello"))
	fmt.Println(strings.HasSuffix(strings.TrimSpace(s), "World!"))
	fmt.Println(strings.Replace(s, "Go", "Golang", 1))
	fmt.Println(strings.Count(s, "l"))
	parts := strings.Split("a,b,c,d", ",")
	fmt.Println("split:", parts)
	fmt.Println("join:", strings.Join(parts, " | "))
	fmt.Println("index:", strings.Index(s, "Go"))

	// strings.Builder — like Java's StringBuilder (efficient string building)
	var sb strings.Builder
	for i := 0; i < 5; i++ {
		fmt.Fprintf(&sb, "item%d ", i)
	}
	fmt.Println("built:", sb.String())

	// sort package — Java: Collections.sort() or Arrays.sort()
	fmt.Printf("\n=== sort ===\n")
	nums := []int{5, 2, 8, 1, 9, 3, 7, 4, 6}
	sort.Ints(nums)
	fmt.Println("sorted ints:", nums)

	words := []string{"banana", "apple", "cherry", "date"}
	sort.Strings(words)
	fmt.Println("sorted strings:", words)

	// Custom sort — sort.Slice with comparator (Java: Comparator.comparing)
	type Person struct{ Name string; Age int }
	people := []Person{
		{"Charlie", 30}, {"Alice", 25}, {"Bob", 35},
	}
	sort.Slice(people, func(i, j int) bool {
		return people[i].Age < people[j].Age // sort by age ascending
	})
	fmt.Println("sorted by age:", people)

	// Sort stable (preserves original order of equal elements)
	sort.SliceStable(people, func(i, j int) bool {
		return people[i].Name < people[j].Name
	})
	fmt.Println("sorted by name:", people)

	// Binary search
	idx, found := sort.Find(len(nums), func(i int) int {
		return nums[i] - 7 // return <0, 0, or >0
	})
	fmt.Printf("binary search for 7: idx=%d found=%v\n", idx, found)

	// math/rand
	fmt.Printf("\n=== rand ===\n")
	r := rand.New(rand.NewSource(42)) // seeded source
	fmt.Println("rand ints:", r.Intn(100), r.Intn(100), r.Intn(100))

	data := []int{1, 2, 3, 4, 5}
	r.Shuffle(len(data), func(i, j int) { data[i], data[j] = data[j], data[i] })
	fmt.Println("shuffled:", data)
}

// === MODULE COMMANDS REFERENCE ===
// go mod init mymodule          — create go.mod
// go get github.com/foo/bar     — add dependency
// go get github.com/foo/bar@v2  — specific version
// go mod tidy                   — remove unused, add missing deps
// go mod download               — download all deps
// go mod vendor                 — copy deps into vendor/ folder
// go list -m all                — list all dependencies
// go build ./...                — build all packages
// go test ./...                 — test all packages
// go vet ./...                  — run static analysis

// === EXERCISES ===
// 1. Create a new file: week4/day16/strutil.go with package main
//    Add a function: func TitleCase(s string) string
//    Use strings.Fields and strings.Title (or manual capitalization).
//    Call it from main().
//    This shows how multiple files share a package.
//
// 2. Write a function sortByMultipleFields(people []Person) that sorts
//    first by City, then by Age within the same city.
//    Use sort.SliceStable twice (stable preserves previous sort order).
//
// 3. Use strings.Builder to build an HTML table from a [][]string.
//    <table><tr><td>cell</td></tr></table>
//
// 4. Run: go mod tidy && go list -m all
//    Note what's in go.mod and go.sum. What does go.sum protect against?
//
// 5. Create a utils subpackage: week4/utils/strings.go with package utils
//    Export a function Reverse(s string) string.
//    Import it in day16/main.go as: import "golearn/week4/utils"
//    Call utils.Reverse("hello").
