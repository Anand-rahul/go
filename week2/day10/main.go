// Day 10: Error Handling — Errors as Values
// HOW TO RUN: go run week2/day10/main.go
//
// Java dev key shifts:
//   - No try/catch/finally/throws — errors are just values returned alongside results
//   - error is a built-in interface: type error interface { Error() string }
//   - By convention, error is ALWAYS the last return value
//   - Ignoring errors with _ is possible but dangerous (like catching and swallowing)
//   - errors.Is() — check error chain (like instanceof)
//   - errors.As() — extract typed error (like instanceof + cast)
//   - fmt.Errorf("context: %w", err) — wraps error with context (%w = wrap)
//   - panic/recover exists but is NOT for normal error handling (Day 22)

package main

import (
	"errors"
	"fmt"
	"strconv"
)

// === BASIC ERROR ===
// error interface: just { Error() string }
// errors.New — creates a simple error value
var ErrDivisionByZero = errors.New("division by zero") // sentinel error

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

// === CUSTOM ERROR TYPE ===
// Java: class ValidationException extends RuntimeException { int field; }
// Go: struct that implements the error interface
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %q: %s", e.Field, e.Message)
}

func validateAge(age int) error {
	if age < 0 {
		return &ValidationError{Field: "age", Message: "cannot be negative"}
	}
	if age > 150 {
		return &ValidationError{Field: "age", Message: "unrealistically large"}
	}
	return nil
}

// === WRAPPED ERRORS — adding context ===
// Java: throw new RuntimeException("context", cause);
// Go:   return fmt.Errorf("context: %w", err)
// %w wraps the error so errors.Is/As can unwrap the chain

type DBError struct {
	Query string
	Err   error
}

func (e *DBError) Error() string {
	return fmt.Sprintf("db error running %q: %v", e.Query, e.Err)
}

func (e *DBError) Unwrap() error { // enables errors.Is/As to dig into the chain
	return e.Err
}

var ErrNotFound = errors.New("not found")
var ErrTimeout = errors.New("timeout")

func queryDB(query string) error {
	// Simulate: wrap a sentinel error in a DBError
	return &DBError{Query: query, Err: ErrNotFound}
}

func getUser(id int) (string, error) {
	err := queryDB(fmt.Sprintf("SELECT * FROM users WHERE id=%d", id))
	if err != nil {
		return "", fmt.Errorf("getUser(%d): %w", id, err) // adds context, wraps
	}
	return "alice", nil
}

// === errors.Is and errors.As ===
// errors.Is: checks if any error in the chain matches (like Java's instanceof on cause chain)
// errors.As: extracts the first error in chain of a given type

// === MULTIPLE ERROR STRATEGIES ===

// 1. Sentinel errors — compare with ==  or errors.Is
// 2. Typed errors — use errors.As to extract and inspect
// 3. Opaque errors — caller doesn't know/care about type, just displays it

// === ERROR ACCUMULATION ===
type MultiError []error

func (m MultiError) Error() string {
	if len(m) == 1 {
		return m[0].Error()
	}
	msg := fmt.Sprintf("%d errors:", len(m))
	for _, e := range m {
		msg += "\n  - " + e.Error()
	}
	return msg
}

type UserInput struct {
	Name  string
	Email string
	Age   int
}

func validateUser(u UserInput) error {
	var errs MultiError
	if u.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if u.Email == "" {
		errs = append(errs, errors.New("email is required"))
	}
	if err := validateAge(u.Age); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// === THE PATTERN: handle errors immediately ===
// Java style — throw and catch far away
// Go style — check right where the error occurs

func processInput(input string) (int, error) {
	n, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("processInput: invalid number %q: %w", input, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("processInput: %w", ErrDivisionByZero) // reuse sentinel
	}
	return n * 2, nil
}

func main() {
	// Basic error check pattern
	if result, err := divide(10, 3); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Printf("10/3 = %.4f\n", result)
	}

	if _, err := divide(10, 0); err != nil {
		fmt.Println("divide by zero error:", err)
	}

	// errors.Is — check sentinel error through chain
	_, err := divide(5, 0)
	fmt.Println("is ErrDivisionByZero:", errors.Is(err, ErrDivisionByZero))

	// Custom error type
	if err := validateAge(-5); err != nil {
		fmt.Println("validation:", err)
	}

	// errors.As — extract typed error
	if err := validateAge(200); err != nil {
		var ve *ValidationError
		if errors.As(err, &ve) {
			fmt.Printf("field: %q message: %q\n", ve.Field, ve.Message)
		}
	}

	// Wrapped errors — errors.Is and errors.As dig through the chain
	_, err2 := getUser(42)
	fmt.Println("\ngetUser error:", err2)
	fmt.Println("is ErrNotFound:", errors.Is(err2, ErrNotFound)) // true — unwraps chain

	var dbErr *DBError
	if errors.As(err2, &dbErr) {
		fmt.Println("db query was:", dbErr.Query)
	}

	// Multi-error accumulation
	bad := UserInput{Name: "", Email: "", Age: -1}
	if err := validateUser(bad); err != nil {
		fmt.Println("\nvalidation errors:", err)
	}

	good := UserInput{Name: "Rahul", Email: "r@example.com", Age: 28}
	if err := validateUser(good); err != nil {
		fmt.Println("unexpected error:", err)
	} else {
		fmt.Println("valid user!")
	}

	// Error in a chain
	result, err3 := processInput("abc")
	if err3 != nil {
		fmt.Println("\nprocessInput error:", err3)
		// Unwrap to see underlying error
		fmt.Println("underlying:", errors.Unwrap(errors.Unwrap(err3)))
	}
	_ = result
}

// === EXERCISES ===
// 1. Write a function parseConfig(path string) (*Config, error) that:
//    - Returns a sentinel ErrConfigNotFound if path is empty
//    - Returns a wrapped error if any field is invalid
//    Use errors.Is and errors.As to handle both cases in the caller.
//
// 2. Create a type NetworkError with Code int, Message string.
//    Make it implement error and Unwrap. Write a chain:
//    low-level error → NetworkError → fmt.Errorf wrap.
//    Verify errors.Is and errors.As work through all levels.
//
// 3. Go doesn't have finally, but defer achieves the same thing.
//    Write a function openAndProcess(filename string) error that:
//    - Opens a file (simulate with a flag variable)
//    - Uses defer to "close" it (print "closing file")
//    - Returns early with an error midway through
//    Verify the defer still runs.
//
// 4. The common mistake: if err != nil { return err } — this loses context.
//    Rewrite it using fmt.Errorf with %w to add function name context.
//
// 5. Why should you return nil directly instead of returning a typed nil?
//    Demonstrate the bug with:
//    func getError() error { var e *ValidationError = nil; return e }
//    fmt.Println(getError() == nil)  // what does this print and why?
