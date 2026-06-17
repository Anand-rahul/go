// Break Day 1 (Week 3): Goroutines & Channels — Deeper Practice
// HOW TO RUN: go run week3/break01/main.go
//
// Day 11 taught you goroutines + WaitGroup.
// Day 12 taught you channels, pipelines, semaphores, Future.
// This file drills the patterns that show up constantly in real Go services
// (and in graph-harness specifically):
//
//   1. Error propagation from goroutines — the most common beginner mistake
//   2. Generator pattern — infinite/lazy sequences via channels
//   3. Rate limiting — time.Ticker to control throughput
//   4. Or-done pattern — reading from a channel while respecting cancellation
//   5. Fan-out fan-in with results — N workers process jobs, results merged
//   6. Heartbeat — goroutine proves it's alive on a separate channel
//
// Java mappings are noted inline.

package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================================
// PATTERN 1: ERROR PROPAGATION FROM GOROUTINES
// ============================================================
//
// The #1 beginner mistake: launching goroutines that can fail but
// have no way to return the error to the caller.
//
// Wrong approach: goroutine just fmt.Println("error: ...") and exits silently.
// Right approach: send errors on an error channel alongside results.
//
// Java: CompletableFuture.exceptionally() handles this automatically.
// Go: you wire it manually — but you have full control.
//
// Pattern: results chan + errs chan, both closed when work is done.
// Caller reads results and errs concurrently.

type Result struct {
	URL  string
	Body string
}

// fetchAll launches one goroutine per URL.
// Returns (results channel, errors channel) — both closed when all goroutines finish.
func fetchAll(urls []string) (<-chan Result, <-chan error) {
	results := make(chan Result, len(urls))
	errs := make(chan error, len(urls))

	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		url := url
		go func() {
			defer wg.Done()
			body, err := simulateFetch(url)
			if err != nil {
				errs <- err // send error, don't panic, don't print
				return
			}
			results <- Result{URL: url, Body: body}
		}()
	}

	// Close both channels when ALL goroutines are done
	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	return results, errs
}

func simulateFetch(url string) (string, error) {
	time.Sleep(time.Duration(rand.Intn(30)+10) * time.Millisecond)
	if url == "url3" || url == "url5" {
		return "", fmt.Errorf("fetch failed: %s unreachable", url)
	}
	return fmt.Sprintf("<html>content from %s</html>", url), nil
}

func errorPropagationDemo() {
	fmt.Println("=== 1. Error propagation from goroutines ===")

	urls := []string{"url1", "url2", "url3", "url4", "url5"}
	results, errs := fetchAll(urls)

	// Collect results and errors concurrently.
	// IMPORTANT: drain BOTH channels fully before stopping —
	// goroutines sending to full channels would deadlock.
	var allResults []Result
	var allErrors []error

	// Use a WaitGroup to drain both channels in parallel
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for r := range results {
			allResults = append(allResults, r)
		}
	}()
	go func() {
		defer wg.Done()
		for e := range errs {
			allErrors = append(allErrors, e)
		}
	}()

	wg.Wait()

	fmt.Printf("  %d succeeded, %d failed\n", len(allResults), len(allErrors))
	for _, r := range allResults {
		fmt.Printf("  ✓ %s → %d bytes\n", r.URL, len(r.Body))
	}
	for _, e := range allErrors {
		fmt.Printf("  ✗ %v\n", e)
	}
}

// ============================================================
// PATTERN 2: GENERATOR — lazy infinite sequence via channel
// ============================================================
//
// A generator is a goroutine that produces values on demand.
// The channel acts as a lazy iterator — values are computed only
// when the consumer is ready to receive.
//
// Java: Stream.iterate() or an Iterator<T>. Go's version is simpler.
//
// Key: always pair a generator with a done/quit channel so the
// goroutine can exit when the consumer stops early.
// Without it, the goroutine leaks (blocks on send forever).

// integers generates 0, 1, 2, 3, ... until done is closed.
func integers(done <-chan struct{}) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; ; i++ {
			select {
			case out <- i: // send next value when consumer is ready
			case <-done:   // consumer signalled "stop"
				return
			}
		}
	}()
	return out
}

// fibonacci generates Fibonacci numbers until done is closed.
func fibonacci(done <-chan struct{}) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		a, b := 0, 1
		for {
			select {
			case out <- a:
				a, b = b, a+b
			case <-done:
				return
			}
		}
	}()
	return out
}

func generatorDemo() {
	fmt.Println("\n=== 2. Generator pattern ===")

	done := make(chan struct{})

	// Take first 8 integers
	nums := integers(done)
	fmt.Print("  integers: ")
	for i := 0; i < 8; i++ {
		fmt.Print(<-nums, " ")
	}
	close(done) // stop the generator goroutine — no leak
	fmt.Println()

	// Fibonacci — take until value > 100
	done2 := make(chan struct{})
	fibs := fibonacci(done2)
	fmt.Print("  fibonacci (≤100): ")
	for v := range fibs {
		if v > 100 {
			close(done2) // signal generator to stop
			break
		}
		fmt.Print(v, " ")
	}
	fmt.Println()
}

