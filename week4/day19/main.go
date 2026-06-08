// Day 19: net/http — HTTP Server and Client
// HOW TO RUN: go run week4/day19/main.go
// Then in another terminal: curl http://localhost:8080/hello
//
// Java dev key shifts:
//   - http.ListenAndServe is your embedded server — no Tomcat/Jetty needed
//   - Handler interface: ServeHTTP(ResponseWriter, *Request)
//   - http.HandleFunc registers path → function (like @GetMapping in Spring)
//   - No dependency injection — just pass what you need as function arguments
//   - Middleware = a function that wraps an http.Handler (like Spring filters)
//   - http.Client is the HTTP client (like RestTemplate or OkHttpClient)
//   - Go's stdlib HTTP is production-ready for many use cases

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// === HANDLER ===
// Java: @RestController with @GetMapping
// Go: any function func(w http.ResponseWriter, r *http.Request)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// w is like HttpServletResponse, r is like HttpServletRequest
	name := r.URL.Query().Get("name") // ?name=rahul
	if name == "" {
		name = "World"
	}
	fmt.Fprintf(w, "Hello, %s!\n", name)
}

// === HANDLER WITH STATE ===
// Java: @Service / @Component injected into @RestController
// Go: struct that implements http.Handler interface
type CounterHandler struct {
	count int
}

func (h *CounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.count++
	fmt.Fprintf(w, "request #%d\n", h.count)
}

// === JSON RESPONSE ===
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	users := []User{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// === ROUTING WITH PATH PARAMETERS ===
// Note: Go 1.22+ has built-in path params: mux.HandleFunc("GET /users/{id}", ...)
// Before 1.22: parse from URL manually or use gorilla/mux etc.
func userByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Path: /user/123 — extract "123" manually (pre-1.22 style)
	path := strings.TrimPrefix(r.URL.Path, "/user/")
	if path == "" {
		http.Error(w, "missing user ID", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(User{ID: 1, Name: path, Email: path + "@example.com"})
}

// === MIDDLEWARE ===
// Java: Filter / HandlerInterceptor
// Go: a function that takes an http.Handler and returns an http.Handler
type Middleware func(http.Handler) http.Handler

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("→ %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("← %s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "Bearer secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Chain multiple middleware
func chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// === HTTP CLIENT ===
// Java: RestTemplate, WebClient, OkHttpClient
func demoClient() {
	client := &http.Client{
		Timeout: 5 * time.Second, // always set a timeout!
	}

	resp, err := client.Get("http://localhost:8080/hello?name=Rahul")
	if err != nil {
		log.Println("client error:", err)
		return
	}
	defer resp.Body.Close() // ALWAYS close the body

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("read error:", err)
		return
	}
	fmt.Printf("client got: status=%d body=%q\n", resp.StatusCode, body)

	// POST with JSON body
	payload := `{"name":"Rahul","email":"rahul@example.com"}`
	resp2, err := client.Post(
		"http://localhost:8080/users",
		"application/json",
		strings.NewReader(payload),
	)
	if err == nil {
		defer resp2.Body.Close()
		fmt.Println("POST status:", resp2.Status)
	}
}

func main() {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/users", usersHandler)
	mux.HandleFunc("/user/", userByIDHandler)
	mux.Handle("/counter", &CounterHandler{})

	// Protected route with middleware
	protectedHandler := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "secret data")
		}),
		loggingMiddleware,
		authMiddleware,
	)
	mux.Handle("/protected", protectedHandler)

	// Wrap entire mux with logging
	handler := loggingMiddleware(mux)

	// Start client in background after server starts
	go func() {
		time.Sleep(100 * time.Millisecond)
		demoClient()
	}()

	// Shutdown after 3 seconds for demo
	go func() {
		time.Sleep(3 * time.Second)
		log.Println("shutting down demo server")
		// In real apps use: http.Server{}.Shutdown(ctx)
		// Here we just exit the program
	}()

	fmt.Println("server starting on :8080")
	fmt.Println("try: curl http://localhost:8080/hello?name=Rahul")
	fmt.Println("try: curl http://localhost:8080/users")
	fmt.Println("try: curl -H 'Authorization: Bearer secret' http://localhost:8080/protected")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// === EXERCISES ===
// 1. Add a POST /users endpoint that:
//    - Reads a JSON body into a User struct
//    - Validates that Name and Email are not empty
//    - Returns 201 Created with the user, or 400 Bad Request with error
//
// 2. Add a request-timeout middleware:
//    func timeoutMiddleware(d time.Duration, next http.Handler) http.Handler
//    Use context.WithTimeout on r.Context().
//
// 3. Write a retry-capable HTTP client:
//    func getWithRetry(url string, maxRetries int) (*http.Response, error)
//    Retry on 5xx responses or network errors with exponential backoff.
//
// 4. Add a /health endpoint returning JSON:
//    {"status":"ok","uptime":"123s","version":"1.0.0"}
//    Store server start time in a package-level variable.
//
// 5. Implement basic path routing:
//    GET /items      → list items
//    GET /items/{id} → get single item
//    POST /items     → create item
//    Use a map[string]Item as in-memory storage (protect with sync.RWMutex).
