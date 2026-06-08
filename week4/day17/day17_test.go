// Day 17 TEST FILE — testing package demonstration
// Run: go test -v ./week4/day17/...
// Run: go test -bench=. -benchmem ./week4/day17/...

package main

import (
	"strings"
	"testing"
)

// === BASIC TEST ===
// Java: @Test public void testDivide() { assertEquals(5.0, divide(10, 2)); }
// Go:
func TestDivideBasic(t *testing.T) {
	result, err := divide(10, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err) // t.Fatalf stops test immediately
	}
	if result != 5.0 {
		t.Errorf("expected 5.0 got %v", result) // t.Errorf marks failure but continues
	}
}

func TestDivideByZero(t *testing.T) {
	_, err := divide(10, 0)
	if err == nil {
		t.Fatal("expected error for division by zero, got nil")
	}
}

// === TABLE-DRIVEN TESTS ===
// Java: @ParameterizedTest with @ValueSource or @MethodSource
// Go: slice of test cases, loop with t.Run
func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr bool
	}{
		{"positive", 10, 2, 5, false},
		{"negative", -10, 2, -5, false},
		{"decimal", 7, 2, 3.5, false},
		{"divide by zero", 5, 0, 0, true},
		{"zero numerator", 0, 5, 0, false},
	}

	for _, tc := range tests {
		tc := tc // capture for subtests (pre-Go 1.22)
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() — run subtests in parallel (uncomment to use)
			result, err := divide(tc.a, tc.b)
			if (err != nil) != tc.wantErr {
				t.Errorf("divide(%v, %v) error = %v, wantErr %v", tc.a, tc.b, err, tc.wantErr)
				return
			}
			if !tc.wantErr && result != tc.want {
				t.Errorf("divide(%v, %v) = %v, want %v", tc.a, tc.b, result, tc.want)
			}
		})
	}
}

// === TABLE-DRIVEN: reverseString ===
func TestReverseString(t *testing.T) {
	cases := []struct{ input, expected string }{
		{"hello", "olleh"},
		{"", ""},
		{"a", "a"},
		{"ab", "ba"},
		{"café", "éfac"}, // Unicode!
		{"racecar", "racecar"},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			got := reverseString(c.input)
			if got != c.expected {
				t.Errorf("reverseString(%q) = %q, want %q", c.input, got, c.expected)
			}
		})
	}
}

// === TEST HELPERS ===
// Helper functions marked with t.Helper() — errors point to the caller, not here
func assertEqual(t *testing.T, got, want interface{}) {
	t.Helper() // makes error messages point to the caller
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsPalindrome(t *testing.T) {
	trueTests := []string{"racecar", "madam", "a", ""}
	for _, s := range trueTests {
		t.Run("true/"+s, func(t *testing.T) {
			assertEqual(t, isPalindrome(s), true)
		})
	}

	falseTests := []string{"hello", "world", "ab"}
	for _, s := range falseTests {
		t.Run("false/"+s, func(t *testing.T) {
			assertEqual(t, isPalindrome(s), false)
		})
	}
}

// === SETUP AND TEARDOWN ===
// Java: @BeforeEach / @AfterEach
// Go: no built-in annotations — use t.Cleanup or explicit setup in each test

func TestWithSetup(t *testing.T) {
	// Setup
	data := []string{"apple", "banana", "cherry"}

	// Teardown with t.Cleanup (runs when test finishes, even on failure)
	t.Cleanup(func() {
		// In real code: close files, delete temp dirs, etc.
		t.Log("cleanup ran")
	})

	if len(data) != 3 {
		t.Errorf("expected 3 items")
	}
}

// === TestMain — package-level setup/teardown ===
// Java: @BeforeAll / @AfterAll
// Uncomment to enable:
// func TestMain(m *testing.M) {
//     fmt.Println("setup before all tests")
//     code := m.Run() // runs all tests
//     fmt.Println("teardown after all tests")
//     os.Exit(code)
// }

// === BENCHMARKS ===
// Go benchmark: func BenchmarkXxx(b *testing.B)
// Run: go test -bench=. -benchmem ./week4/day17/...
// Output: BenchmarkFib20-8   50000   23456 ns/op   0 B/op   0 allocs/op

func BenchmarkFibonacci20(b *testing.B) {
	for i := 0; i < b.N; i++ { // b.N is set by Go to get stable timing
		fibonacci(20)
	}
}

func BenchmarkReverseString(b *testing.B) {
	s := strings.Repeat("hello", 100)
	b.ResetTimer() // don't count setup time
	for i := 0; i < b.N; i++ {
		reverseString(s)
	}
}

// Compare two implementations
func reverseStringBytes(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

func BenchmarkReverseRunes(b *testing.B) {
	s := strings.Repeat("hello", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reverseString(s) // rune-based
	}
}

func BenchmarkReverseBytes(b *testing.B) {
	s := strings.Repeat("hello", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reverseStringBytes(s) // byte-based (faster for ASCII)
	}
}

// === EXERCISES ===
// 1. Write tests for a Stack type (from Day 7):
//    TestPush, TestPop (empty and non-empty), TestPeek.
//    Use table-driven tests for Pop.
//
// 2. Write a benchmark comparing:
//    a) string concatenation with +=
//    b) strings.Builder
//    c) fmt.Sprintf
//    for building a string of 1000 parts.
//    Run with -benchmem to see allocation difference.
//
// 3. Write a test that checks isPalindrome with t.Parallel() subtests.
//    Add -race flag to verify no data races.
//
// 4. Write TestMain to:
//    - Create a temp directory before all tests
//    - Remove it after all tests
//    Use os.MkdirTemp and defer.
//
// 5. What does go test -count=3 do? When is it useful?
//    What does go test -short do? How do you check t.Short() in a test?
