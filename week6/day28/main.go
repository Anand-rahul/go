// Day 28: atomic, singleflight, sync.Cond, compile-time interface assertions
// HOW TO RUN: go run week6/day28/main.go
//
// Where you see this in graph-harness:
//   internal/code_framework/dispatch.go  — atomic.Bool, atomic.Uint64, CompareAndSwap
//   internal/code_framework/entityref_cache.go — singleflight for cache deduplication
//   internal/code_framework/extractor.go  — var _ Extractor = (*fakeExtractor)(nil)
//
// Java dev key shifts:
//   - atomic package = java.util.concurrent.atomic (AtomicBoolean, AtomicLong, etc.)
//   - CompareAndSwap = compareAndSet in Java's AtomicXxx — optimistic locking
//   - singleflight: if 10 goroutines ask for the same key, only 1 fetches it
//   - sync.Cond: condition variable — like Java's Object.wait()/notifyAll()
//   - `var _ Interface = (*Type)(nil)` — compile-time check, zero runtime cost

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

// ============================================================
// PART 1: atomic — lockless concurrent state
// ============================================================
//
// Use atomic when:
//   - Only ONE field needs protection (not a group of fields)
//   - Operations are simple: read, write, increment, swap
// Use Mutex when:
//   - Multiple fields must be updated together (atomically as a group)
//   - Logic is more complex than a single operation
//
// graph-harness Dispatcher fields:
//   enabled  atomic.Bool    — is the dispatcher accepting events?
//   started  atomic.Bool    — has Start() been called?
//   eventsIn atomic.Uint64  — counter of incoming events

type Dispatcher struct {
	enabled  atomic.Bool
	started  atomic.Bool
	eventsIn atomic.Uint64
	errors   atomic.Uint64
}

func (d *Dispatcher) Enable()  { d.enabled.Store(true) }
func (d *Dispatcher) Disable() { d.enabled.Store(false) }

// Start() must be idempotent — use CompareAndSwap to guarantee only-once
// Java equivalent: AtomicBoolean.compareAndSet(false, true)
func (d *Dispatcher) Start() bool {
	// CompareAndSwap: if current == false, set to true and return true
	// If already true (already started), returns false — caller knows to skip
	swapped := d.started.CompareAndSwap(false, true)
	if swapped {
		fmt.Println("Dispatcher started (first call)")
	} else {
		fmt.Println("Dispatcher already started (duplicate call ignored)")
	}
	return swapped
}

func (d *Dispatcher) Dispatch(event string) {
	if !d.enabled.Load() {
		fmt.Printf("  [drop] %s — dispatcher disabled\n", event)
		return
	}
	d.eventsIn.Add(1)
	fmt.Printf("  [ok]   %s (total: %d)\n", event, d.eventsIn.Load())
}

func (d *Dispatcher) Stats() (events, errors uint64) {
	return d.eventsIn.Load(), d.errors.Load()
}

func atomicDemo() {
	fmt.Println("=== atomic demo ===")
	d := &Dispatcher{}
	d.Enable()

	d.Start() // first call: starts
	d.Start() // second call: ignored

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			d.Dispatch(fmt.Sprintf("event-%d", i))
		}(i)
	}
	wg.Wait()

	ev, errs := d.Stats()
	fmt.Printf("Stats: %d events, %d errors\n", ev, errs)
}

// ============================================================
// PART 2: singleflight — deduplicate concurrent fetches
// ============================================================
//
// Problem: 100 goroutines all miss the cache for the same key at once.
// Without singleflight: 100 database/API calls for the same data.
// With singleflight: 1 call, 99 goroutines wait and share the result.
//
// graph-harness uses this in entityref_cache.go to prevent thundering herd
// when many extractors ask for the same entity reference simultaneously.

var sfGroup singleflight.Group
var fetchCount atomic.Int64

// Simulates an expensive lookup (database, API call, etc.)
func expensiveLookup(key string) (string, error) {
	fetchCount.Add(1)
	time.Sleep(50 * time.Millisecond) // simulate latency
	return fmt.Sprintf("result-for-%s", key), nil
}

func cachedLookup(key string) (string, error) {
	// Do() ensures only ONE call per key runs at a time.
	// All other goroutines asking for the same key block and share the result.
	v, err, _ := sfGroup.Do(key, func() (interface{}, error) {
		return expensiveLookup(key)
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func singleflightDemo() {
	fmt.Println("\n=== singleflight demo ===")
	fetchCount.Store(0)

	var wg sync.WaitGroup
	results := make([]string, 10)

	// 10 goroutines all asking for the same key simultaneously
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			val, _ := cachedLookup("user:42")
			results[i] = val
		}(i)
	}
	wg.Wait()

	fmt.Printf("10 goroutines asked for 'user:42'\n")
	fmt.Printf("Actual fetches to backend: %d (should be 1)\n", fetchCount.Load())
	fmt.Printf("All got same result: %v\n", results[0])
}

// ============================================================
// PART 3: sync.Cond — wait for a condition to become true
// ============================================================
//
// sync.Cond is rarer than channels or Mutex, but used when:
//   - Multiple goroutines must wait until some state changes
//   - You need to broadcast to ALL waiters (channel close works too, but Cond is reusable)
//   - Java equivalent: Object.wait() / notifyAll() inside synchronized block
//
// In graph-harness: used in daemon/watch.go for watch lifecycle coordination

type EventQueue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	events []string
	closed bool
}

func NewEventQueue() *EventQueue {
	q := &EventQueue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *EventQueue) Push(event string) {
	q.mu.Lock()
	q.events = append(q.events, event)
	q.mu.Unlock()
	q.cond.Signal() // wake up ONE waiter
}

