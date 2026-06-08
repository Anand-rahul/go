// Day 11: Goroutines — Concurrency Basics
// HOW TO RUN: go run week3/day11/main.go
//
// Java dev key shifts:
//   - Goroutines are NOT threads — they're multiplexed onto OS threads by the Go runtime
//   - Starting one: just add 'go' before a function call
//   - A goroutine costs ~2KB (a Java thread costs ~1MB default stack)
//   - Running 100,000 goroutines is normal; running 100,000 threads is not
//   - Go motto: "Don't communicate by sharing memory; share memory by communicating"
//   - sync.WaitGroup is like Java's CountDownLatch — wait for goroutines to finish
//   - Data races are real — use channels or sync primitives (Day 13)

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// === BASIC GOROUTINE ===
func sayHello(name string) {
	fmt.Printf("Hello from %s (goroutine)\n", name)
}

// Simulates work with a sleep
func worker(id int, duration time.Duration, wg *sync.WaitGroup) {
	defer wg.Done() // signal completion when function returns
	fmt.Printf("worker %d starting\n", id)
	time.Sleep(duration)
	fmt.Printf("worker %d done\n", id)
}

// === DATA RACE — the danger of shared state without synchronization ===
// This is intentionally wrong — Day 13 shows how to fix it
func unsafeCounter() {
	counter := 0
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++ // RACE CONDITION — multiple goroutines write concurrently
		}()
	}
	wg.Wait()
	fmt.Println("unsafe counter (should be 1000, may not be):", counter)
}

// === SAFE WITH MUTEX (preview of Day 13) ===
func safeCounter() {
	counter := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println("safe counter:", counter) // always 1000
}

// === GOROUTINE LIFECYCLE ===
// Goroutines live until their function returns (or the program exits)
// The main goroutine is special — when it exits, ALL goroutines are killed

func demonstrateLifecycle() {
	done := make(chan struct{})

	go func() {
		fmt.Println("background goroutine started")
		time.Sleep(50 * time.Millisecond)
		fmt.Println("background goroutine finished")
		close(done)
	}()

	<-done // wait for signal (channels covered Day 12)
}

// === GOROUTINE LEAK — a common bug ===
// A goroutine that blocks forever on a channel is a leak
// Always ensure goroutines can exit

func leakyVersion() chan int {
	ch := make(chan int)
	go func() {
		// This goroutine blocks forever if nobody reads from ch!
		ch <- 42
	}()
	return ch
}

// Fixed version: use a done channel or buffer
func nonLeakyVersion() chan int {
	ch := make(chan int, 1) // buffered — goroutine won't block
	go func() {
		ch <- 42
	}()
	return ch
}

// === GOROUTINES ARE CHEAP — spawn thousands ===
func demonstrateScale() {
	const N = 10_000
	var wg sync.WaitGroup
	results := make([]int, N)

	for i := 0; i < N; i++ {
		wg.Add(1)
		i := i // capture loop variable — important! (same as Java lambda gotcha)
		go func() {
			defer wg.Done()
			results[i] = i * 2 // safe because each goroutine writes unique index
		}()
	}
	wg.Wait()
	fmt.Printf("launched %d goroutines, results[9999]=%d\n", N, results[N-1])
}

func main() {
	// Show available CPUs and how Go uses them
	fmt.Println("CPUs:", runtime.NumCPU())
	fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0)) // 0 = query current value

	fmt.Println()

	// Basic goroutine — fire and forget
	go sayHello("goroutine-1")
	go sayHello("goroutine-2")
	go sayHello("goroutine-3")
	// WARNING: main might exit before goroutines run!
	time.Sleep(10 * time.Millisecond) // poor man's wait — use WaitGroup instead

	fmt.Println()

	// WaitGroup — proper way to wait for goroutines
	// Java equivalent: CountDownLatch or CompletableFuture
	var wg sync.WaitGroup
	durations := []time.Duration{30, 50, 20, 40} // milliseconds
	start := time.Now()

	for i, d := range durations {
		wg.Add(1) // increment before launching goroutine
		go worker(i+1, d*time.Millisecond, &wg)
	}

	wg.Wait() // block until all workers call wg.Done()
	fmt.Printf("all workers done in %v (sequential would be %v)\n",
		time.Since(start), 140*time.Millisecond)

	fmt.Println()

	// Data race demonstration
	unsafeCounter()
	safeCounter()

	fmt.Println()

	// Goroutine lifecycle
	demonstrateLifecycle()

	fmt.Println()

	// Scale
	demonstrateScale()

	// Active goroutines
	fmt.Println("goroutines at end:", runtime.NumGoroutine())
}

// === EXERCISES ===
// 1. Write a parallel downloader simulation:
//    URLs := []string{"url1", "url2", "url3", "url4", "url5"}
//    For each URL, launch a goroutine that "downloads" (sleeps randomly 10-100ms)
//    and prints the URL when done. Use WaitGroup. Print total elapsed time.
//
// 2. Fix the loop variable capture bug:
//    for i := 0; i < 5; i++ {
//        go func() { fmt.Println(i) }()
//    }
//    Why does this often print "5 5 5 5 5"? Fix it two ways:
//    a) capture: i := i  b) pass as argument: go func(i int) { }(i)
//
// 3. Run: go run -race week3/day11/main.go
//    The race detector will catch unsafeCounter's data race. Observe the output.
//
// 4. Write a function that launches N goroutines, each incrementing a shared
//    counter. Use sync.Mutex to make it safe. Verify result is always N.
//
// 5. In Java, threads have priorities and names. In Go, goroutines are anonymous
//    and equal. How would you track "which goroutine did what"?
//    Hint: pass an ID as a parameter and collect results in a slice or channel.
