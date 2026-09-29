// Day 33: Multithreading in Go — goroutines, channels, sync.WaitGroup
// HOW TO RUN: go run week7/day33/main.go
//
// Java dev key shifts:
//   - goroutine = lightweight thread managed by the Go runtime (not the OS)
//   - channel = typed pipe for goroutines to communicate safely (avoids shared-memory locking)
//   - sync.WaitGroup = like Java's CountDownLatch, waits for a group of goroutines to finish
//   - sync.Mutex = like Java's synchronized block, for protecting shared state

package main

import (
	"fmt"
	"sync"
)

// worker simulates doing some work and sends its result on the results channel.
func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		results <- j * j // square the number as "work"
		fmt.Printf("worker %d processed job %d\n", id, j)
	}
}

func main() {
	const numWorkers = 4
	const numJobs = 10

	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	var wg sync.WaitGroup

	// Start a pool of worker goroutines
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// Send jobs
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	// Close results once all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results as they arrive
	sum := 0
	for r := range results {
		sum += r
	}
	fmt.Printf("sum of squares: %d\n", sum)
}

// ============================================================
// EXERCISES
// ============================================================
//
// Exercise 1: Mutex-protected counter
//   Launch 100 goroutines that each increment a shared counter 1000 times.
//   Protect the counter with a sync.Mutex and print the final value (should be 100000).
//
// Exercise 2: Fan-in pattern
//   Create three goroutines that each generate numbers on their own channel.
//   Merge ("fan-in") all three channels into a single output channel and print
//   everything received on it.
//
// Exercise 3: Context cancellation
//   Start a goroutine that loops doing work until its context is cancelled.
//   Use context.WithTimeout to cancel it after 2 seconds and observe the goroutine
//   exit cleanly.
