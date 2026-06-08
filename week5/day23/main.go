// Day 23: Common Go Patterns — Functional Options, Middleware, Options Pattern
// HOW TO RUN: go run week5/day23/main.go
//
// These are patterns you'll see in almost every real Go codebase.
// Understanding these elevates you from beginner to mid-level.

package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ============================================================
// PATTERN 1: FUNCTIONAL OPTIONS
// Problem: constructor with many optional parameters
// Java: Builder pattern (MyObject.builder().timeout(5).maxRetries(3).build())
// Go: variadic options functions
// ============================================================

type Server struct {
	host        string
	port        int
	timeout     time.Duration
	maxConns    int
	tlsEnabled  bool
	rateLimitRPS int
}

type ServerOption func(*Server)

func WithHost(host string) ServerOption {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port int) ServerOption {
	return func(s *Server) {
		s.port = port
	}
}

func WithTimeout(d time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = d
	}
}

func WithMaxConnections(n int) ServerOption {
	return func(s *Server) {
		s.maxConns = n
	}
}

func WithTLS() ServerOption {
	return func(s *Server) {
		s.tlsEnabled = true
	}
}

func WithRateLimit(rps int) ServerOption {
	return func(s *Server) {
		s.rateLimitRPS = rps
	}
}

func NewServer(opts ...ServerOption) *Server {
	// Sensible defaults
	s := &Server{
		host:     "localhost",
		port:     8080,
		timeout:  30 * time.Second,
		maxConns: 100,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ============================================================
// PATTERN 2: MIDDLEWARE CHAIN (HTTP)
// Java: Spring Filters / HandlerInterceptors
// ============================================================

type HandlerFunc func(http.ResponseWriter, *http.Request)
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[LOG] %s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func RateLimit(rps int) Middleware {
	// Simplified — real implementation uses token bucket
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("[RATE:%d] allowing request\n", rps)
			next.ServeHTTP(w, r)
		})
	}
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ============================================================
// PATTERN 3: RESULT TYPE
// Go doesn't have Result<T, E> (Rust) but you can make one
// ============================================================

type Result[T any] struct {
	value T
	err   error
}

func Ok[T any](v T) Result[T]      { return Result[T]{value: v} }
func Err[T any](e error) Result[T] { return Result[T]{err: e} }

func (r Result[T]) IsOk() bool     { return r.err == nil }
func (r Result[T]) Unwrap() T      {
	if r.err != nil {
		panic(fmt.Sprintf("unwrap on error result: %v", r.err))
	}
	return r.value
}
func (r Result[T]) UnwrapOr(def T) T {
	if r.err != nil {
		return def
	}
	return r.value
}

// ============================================================
// PATTERN 4: CONTEXT VALUES WITH TYPED KEYS
// Thread-safe request-scoped data
// ============================================================

type contextKey int

const (
	keyUserID contextKey = iota
	keyRequestID
	keyLogger
)

type Logger2 struct{ prefix string }
func (l *Logger2) Log(msg string) { fmt.Printf("[%s] %s\n", l.prefix, msg) }

func ContextWithUser(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, keyUserID, userID)
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(keyUserID).(int)
	return id, ok
}

// ============================================================
// PATTERN 5: OPTION TYPE (nullable without pointers)
// Avoid *T for optional values — use a wrapper
// ============================================================

type Optional[T any] struct {
	value   T
	present bool
}

func Some[T any](v T) Optional[T]   { return Optional[T]{value: v, present: true} }
func None[T any]() Optional[T]      { return Optional[T]{} }

func (o Optional[T]) Get() (T, bool) { return o.value, o.present }
func (o Optional[T]) OrElse(def T) T {
	if o.present {
		return o.value
	}
	return def
}
func (o Optional[T]) IsPresent() bool { return o.present }

// ============================================================
// PATTERN 6: GRACEFUL SHUTDOWN
// ============================================================

type GracefulServer struct {
	srv  *http.Server
	done chan struct{}
}

