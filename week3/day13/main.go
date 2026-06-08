// Day 13: select, sync.WaitGroup, sync.Mutex, sync.Once, sync.RWMutex
// HOW TO RUN: go run week3/day13/main.go
//
// Java dev key shifts:
//   - select is like a switch for channels — waits on multiple channel ops
//   - sync.WaitGroup = CountDownLatch (wait for N goroutines to finish)
//   - sync.Mutex = synchronized block or ReentrantLock
//   - sync.RWMutex = ReadWriteLock — many readers OR one writer
//   - sync.Once = @Lazy / double-checked locking, guaranteed safe
//   - sync.Map = ConcurrentHashMap (but prefer mutex + map for custom logic)
//   - atomic package = java.util.concurrent.atomic package

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// === SELECT ===
// Like a switch that operates on channel send/receive operations
// Picks whichever case is ready; if multiple ready, picks randomly
func selectDemo() {
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	go func() {
		time.Sleep(30 * time.Millisecond)
		ch1 <- "one"
	}()
	go func() {
		time.Sleep(20 * time.Millisecond)
		ch2 <- "two"
	}()

	// Block until one is ready
	select {
	case msg := <-ch1:
		fmt.Println("received from ch1:", msg)
	case msg := <-ch2:
		fmt.Println("received from ch2:", msg) // this wins (20ms < 30ms)
	}

	// select with default — non-blocking channel check
	select {
	case msg := <-ch1:
		fmt.Println("ch1:", msg)
	default:
		fmt.Println("ch1 not ready yet")
	}
}

// === SELECT WITH DONE AND TICK ===
func ticker(done <-chan struct{}) {
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()

	count := 0
	for {
		select {
		case <-tick.C:
			count++
			fmt.Printf("tick %d\n", count)
		case <-done:
			fmt.Println("ticker stopped after", count, "ticks")
			return
		}
	}
}

// === MUTEX ===
// Java: synchronized(this) { } or Lock lock = new ReentrantLock();
type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock() // idiomatic — defer ensures unlock even on panic
	c.count++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// === RW MUTEX — optimal for read-heavy workloads ===
// Java: ReadWriteLock
type SafeCache struct {
	mu    sync.RWMutex
	store map[string]string
}

func NewSafeCache() *SafeCache {
	return &SafeCache{store: make(map[string]string)}
}

func (c *SafeCache) Set(key, value string) {
	c.mu.Lock() // exclusive write lock
	defer c.mu.Unlock()
	c.store[key] = value
}

func (c *SafeCache) Get(key string) (string, bool) {
	c.mu.RLock() // shared read lock — multiple readers OK simultaneously
	defer c.mu.RUnlock()
	v, ok := c.store[key]
	return v, ok
}

// === SYNC.ONCE — run something exactly once ===
// Java: double-checked locking pattern, or @PostConstruct singleton
type Singleton struct {
	Config string
}

var (
	instance *Singleton
	once     sync.Once
)

func GetInstance() *Singleton {
	once.Do(func() {
		fmt.Println("initializing singleton (runs exactly once)")
		instance = &Singleton{Config: "production"}
	})
	return instance
}

// === ATOMIC OPERATIONS ===
// Java: AtomicInteger
// For simple counters, atomics are faster than mutex
type AtomicCounter struct {
	val int64
}

func (c *AtomicCounter) Increment() {
	atomic.AddInt64(&c.val, 1)
}

func (c *AtomicCounter) Value() int64 {
	return atomic.LoadInt64(&c.val)
}

// === SYNC.MAP — concurrent map ===
// Use when: many goroutines read/write, keys are stable once written
// For custom logic, mutex + regular map is usually better

func syncMapDemo() {
	var m sync.Map

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			m.Store(key, i*10)
		}()
	}
	wg.Wait()

	m.Range(func(k, v any) bool {
		fmt.Printf("  %v = %v\n", k, v)
		return true // return false to stop iteration
	})
}

func main() {
	fmt.Println("=== SELECT ===")
	selectDemo()

	fmt.Println("\n=== TICKER + DONE ===")
	done := make(chan struct{})
	go ticker(done)
	time.Sleep(45 * time.Millisecond)
	close(done)
	time.Sleep(10 * time.Millisecond)

	fmt.Println("\n=== MUTEX ===")
	counter := &SafeCounter{}
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	wg.Wait()
	fmt.Println("safe count:", counter.Value()) // always 1000

	fmt.Println("\n=== RW MUTEX CACHE ===")
	cache := NewSafeCache()
	cache.Set("lang", "Go")
	cache.Set("version", "1.22")

	var rwWg sync.WaitGroup
	for i := 0; i < 5; i++ {
		rwWg.Add(1)
		go func() {
			defer rwWg.Done()
			if v, ok := cache.Get("lang"); ok {
				fmt.Println("  read lang:", v)
			}
		}()
	}
	rwWg.Wait()

	fmt.Println("\n=== SYNC.ONCE ===")
	var sOnce sync.WaitGroup
	for i := 0; i < 3; i++ {
		sOnce.Add(1)
		go func() {
			defer sOnce.Done()
			inst := GetInstance()
			fmt.Println("  got instance:", inst.Config)
		}()
	}
	sOnce.Wait()

	fmt.Println("\n=== ATOMIC COUNTER ===")
	ac := &AtomicCounter{}
	var acWg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		acWg.Add(1)
		go func() {
			defer acWg.Done()
			ac.Increment()
		}()
	}
	acWg.Wait()
	fmt.Println("atomic count:", ac.Value()) // always 1000

	fmt.Println("\n=== SYNC.MAP ===")
	syncMapDemo()
}

// === EXERCISES ===
// 1. Write a concurrent task queue:
//    - A Queue with Add(task func()) and a worker pool of N goroutines
//    - Workers pick tasks off the queue and run them
//    - Use channels (buffered) as the queue, WaitGroup to wait for completion
//
// 2. Implement a simple connection pool using a buffered channel as the semaphore.
//    Pool size 3, 10 goroutines competing for connections.
//    Each "connection" is just an int. Print which goroutine got which connection.
//
// 3. Use select to implement a "first result wins":
//    Launch 3 goroutines that each sleep a random time then send a result.
//    Return the FIRST result and cancel the rest using a done channel.
//
// 4. Write a rate limiter using time.Tick:
//    Allow max N operations per second. Use a token bucket pattern.
//    Test by trying to do 20 operations with a limit of 5/second.
//
// 5. What's the difference between:
//    a) sync.Mutex  b) sync.RWMutex  c) sync/atomic  d) channel
//    For each, give a scenario where it's the best choice.
