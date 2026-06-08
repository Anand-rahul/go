// Day 14: context.Context — Cancellation, Deadlines, Request-Scoped Values
// HOW TO RUN: go run week3/day14/main.go
//
// Java dev key shifts:
//   - context.Context is Go's answer to thread interruption + request-scoped data
//   - Java: Thread.interrupt() + ThreadLocal. Go: context.Context passed explicitly
//   - ALWAYS pass Context as the FIRST parameter: func Do(ctx context.Context, ...)
//   - context.Background() — root context (never cancelled), use in main/tests
//   - context.TODO()       — placeholder while refactoring (same as Background)
//   - WithCancel  — manual cancel button
//   - WithTimeout — auto-cancel after duration
//   - WithDeadline — auto-cancel at absolute time
//   - WithValue   — attach request-scoped data (use sparingly, prefer explicit params)
//   - When ctx.Done() is closed, check ctx.Err() for reason (Canceled or DeadlineExceeded)

package main

import (
	"context"
	"fmt"
	"time"
)

// === WITHCANCEL — manual cancellation ===
func doWork(ctx context.Context, name string) {
	for {
		select {
		case <-ctx.Done():
			// ctx.Err() tells us WHY it was cancelled
			fmt.Printf("%s: cancelled — %v\n", name, ctx.Err())
			return
		default:
			fmt.Printf("%s: working...\n", name)
			time.Sleep(20 * time.Millisecond)
		}
	}
}

// === WITHTIMEOUT — auto-cancel after duration ===
// Java: Future.get(timeout, TimeUnit.SECONDS)
func slowQuery(ctx context.Context) (string, error) {
	resultCh := make(chan string, 1)

	go func() {
		// Simulate slow database query
		time.Sleep(100 * time.Millisecond)
		resultCh <- "query result"
	}()

	select {
	case result := <-resultCh:
		return result, nil
	case <-ctx.Done():
		return "", ctx.Err() // context.DeadlineExceeded or context.Canceled
	}
}

// === WITHVALUE — request-scoped values ===
// Use for: request IDs, auth tokens, tracing spans
// Do NOT use for: passing optional parameters to functions (use real params)
type contextKey string // unexported type avoids collisions between packages

const (
	requestIDKey contextKey = "requestID"
	userIDKey    contextKey = "userID"
)

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

// Real-world example: HTTP handler pattern
func handleRequest(ctx context.Context) {
	reqID := getRequestID(ctx)
	fmt.Printf("[%s] handling request\n", reqID)

	// Pass context down to all sub-operations
	if err := fetchUser(ctx, 42); err != nil {
		fmt.Printf("[%s] error: %v\n", reqID, err)
		return
	}
	fmt.Printf("[%s] request complete\n", reqID)
}

func fetchUser(ctx context.Context, id int) error {
	// Check if already cancelled before doing expensive work
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Simulate DB call with its own timeout
	ctx2, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel() // always defer cancel to release resources

	result, err := slowQuery(ctx2)
	if err != nil {
		return fmt.Errorf("fetchUser(%d): %w", id, err)
	}
	fmt.Printf("  fetched user %d: %q\n", id, result)
	return nil
}

// === PROPAGATING CANCELLATION ===
// When parent is cancelled, all children are cancelled automatically
func demonstratePropagation() {
	parent, parentCancel := context.WithCancel(context.Background())

	child1, child1Cancel := context.WithCancel(parent)
	child2, _ := context.WithTimeout(parent, 10*time.Second)

	defer child1Cancel()

	fmt.Println("child1 done (before):", child1.Err())
	fmt.Println("child2 done (before):", child2.Err())

	parentCancel() // cancelling parent cancels ALL children

	// Give goroutines a moment to propagate
	time.Sleep(1 * time.Millisecond)
	fmt.Println("child1 done (after parent cancel):", child1.Err())
	fmt.Println("child2 done (after parent cancel):", child2.Err())
}

// === CONTEXT IN A WORKER POOL ===
func contextualWorkerPool(ctx context.Context, jobs <-chan int) <-chan string {
	results := make(chan string)
	go func() {
		defer close(results)
		for {
			select {
			case job, ok := <-jobs:
				if !ok {
					return
				}
				// Check context before doing expensive work
				select {
				case <-ctx.Done():
					return
				default:
				}
				result := fmt.Sprintf("processed job %d", job)
				select {
				case results <- result:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				fmt.Println("worker pool cancelled:", ctx.Err())
				return
			}
		}
	}()
	return results
}

func main() {
	// === WithCancel ===
	fmt.Println("=== WithCancel ===")
	ctx, cancel := context.WithCancel(context.Background())
	go doWork(ctx, "worker-A")
	time.Sleep(60 * time.Millisecond)
	cancel() // stop the worker
	time.Sleep(10 * time.Millisecond)

	// === WithTimeout ===
	fmt.Println("\n=== WithTimeout ===")

	// Fast query — succeeds
	ctx2, cancel2 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel2()
	result, err := slowQuery(ctx2)
	fmt.Printf("fast query: result=%q err=%v\n", result, err)

	// Slow query with tight timeout — times out
	ctx3, cancel3 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel3()
	_, err = slowQuery(ctx3)
	fmt.Printf("slow query: err=%v\n", err)

	// === WithDeadline ===
	fmt.Println("\n=== WithDeadline ===")
	deadline := time.Now().Add(30 * time.Millisecond)
	ctx4, cancel4 := context.WithDeadline(context.Background(), deadline)
	defer cancel4()
	fmt.Printf("deadline set: %v (in %v)\n", deadline.Format("15:04:05.000"), time.Until(deadline))
	<-ctx4.Done()
	fmt.Println("deadline reached:", ctx4.Err())

	// === WithValue ===
	fmt.Println("\n=== WithValue ===")
	ctx5 := withRequestID(context.Background(), "req-abc-123")
	ctx5 = context.WithValue(ctx5, userIDKey, 42)
	handleRequest(ctx5)

	// === Propagation ===
	fmt.Println("\n=== Cancellation Propagation ===")
	demonstratePropagation()

	// === Worker pool with context ===
	fmt.Println("\n=== Worker Pool ===")
	ctx6, cancel6 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel6()

	jobs := make(chan int, 10)
	for i := 1; i <= 5; i++ {
		jobs <- i
	}
	close(jobs)

	results := contextualWorkerPool(ctx6, jobs)
	for r := range results {
		fmt.Println(" ", r)
	}
}

// === EXERCISES ===
// 1. Write an HTTP-style middleware chain where each step:
//    - Adds a value to context (request ID, start time)
//    - Checks if context is cancelled before proceeding
//    - Calls the next handler
//
// 2. Implement retry with context:
//    func retryWithContext(ctx context.Context, fn func() error, maxRetries int) error
//    Retry up to maxRetries times but stop early if context is cancelled.
//
// 3. Write a grep-like function:
//    func grepFiles(ctx context.Context, pattern string, files []string) []string
//    Search each file in a goroutine. Stop all workers if ctx is cancelled.
//    Hint: each goroutine should select on its work AND ctx.Done()
//
// 4. What happens if you call cancel() multiple times? Is that safe?
//    What happens if you forget to call cancel() — is there a resource leak?
//    (Hint: run with leak detector or just reason about it)
//
// 5. Why is it considered bad practice to store context in a struct?
//    When is it acceptable? (Hint: HTTP request objects do this)