func (gs *GracefulServer) Shutdown(ctx context.Context) error {
	defer close(gs.done)
	fmt.Println("shutting down server...")
	return gs.srv.Shutdown(ctx)
}

func (gs *GracefulServer) WaitForShutdown() {
	<-gs.done
}

func main() {
	// === FUNCTIONAL OPTIONS ===
	fmt.Println("=== Functional Options ===")

	// Minimal — use defaults
	s1 := NewServer()
	fmt.Printf("default: %s:%d timeout=%v\n", s1.host, s1.port, s1.timeout)

	// Full customization
	s2 := NewServer(
		WithHost("0.0.0.0"),
		WithPort(9090),
		WithTimeout(60*time.Second),
		WithMaxConnections(1000),
		WithTLS(),
		WithRateLimit(100),
	)
	fmt.Printf("custom: %s:%d tls=%v rateLimit=%d\n",
		s2.host, s2.port, s2.tlsEnabled, s2.rateLimitRPS)

	// Composable: partial configs
	prodOpts := []ServerOption{
		WithHost("0.0.0.0"),
		WithTLS(),
		WithRateLimit(500),
	}
	devOpts := []ServerOption{
		WithHost("localhost"),
		WithPort(3000),
	}
	sProd := NewServer(prodOpts...)
	sDev := NewServer(devOpts...)
	fmt.Printf("prod: %s tls=%v\n", sProd.host, sProd.tlsEnabled)
	fmt.Printf("dev: %s:%d\n", sDev.host, sDev.port)

	// === MIDDLEWARE CHAIN ===
	fmt.Println("\n=== Middleware Chain ===")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("[HANDLER] request processed")
	})

	// Build chain: request goes Logger → RateLimit → Auth → handler
	chained := Chain(handler, Logger, RateLimit(10), Auth)

	// Simulate a request
	req, _ := http.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Authorization", "Bearer token")
	chained.ServeHTTP(nil, req)

	// === RESULT TYPE ===
	fmt.Println("\n=== Result Type ===")
	r1 := Ok(42)
	r2 := Err[int](fmt.Errorf("something went wrong"))

	fmt.Println("r1 ok:", r1.IsOk(), "value:", r1.Unwrap())
	fmt.Println("r2 ok:", r2.IsOk(), "default:", r2.UnwrapOr(-1))

	// === CONTEXT TYPED KEYS ===
	fmt.Println("\n=== Typed Context Values ===")
	ctx := ContextWithUser(context.Background(), 42)
	if id, ok := UserIDFromContext(ctx); ok {
		fmt.Println("user ID from context:", id)
	}

	// === OPTIONAL TYPE ===
	fmt.Println("\n=== Optional ===")
	name := Some("Rahul")
	noName := None[string]()

	fmt.Println("name:", name.OrElse("anonymous"))
	fmt.Println("noName:", noName.OrElse("anonymous"))

	if v, ok := name.Get(); ok {
		fmt.Println("got name:", v)
	}
}

// === EXERCISES ===
// 1. Add WithLogger(l *log.Logger) option to NewServer.
//    The server should log requests using this logger.
//    If not provided, use a no-op logger.
//
// 2. Create a database client using functional options:
//    type DBClient struct { host, user, password string; maxOpen, maxIdle int }
//    Options: WithDBHost, WithCredentials(user, pass), WithPool(maxOpen, maxIdle)
//    Add validation: error if host is empty.
//
// 3. Write a middleware that adds a request ID to the context:
//    Generate a UUID-like ID (fmt.Sprintf("req-%d", rand.Int63()))
//    and attach to context with a typed key.
//    All subsequent middleware and handlers can retrieve it.
//
// 4. The Result[T] type above is functional.
//    Add: func MapResult[T, U any](r Result[T], fn func(T) U) Result[U]
//    that transforms the value if it's Ok, or propagates the error.
//
// 5. Convert this Java builder to Go functional options:
//    HttpClient.newBuilder()
//      .connectTimeout(Duration.ofSeconds(5))
//      .followRedirects(Redirect.NORMAL)
//      .build()