// ============================================================
// PATTERN 3: RATE LIMITING
// ============================================================
//
// Control how fast you process events. Two forms:
//   - Steady rate: time.Ticker — fires at a fixed interval
//   - Burst + steady: allow N requests immediately, then throttle
//
// graph-harness uses this to pace how fast it fires LSP requests.
// Java: RateLimiter from Guava, or ScheduledExecutorService.

func rateLimitDemo() {
	fmt.Println("\n=== 3. Rate limiting ===")

	requests := []string{"req-1", "req-2", "req-3", "req-4", "req-5"}

	// === Steady rate: 1 request per 30ms ===
	ticker := time.NewTicker(30 * time.Millisecond)
	defer ticker.Stop()

	start := time.Now()
	fmt.Print("  steady (1/30ms): ")
	for _, req := range requests {
		<-ticker.C // wait for tick before processing
		fmt.Printf("%s@%dms ", req, time.Since(start).Milliseconds())
	}
	fmt.Println()

	// === Burst + steady: allow 3 immediately, then 1/50ms ===
	// Buffered channel pre-filled = burst capacity
	burst := make(chan struct{}, 3)
	for i := 0; i < 3; i++ {
		burst <- struct{}{} // pre-fill burst tokens
	}

	// Refill one token every 50ms
	go func() {
		refill := time.NewTicker(50 * time.Millisecond)
		defer refill.Stop()
		for range refill.C {
			select {
			case burst <- struct{}{}: // add token if space available
			default: // burst channel full — discard
			}
		}
	}()

	reqs := []string{"A", "B", "C", "D", "E", "F"}
	start2 := time.Now()
	fmt.Print("  burst+steady: ")
	for _, req := range reqs {
		<-burst // consume a token — blocks if burst exhausted
		fmt.Printf("%s@%dms ", req, time.Since(start2).Milliseconds())
	}
	fmt.Println()
}

// ============================================================
// PATTERN 4: OR-DONE — read from channel, respect cancellation
// ============================================================
//
// Problem: you have a source channel and a done channel.
// You want to read from source, but stop immediately when done closes.
//
// Without or-done, you'd write this select everywhere. With it,
// you wrap the source into a new channel that auto-closes when done fires.
//
// graph-harness uses this everywhere — every goroutine that reads
// from an event stream must also watch ctx.Done() (the channel form).

// orDone wraps a channel so receiving from it also respects a done signal.
// Returns a channel that closes when either src closes or done closes.
func orDone(done, src <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case v, ok := <-src:
				if !ok {
					return // src closed
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

func orDoneDemo() {
	fmt.Println("\n=== 4. Or-done pattern ===")

	// Source produces values slowly
	src := make(chan int)
	go func() {
		defer close(src)
		for i := 1; i <= 20; i++ {
			time.Sleep(10 * time.Millisecond)
			src <- i
		}
	}()

	// We only want 5 values — cancel after that
	done := make(chan int)
	go func() {
		time.Sleep(55 * time.Millisecond) // cancel after ~5 values (5 × 10ms)
		close(done)
	}()

	fmt.Print("  got values: ")
	for v := range orDone(done, src) {
		fmt.Print(v, " ")
	}
	fmt.Println("(cancelled, didn't wait for all 20)")
}

// ============================================================
// PATTERN 5: FAN-OUT FAN-IN WITH RESULTS
// ============================================================
//
// Full pattern:
//   jobs channel → N worker goroutines → N result channels → merge → output
//
// Each worker reads from the shared jobs channel (fan-out).
// Each worker writes to its own result channel.
// A merger reads all result channels into one output (fan-in).
//
// This is different from day12's simple merge — here the workers
// DO WORK between reading and writing. The fan-in is over worker outputs,
// not arbitrary input channels.
//
// graph-harness: Dispatcher routes file-change events to N extractors (fan-out),
// each extractor writes entities to a shared store (implicit fan-in).
// Java: ExecutorService.submit() → List<Future<Result>> → combine.

type Job struct {
	ID    int
	Input string
}

type WorkResult struct {
	JobID  int
	Output string
	Err    error
}

func fanOutFanInDemo() {
	fmt.Println("\n=== 5. Fan-out fan-in with results ===")

	const numWorkers = 3
	const numJobs = 9

	// ── fan-out: single jobs channel shared by all workers ──
	jobs := make(chan Job, numJobs)
	for i := 1; i <= numJobs; i++ {
		jobs <- Job{ID: i, Input: fmt.Sprintf("data-%d", i)}
	}
	close(jobs) // close so workers' range loops terminate

	// ── each worker gets its OWN result channel ──
	workerChans := make([]<-chan WorkResult, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workerChans[i] = startWorker(i+1, jobs)
	}

	// ── fan-in: merge all worker channels into one ──
	merged := mergeResults(workerChans...)

	// ── collect ──
	var results []WorkResult
	for r := range merged {
		results = append(results, r)
	}

	fmt.Printf("  processed %d jobs with %d workers\n", len(results), numWorkers)
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("  ✗ job-%d: %v\n", r.JobID, r.Err)
		} else {
			fmt.Printf("  ✓ job-%d: %s\n", r.JobID, r.Output)
		}
	}
}

