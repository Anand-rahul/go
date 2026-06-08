// Day 15: Concurrency Patterns — Worker Pool, Fan-Out/Fan-In, Pipeline
// HOW TO RUN: go run week3/day15/main.go
//
// These are the building blocks you'll use in real Go services.
// Each pattern solves a specific concurrency problem.

package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================================
// PATTERN 1: WORKER POOL
// N goroutines process M jobs from a shared channel
// Java: ExecutorService with fixed thread pool
// ============================================================

type Job struct{ ID int }
type Result struct {
	JobID  int
	Output string
}

func workerPool(ctx context.Context, numWorkers int, jobs <-chan Job) <-chan Result {
	results := make(chan Result, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		workerID := i
		go func() {
			defer wg.Done()
			for {
				select {
				case job, ok := <-jobs:
					if !ok {
						return
					}
					// Simulate work
					time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
					results <- Result{
						JobID:  job.ID,
						Output: fmt.Sprintf("worker%d processed job%d", workerID, job.ID),
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

// ============================================================
// PATTERN 2: FAN-OUT / FAN-IN
// Fan-out: one source → N goroutines doing work in parallel
// Fan-in:  N channels → one output channel (merge)
// Java: parallel streams or CompletableFuture.allOf()
// ============================================================

func fanOut(in <-chan int, n int) []<-chan int {
	outs := make([]<-chan int, n)
	for i := 0; i < n; i++ {
		out := make(chan int)
		outs[i] = out
		go func(out chan<- int) {
			for v := range in {
				time.Sleep(10 * time.Millisecond) // simulate processing
				out <- v * v                       // each worker squares its value
			}
			close(out)
		}(out)
	}
	return outs
}

func fanIn(done <-chan struct{}, channels ...<-chan int) <-chan int {
	merged := make(chan int)
	var wg sync.WaitGroup

	merge := func(ch <-chan int) {
		defer wg.Done()
		for v := range ch {
			select {
			case merged <- v:
			case <-done:
				return
			}
		}
	}

	wg.Add(len(channels))
	for _, ch := range channels {
		go merge(ch)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

// ============================================================
// PATTERN 3: PIPELINE
// Stage 1 → Stage 2 → Stage 3 (each stage is a goroutine)
// ============================================================

// Stage 1: generate numbers
func generate(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Stage 2: filter even numbers
func filterEven(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n%2 == 0 {
				select {
				case out <- n:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// Stage 3: square the values
func squareAll(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// ============================================================
// PATTERN 4: SEMAPHORE — limit concurrency
// Java: Semaphore(N) from java.util.concurrent
// ============================================================

type Semaphore chan struct{}

func NewSemaphore(n int) Semaphore {
	return make(Semaphore, n)
}

func (s Semaphore) Acquire() { s <- struct{}{} }
func (s Semaphore) Release() { <-s }

func limitedParallel(sem Semaphore, tasks []func() string) []string {
	results := make([]string, len(tasks))
	var wg sync.WaitGroup
	for i, task := range tasks {
		wg.Add(1)
		i, task := i, task
		go func() {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()
			results[i] = task()
		}()
	}
	wg.Wait()
	return results
}

// ============================================================
// PATTERN 5: OR-DONE — drain channel respecting cancellation
// ============================================================

func orDone(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-done:
					return
				}
			}
		}
	}()
	return out
}

func main() {
	// === WORKER POOL ===
	fmt.Println("=== Worker Pool ===")
	ctx := context.Background()
	jobs := make(chan Job, 10)
	for i := 1; i <= 8; i++ {
		jobs <- Job{ID: i}
	}
	close(jobs)

	results := workerPool(ctx, 3, jobs)
	for r := range results {
		fmt.Println(" ", r.Output)
	}

	// === FAN-OUT / FAN-IN ===
	fmt.Println("\n=== Fan-Out / Fan-In ===")
	in := make(chan int, 5)
	for _, n := range []int{1, 2, 3, 4, 5} {
		in <- n
	}
	close(in)

	done := make(chan struct{})
	outs := fanOut(in, 3) // 3 workers reading from same input
	merged := fanIn(done, outs...)
	for v := range merged {
		fmt.Print(v, " ")
	}
	fmt.Println()
	close(done)

	// === PIPELINE ===
	fmt.Println("\n=== Pipeline ===")
	ctx2, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := generate(ctx2, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	evens := filterEven(ctx2, nums)
	squares := squareAll(ctx2, evens)

	for v := range squares {
		fmt.Print(v, " ") // 4 16 36 64 100
	}
	fmt.Println()

	// === SEMAPHORE ===
	fmt.Println("\n=== Semaphore (max 2 concurrent) ===")
	sem := NewSemaphore(2)
	tasks := make([]func() string, 6)
	for i := range tasks {
		i := i
		tasks[i] = func() string {
			time.Sleep(20 * time.Millisecond)
			return fmt.Sprintf("task %d done", i)
		}
	}
	start := time.Now()
	res := limitedParallel(sem, tasks)
	fmt.Printf("6 tasks with sem=2 in %v\n", time.Since(start))
	for _, r := range res {
		fmt.Println(" ", r)
	}
}

// === EXERCISES ===
// 1. Extend the worker pool to track which jobs failed.
//    Add Result.Err error field. Make some jobs randomly fail.
//    Collect and print failed jobs at the end.
//
// 2. Implement a "bounded pipeline":
//    Generate 1000 numbers → filter primes → square → collect first 10 results.
//    Use context cancellation to stop all stages after 10 results are collected.
//
// 3. Write a parallel file processor (simulated):
//    Given []string of filenames, process each in parallel with max 5 concurrent,
//    collect results in order (not arrival order — use index-based result slice).
//
// 4. Implement "retry with exponential backoff":
//    func withRetry(ctx context.Context, fn func() error, maxAttempts int) error
//    Double the wait time after each failure: 100ms, 200ms, 400ms, 800ms...
//    Stop retrying if context is cancelled.
//
// 5. Compare the fan-out pattern here (multiple goroutines reading from one channel)
//    vs. the pipeline pattern (one goroutine per stage).
//    When would you use each? What are the trade-offs?