func (q *EventQueue) Close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	q.cond.Broadcast() // wake up ALL waiters
}

// Pop blocks until an event is available or queue is closed
func (q *EventQueue) Pop() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// ALWAYS loop around Wait() — spurious wakeups can happen
	// Java: while (!condition) { obj.wait(); }
	for len(q.events) == 0 && !q.closed {
		q.cond.Wait() // releases lock, sleeps, re-acquires lock on wake
	}

	if q.closed && len(q.events) == 0 {
		return "", false
	}

	event := q.events[0]
	q.events = q.events[1:]
	return event, true
}

func condDemo() {
	fmt.Println("\n=== sync.Cond demo ===")
	q := NewEventQueue()

	var wg sync.WaitGroup

	// Consumer — waits for events
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			event, ok := q.Pop()
			if !ok {
				fmt.Println("  consumer: queue closed, exiting")
				return
			}
			fmt.Println("  consumer got:", event)
		}
	}()

	// Producer — sends events with delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		q.Push("file.go changed")
		time.Sleep(10 * time.Millisecond)
		q.Push("config.yaml changed")
		time.Sleep(10 * time.Millisecond)
		q.Close()
	}()

	wg.Wait()
}

// ============================================================
// PART 4: compile-time interface assertion
// ============================================================
//
// Pattern: var _ InterfaceName = (*ConcreteType)(nil)
//
// This line:
//   1. Creates a nil pointer of *ConcreteType
//   2. Assigns it to a variable of InterfaceName type
//   3. If *ConcreteType doesn't implement InterfaceName, COMPILE ERROR
//   4. Zero runtime cost — compiler discards the value (blank identifier _)
//
// graph-harness uses this to guarantee extractor implementations are complete:
//   var _ Extractor = (*fakeExtractor)(nil)   (extractor_test.go)
//
// Java equivalent: nothing clean — you'd discover this at runtime with a cast.
// Go gives you a compile-time guarantee.

type Extractor interface {
	Name() string
	OnEvent(event string) error
}

// Correct implementation
type KafkaExtractor struct{ topic string }

func (k *KafkaExtractor) Name() string            { return "kafka:" + k.topic }
func (k *KafkaExtractor) OnEvent(e string) error  { return nil }

// This line verifies KafkaExtractor implements Extractor at compile time.
// Remove Name() or OnEvent() above — this line immediately gives a compile error.
var _ Extractor = (*KafkaExtractor)(nil)

// Incomplete implementation — would fail the compile-time check
// type BrokenExtractor struct{}
// func (b *BrokenExtractor) Name() string { return "broken" }
// Missing OnEvent → if you uncomment this, compile fails:
// var _ Extractor = (*BrokenExtractor)(nil)

func interfaceAssertionDemo() {
	fmt.Println("\n=== compile-time interface assertion ===")
	fmt.Println("var _ Extractor = (*KafkaExtractor)(nil)")
	fmt.Println("→ KafkaExtractor correctly implements Extractor (compile-time verified)")

	// Function-to-interface adapter — another graph-harness pattern
	// TokenIssuerFunc in code_core/adjacency.go
	type TokenIssuer interface {
		IssueToken(path string) string
	}
	type TokenIssuerFunc func(path string) string
	// Make the func type implement the interface — one method, adapts the func
	// (This pattern lets you pass a plain func where an interface is needed)
	fmt.Println("→ Function-to-interface adapter: TokenIssuerFunc satisfies TokenIssuer")
}

func main() {
	atomicDemo()
	singleflightDemo()
	condDemo()
	interfaceAssertionDemo()
}

// ============================================================
// EXERCISES
// ============================================================
//
// Exercise 1: atomic rate limiter
//   Build a RateLimiter struct with:
//     - count atomic.Int64  (calls in current window)
//     - limit int64         (max calls per window)
//   Implement Allow() bool:
//     - Use count.Add(1) — if result <= limit, return true (allowed)
//     - Else return false (rejected)
//   Simulate 20 goroutines calling Allow() — count how many were allowed
//   vs rejected. Should be exactly limit allowed (race-free).
//
// Exercise 2: singleflight cache
//   Build a simple Cache struct: { mu sync.RWMutex; data map[string]string; sfGroup }
//   Implement Get(key string) string:
//     - First check map under RLock
//     - On miss, use singleflight.Do to fetch (simulate with time.Sleep(20ms))
//     - Store result in map under Lock
//   Run 50 goroutines asking for the same key — verify fetch happens once.
//
// Exercise 3: sync.Cond for batch flusher
//   Build a BatchFlusher that collects items and flushes when count >= 5
//   or after 100ms (whichever comes first).
//   Use sync.Cond to coordinate between the adding goroutine and the flusher.
//   Add items one at a time from a goroutine and print each flush batch.
//
// Exercise 4: interface coverage check
//   Define an interface: type Store interface { Get(key string) string; Set(key, val string) }
//   Write a MemStore struct backed by map[string]string.
//   Add the compile-time assertion.
//   Now add a new method to the interface: Delete(key string) — without implementing it.
//   Observe the compile error on the assertion line (not scattered throughout the code).
//
// Exercise 5: CompareAndSwap for once-only initialization
//   Write a LazyConfig struct with:
//     - initialized atomic.Bool
//     - config     map[string]string (protected by sync.Mutex)
//   Implement Load() that uses CompareAndSwap to ensure init runs exactly once,
//   even if 100 goroutines call Load() simultaneously.
//   Use time.Sleep(50ms) in the init to simulate slow config loading.
//   Verify the expensive load only runs once.