func startWorker(id int, jobs <-chan Job) <-chan WorkResult {
	out := make(chan WorkResult)
	go func() {
		defer close(out)
		for job := range jobs {
			time.Sleep(10 * time.Millisecond) // simulate work
			out <- WorkResult{
				JobID:  job.ID,
				Output: fmt.Sprintf("worker-%d processed %s", id, job.Input),
			}
		}
	}()
	return out
}

func mergeResults(cs ...<-chan WorkResult) <-chan WorkResult {
	out := make(chan WorkResult)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		c := c
		go func() {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// ============================================================
// PATTERN 6: HEARTBEAT
// ============================================================
//
// A long-running goroutine publishes a "pulse" on a separate channel
// at a regular interval. The parent watches both the heartbeat and a
// results channel. If the heartbeat stops, the goroutine is stuck.
//
// graph-harness uses this in daemon/watch.go to detect whether the
// file watcher goroutine is still alive.
// Java: ScheduledFuture polling a health flag, or a watchdog thread.

func longRunningWorker(done <-chan struct{}) (results <-chan string, heartbeat <-chan time.Time) {
	r := make(chan string)
	hb := make(chan time.Time, 1) // buffered so send never blocks

	go func() {
		defer close(r)
		defer close(hb)

		ticker := time.NewTicker(20 * time.Millisecond) // heartbeat interval
		defer ticker.Stop()

		workItems := []string{"alpha", "beta", "gamma", "delta"}
		i := 0

		for {
			select {
			case <-done:
				return

			case t := <-ticker.C:
				// Send heartbeat — non-blocking (buffer=1, drop if full)
				select {
				case hb <- t:
				default:
				}

				// Do one unit of work per tick
				if i < len(workItems) {
					time.Sleep(5 * time.Millisecond) // simulate processing
					select {
					case r <- workItems[i]:
						i++
					case <-done:
						return
					}
				}
			}
		}
	}()

	return r, hb
}

func heartbeatDemo() {
	fmt.Println("\n=== 6. Heartbeat pattern ===")

	done := make(chan struct{})
	results, heartbeat := longRunningWorker(done)

	timeout := time.After(200 * time.Millisecond)
	var collected []string

	for {
		select {
		case v, ok := <-results:
			if !ok {
				fmt.Printf("  worker done, collected: %v\n", collected)
				return
			}
			collected = append(collected, v)
			fmt.Printf("  result: %s\n", v)

		case t := <-heartbeat:
			fmt.Printf("  ♥ heartbeat at %dms\n", t.UnixMilli()%10000)

		case <-timeout:
			fmt.Println("  parent timed out — stopping worker")
			close(done)
			// drain remaining results
			for v := range results {
				collected = append(collected, v)
			}
			fmt.Printf("  collected before stop: %v\n", collected)
			return
		}
	}
}

func main() {
	errorPropagationDemo()
	generatorDemo()
	rateLimitDemo()
	orDoneDemo()
	fanOutFanInDemo()
	heartbeatDemo()
}

var _ = errors.New // keep errors import used (exercises reference it)

// ============================================================
// EXERCISES
// ============================================================
//
// Exercise 1: First-result wins (race multiple sources)
//   Write: func fastest(urls []string) (string, error)
//   Launch one goroutine per URL. Each "fetches" (random 10-80ms sleep).
//   Return the FIRST successful result — cancel the rest.
//   Pattern: results chan (buffered=1), done chan.
//   First goroutine to write to results wins. Others see done closed and exit.
//   Java equivalent: CompletableFuture.anyOf().
//
// Exercise 2: Retry with backoff
//   Write: func withRetry(fn func() error, maxRetries int) error
//   Uses a goroutine + channel internally.
//   On failure, wait 2^attempt * 10ms before retrying.
//   On success or max retries, send result on a channel and return.
//   Test with a fn that fails twice then succeeds.
//
// Exercise 3: Bounded generator with take()
//   Write: func take(done <-chan struct{}, src <-chan int, n int) <-chan int
//   It forwards the first n values from src, then closes done and returns.
//   Use the integers() generator from this file as the source.
//   Collect take(done, integers(done), 10) — should get exactly [0..9].
//
// Exercise 4: Pipeline with error short-circuit
//   Build a 3-stage pipeline: generate → validate → transform
//   generate: emits integers 1..10
//   validate: passes through only odd numbers, drops evens
//   transform: multiplies by 100
//   If transform sees a value > 700, it sends an error and stops the pipeline.
//   Use a shared errs chan. The main goroutine reads results until errs fires.
//
// Exercise 5: Concurrent map with results
//   Write: func concurrentMap[T, R any](items []T, fn func(T) R) []R
//   Launch one goroutine per item, each calls fn(item) and sends to results chan.
//   Preserve ORDER in the output — results[i] must correspond to items[i].
//   Hint: pre-allocate []R of the same length, use index to write directly
//   (safe because each goroutine writes a unique index — no mutex needed).
//   Test: concurrentMap([]int{1,2,3,4,5}, func(n int) int { return n*n })
//   Expected: [1, 4, 9, 16, 25]
