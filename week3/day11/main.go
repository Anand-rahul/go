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
	"math/rand"
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

	fmt.Println("Ex1")
	// 1. Write a parallel downloader simulation:
	//    URLs := []string{"url1", "url2", "url3", "url4", "url5"}
	//    For each URL, launch a goroutine that "downloads" (sleeps randomly 10-100ms)
	//    and prints the URL when done. Use WaitGroup. Print total elapsed time.

	urls := []string{"url1", "url2", "url3", "url4", "url5"}
	var wg1 sync.WaitGroup
	startNew := time.Now()

	for _, url := range urls {
		wg1.Add(1)

		duration := time.Duration(rand.Intn(91)+10) * time.Millisecond

		go downloadSimulation(url, duration, &wg1)
	}

	wg1.Wait()

	fmt.Printf("Total elapsed time: %v\n", time.Since(startNew))

	fmt.Println("Ex2")
	// 2. Fix the loop variable capture bug:
	//    for i := 0; i < 5; i++ {
	//        go func() { fmt.Println(i) }()
	//    }
	//    Why does this often print "5 5 5 5 5"? Fix it two ways:
	//    a) capture: i := i  b) pass as argument: go func(i int) { }(i)
	//
	var wg2 sync.WaitGroup
	for i := 0; i < 5; i++ {
		i := i // create a new variable

		wg2.Add(1)
		go func() {
			defer wg2.Done()
			fmt.Println(i)
		}()
	}

	for i := 0; i < 5; i++ {
		wg2.Add(1)

		go func(i int) {
			defer wg2.Done()
			fmt.Println(i)
		}(i)
	}

	wg2.Wait()

	fmt.Println("Ex3")
	//❯ go run -race week3/day11/main.go
	// CPUs: 12
	// GOMAXPROCS: 12

	// Hello from goroutine-1 (goroutine)
	// Hello from goroutine-2 (goroutine)
	// Hello from goroutine-3 (goroutine)

	// worker 4 starting
	// worker 2 starting
	// worker 3 starting
	// worker 1 starting
	// worker 3 done
	// worker 1 done
	// worker 4 done
	// worker 2 done
	// all workers done in 50.236333ms (sequential would be 140ms)

	// ==================
	// WARNING: DATA RACE
	// Read at 0x00c00028a048 by goroutine 17:
	//   main.unsafeCounter.func1()
	//       /home/rahul/codedump/go/week3/day11/main.go:46 +0x7b

	// Previous write at 0x00c00028a048 by goroutine 18:
	//   main.unsafeCounter.func1()
	//       /home/rahul/codedump/go/week3/day11/main.go:46 +0x8d

	// Goroutine 17 (running) created at:
	//   main.unsafeCounter()
	//       /home/rahul/codedump/go/week3/day11/main.go:44 +0x78
	//   main.main()
	//       /home/rahul/codedump/go/week3/day11/main.go:163 +0x52a

	// Goroutine 18 (finished) created at:
	//   main.unsafeCounter()
	//       /home/rahul/codedump/go/week3/day11/main.go:44 +0x78
	//   main.main()
	//       /home/rahul/codedump/go/week3/day11/main.go:163 +0x52a
	// ==================
	// ==================
	// WARNING: DATA RACE
	// Write at 0x00c00028a048 by goroutine 19:
	//   main.unsafeCounter.func1()
	//       /home/rahul/codedump/go/week3/day11/main.go:46 +0x8d

	// Previous write at 0x00c00028a048 by goroutine 21:
	//   main.unsafeCounter.func1()
	//       /home/rahul/codedump/go/week3/day11/main.go:46 +0x8d

	// Goroutine 19 (running) created at:
	//   main.unsafeCounter()
	//       /home/rahul/codedump/go/week3/day11/main.go:44 +0x78
	//   main.main()
	//       /home/rahul/codedump/go/week3/day11/main.go:163 +0x52a

	// Goroutine 21 (finished) created at:
	//   main.unsafeCounter()
	//       /home/rahul/codedump/go/week3/day11/main.go:44 +0x78
	//   main.main()
	//       /home/rahul/codedump/go/week3/day11/main.go:163 +0x52a
	// ==================
	// unsafe counter (should be 1000, may not be): 866
	// safe counter: 1000

	// background goroutine started
	// background goroutine finished

	// launched 10000 goroutines, results[9999]=19998
	// goroutines at end: 1
	// Ex1
	// downloading from url1 starting
	// downloading from url4 starting
	// downloading from url2 starting
	// downloading from url5 starting
	// downloading from url3 starting
	// downloading from url2 finished
	// downloading from url3 finished
	// downloading from url1 finished
	// downloading from url4 finished
	// downloading from url5 finished
	// Total elapsed time: 71.234754ms
	// Ex2
	// 0
	// 1
	// 2
	// 3
	// 1
	// 0
	// 4
	// 2
	// 3
	// 4
	// Found 2 data race(s)
	// exit status 66

	fmt.Println("Ex4")
	// 4. Write a function that launches N goroutines, each incrementing a shared
	//    counter. Use sync.Mutex to make it safe. Verify result is always N.

	n := 1000
	var wg3 sync.WaitGroup
	var mutex sync.Mutex
	result := newSafeCounter(n, &wg3, &mutex)

	fmt.Printf("Expected: %d\n", n)
	fmt.Printf("Actual:   %d\n", result)

	fmt.Println("Ex5")
	// 5. In Java, threads have priorities and names. In Go, goroutines are anonymous
	//    and equal. How would you track "which goroutine did what"?
	//    Hint: pass an ID as a parameter and collect results in a slice or channel.

	var wg4 sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg4.Add(2)
		go loggingWorker(i, &wg4)
		go loggingWorker(i*5, &wg4)
	}

	wg4.Wait()
}

func loggingWorker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("worker %d processing\n", id)
}

func newSafeCounter(n int, wg *sync.WaitGroup, mu *sync.Mutex) int {
	var counter int

	for i := 0; i < n; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()

	return counter
}

func downloadSimulation(url string, duration time.Duration, wg *sync.WaitGroup) {
	defer wg.Done() // signal completion when function returns
	fmt.Printf("downloading from %s starting\n", url)
	time.Sleep(duration)
	fmt.Printf("downloading from %s finished\n", url)
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
