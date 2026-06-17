// Day 12: Channels — Communicating Between Goroutines
// HOW TO RUN: go run week3/day12/main.go
//
// Java dev key shifts:
//   - Channels are typed pipes: send a value, receive a value
//   - make(chan T) — unbuffered, synchronizes sender and receiver
//   - make(chan T, N) — buffered, sender blocks only when full
//   - ch <- v  — send value v to channel ch (blocks until receiver ready)
//   - v := <-ch — receive from ch (blocks until sender ready)
//   - close(ch) — signals no more values; range over channel stops at close
//   - Closest Java equivalent: BlockingQueue / LinkedBlockingQueue
//   - Go proverb: "Don't communicate by sharing memory; share memory by communicating"

package main

import (
	"fmt"
	"time"
	"strings"
	"sync"
)

// === UNBUFFERED CHANNEL ===
// Both sender and receiver must be ready at the same time (synchronization point)
// Like a baton pass in a relay race

func unbufferedDemo() {
	ch := make(chan string)

	go func() {
		fmt.Println("goroutine: about to send")
		ch <- "hello from goroutine" // blocks until main receives
		fmt.Println("goroutine: sent!")
	}()

	time.Sleep(100 * time.Millisecond) // let goroutine start
	fmt.Println("main: about to receive")
	msg := <-ch // blocks until goroutine sends
	fmt.Println("main: received:", msg)
}

// === BUFFERED CHANNEL ===
// Sender blocks only when buffer is full — like a queue
// Java: new LinkedBlockingQueue(capacity)

func bufferedDemo() {
	ch := make(chan int, 3) // buffer size 3

	// Can send 3 values without a receiver being ready
	ch <- 1
	ch <- 2
	ch <- 3
	// ch <- 4 // would block — buffer full

	fmt.Println("buffered len:", len(ch), "cap:", cap(ch))
	fmt.Println(<-ch, <-ch, <-ch)
}

// === CHANNEL DIRECTION (typed constraints) ===
// chan<- T  — send-only channel (for producers)
// <-chan T  — receive-only channel (for consumers)
// Enforces clear ownership in function signatures

func producer(ch chan<- int, count int) { // can only SEND to ch
	for i := 0; i < count; i++ {
		ch <- i
	}
	close(ch) // producer closes when done
}

func consumer(ch <-chan int) { // can only RECEIVE from ch
	for v := range ch { // range on channel: loops until channel closed
		fmt.Print(v, " ")
	}
	fmt.Println()
}

// === PIPELINE PATTERN ===
// Each stage reads from one channel and writes to another

func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

func double(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * 2
		}
		close(out)
	}()
	return out
}

// === DONE CHANNEL PATTERN ===
// Signal cancellation to goroutines
func processWithCancel(done <-chan struct{}) {
	for {
		select {
		case <-done:
			fmt.Println("processWithCancel: cancelled")
			return
		default:
			// simulate work
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// === TIMEOUT PATTERN ===
func fetchWithTimeout(delay time.Duration) (string, error) {
	result := make(chan string, 1)

	go func() {
		time.Sleep(delay) // simulate slow operation
		result <- "data from server"
	}()

	select {
	case data := <-result:
		return data, nil
	case <-time.After(50 * time.Millisecond):
		return "", fmt.Errorf("timeout after 50ms")
	}
}

// === FAN-OUT: one channel, many readers ===
func fanOut(in <-chan int, n int) []<-chan int {
	outs := make([]<-chan int, n)
	for i := 0; i < n; i++ {
		out := make(chan int)
		outs[i] = out
		go func(out chan<- int) {
			for v := range in {
				out <- v
			}
			close(out)
		}(out)
	}
	return outs
}

func main() {
	fmt.Println("=== UNBUFFERED ===")
	unbufferedDemo()

	fmt.Println("\n=== BUFFERED ===")
	bufferedDemo()

	fmt.Println("\n=== PRODUCER / CONSUMER ===")
	ch := make(chan int, 5)
	go producer(ch, 8)
	consumer(ch)

	fmt.Println("\n=== PIPELINE ===")
	// generate numbers → square them → double the squares
	nums := generate(1, 2, 3, 4, 5)
	squares := square(nums)
	doubled := double(squares)
	for v := range doubled {
		fmt.Print(v, " ") // 2 8 18 32 50
	}
	fmt.Println()

	fmt.Println("\n=== TIMEOUT ===")
	if data, err := fetchWithTimeout(20 * time.Millisecond); err == nil {
		fmt.Println("fast fetch:", data)
	}
	if _, err := fetchWithTimeout(100 * time.Millisecond); err != nil {
		fmt.Println("slow fetch:", err)
	}

	fmt.Println("\n=== DONE CHANNEL ===")
	done := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		close(done)
	}()
	processWithCancel(done)

	fmt.Println("\n=== RANGE OVER CLOSED CHANNEL ===")
	// Receiving from a closed channel yields zero values
	// Two-value receive tells you if channel is open
	ch2 := make(chan int, 2)
	ch2 <- 1
	close(ch2)
	v1, ok1 := <-ch2 // 1, true
	v2, ok2 := <-ch2 // 0, false (zero value, channel closed)
	fmt.Printf("v1=%d ok=%v   v2=%d ok=%v\n", v1, ok1, v2, ok2)

	fmt.Println("Ex1")
	sentences := []string{
			"Go makes concurrency easy",
			"Channels help goroutines communicate",
			"Word counting is simple",
			"Practice makes perfect",
		}

	wordChannel := make(chan int)
	var wg sync.WaitGroup

	for _, sentence := range sentences {
		wg.Add(1)
		go countWords(sentence, wordChannel, &wg)
	}

	go func() {
		wg.Wait()
		close(wordChannel)
	}()

	total := 0
	for count := range wordChannel {
		total += count
	}

	fmt.Printf("Total words: %d\n", total)

	fmt.Println("Ex2")
	c1 := make(chan int)
	c2 := make(chan int)
	c3 := make(chan int)

	go func() {
		defer close(c1)
		c1 <- 1
		c1 <- 2
	}()

	go func() {
		defer close(c2)
		c2 <- 3
		c2 <- 4
	}()

	go func() {
		defer close(c3)
		c3 <- 5
		c3 <- 6
	}()

	for v := range merge(c1, c2, c3) {
		fmt.Println(v)
	}

	fmt.Println("Ex3")
	sem := make(chan struct{}, 3) // max 3 concurrent workers

	var newWg sync.WaitGroup

	start := time.Now()

	for i := 1; i <= 10; i++ {
		newWg.Add(1)
		go concurrentLimit(i, sem, &newWg)
	}

	newWg.Wait()

	fmt.Printf("Total time: %v\n", time.Since(start))

	fmt.Println("Ex4")
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
		}
	}()

	test := make(chan int)

	close(test)

	fmt.Println("Sending...")
	//test <- 10 // panic

	fmt.Println("Done")

	// var ch chan int

	// ch <- 10
	//
	fmt.Println("Ex5")
	f := async(func() int {
		time.Sleep(2 * time.Second)
		return 42
	})

	fmt.Println("Doing other work...")

	result := f.Get()

	fmt.Println("Result:", result)
}
type Future struct {
	ch chan int
}

func (f Future) Get() int {
	return <-f.ch
}

func async(fn func() int) Future {
	ch := make(chan int, 1)

	go func() {
		ch <- fn()
		close(ch)
	}()

	return Future{ch: ch}
}

func countWords(sentence string, ch chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	words := strings.Fields(sentence)
	ch <- len(words)
}

func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)

	var wg sync.WaitGroup

	// Forward values from one input channel to out
	output := func(c <-chan int) {
		defer wg.Done()

		for v := range c {
			out <- v
		}
	}

	wg.Add(len(cs))

	for _, c := range cs {
		go output(c)
	}

	// Close out after all input channels are drained
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func concurrentLimit(id int, sem chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// Acquire
	sem <- struct{}{}

	fmt.Printf("Worker %d started\n", id)

	time.Sleep(100 * time.Millisecond)

	fmt.Printf("Worker %d finished\n", id)

	// Release
	<-sem
}

// === EXERCISES ===
// 1. Write a concurrent word counter:
//    Give each goroutine a sentence. Each goroutine counts words and sends
//    result to a channel. Main collects and sums all counts.
//
// 2. Implement a merge function:
//    func merge(cs ...<-chan int) <-chan int
//    that drains all input channels into a single output channel.
//    Use WaitGroup to close output when all inputs are done.
//
// 3. Write a semaphore using a buffered channel to limit concurrency:
//    sem := make(chan struct{}, 3)  // max 3 concurrent
//    Acquire: sem <- struct{}{}
//    Release: <-sem
//    Test with 10 goroutines that each take 100ms.
//
// 4. What happens if you send to a closed channel? Try it and handle the panic.
//    What does receiving from a nil channel do? (Blocks forever)
//
// 5. Java has Future<T> and CompletableFuture<T>.
//    Implement a simple Future using channels:
//    type Future struct { ch chan int }
//    func (f Future) Get() int { return <-f.ch }
//    func async(fn func() int) Future — launches fn in a goroutine
