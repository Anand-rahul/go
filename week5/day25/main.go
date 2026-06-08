// Day 25: Mini Project — CLI Task Manager
// HOW TO RUN: go run week5/day25/main.go [command] [args...]
//
// This ties together:
//   - Structs, interfaces, methods (Week 1-2)
//   - Error handling (Day 10)
//   - JSON persistence (Day 20)
//   - File I/O (Day 18)
//   - Context (Day 14)
//   - Functional options (Day 23)
//   - Generics (Day 21)
//   - Testing pattern (Day 17)
//
// Commands:
//   go run week5/day25/main.go add "Buy groceries" --priority high
//   go run week5/day25/main.go list
//   go run week5/day25/main.go list --filter pending
//   go run week5/day25/main.go done <id>
//   go run week5/day25/main.go delete <id>
//   go run week5/day25/main.go clear

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// === DOMAIN TYPES ===

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusDone     Status = "done"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Priority    Priority  `json:"priority"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func (t Task) String() string {
	mark := "[ ]"
	if t.Status == StatusDone {
		mark = "[x]"
	}
	prio := map[Priority]string{
		PriorityLow:    "⬇",
		PriorityMedium: "➡",
		PriorityHigh:   "⬆",
	}[t.Priority]
	return fmt.Sprintf("%s %d. %s %s (created: %s)",
		mark, t.ID, prio, t.Title, t.CreatedAt.Format("2006-01-02"))
}

// === STORAGE ===

type Storage struct {
	filepath string
}

type store struct {
	Tasks  []Task `json:"tasks"`
	NextID int    `json:"next_id"`
}

func NewStorage(path string) *Storage {
	return &Storage{filepath: path}
}

func (s *Storage) load() (*store, error) {
	data, err := os.ReadFile(s.filepath)
	if os.IsNotExist(err) {
		return &store{NextID: 1}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("storage.load: %w", err)
	}

	var st store
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("storage.load: corrupted data: %w", err)
	}
	return &st, nil
}

func (s *Storage) save(st *store) error {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("storage.save: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.filepath), 0755); err != nil {
		return fmt.Errorf("storage.save: mkdir: %w", err)
	}
	return os.WriteFile(s.filepath, data, 0644)
}

// === ERRORS ===

var ErrTaskNotFound = errors.New("task not found")

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: %s — %s", e.Field, e.Message)
}

// === SERVICE ===

type TaskService struct {
	storage *Storage
}

func NewTaskService(storage *Storage) *TaskService {
	return &TaskService{storage: storage}
}

func (svc *TaskService) Add(title string, priority Priority) (*Task, error) {
	if strings.TrimSpace(title) == "" {
		return nil, &ValidationError{Field: "title", Message: "cannot be empty"}
	}
	if priority == "" {
		priority = PriorityMedium
	}
	if priority != PriorityLow && priority != PriorityMedium && priority != PriorityHigh {
		return nil, &ValidationError{Field: "priority", Message: "must be low, medium, or high"}
	}

	st, err := svc.storage.load()
	if err != nil {
		return nil, err
	}

	task := Task{
		ID:        st.NextID,
		Title:     strings.TrimSpace(title),
		Priority:  priority,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
	st.Tasks = append(st.Tasks, task)
	st.NextID++

	if err := svc.storage.save(st); err != nil {
		return nil, err
	}
	return &task, nil
}

func (svc *TaskService) List(filter Status) ([]Task, error) {
	st, err := svc.storage.load()
	if err != nil {
		return nil, err
	}

	if filter == "" {
		return st.Tasks, nil
	}

	var filtered []Task
	for _, t := range st.Tasks {
		if t.Status == filter {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

func (svc *TaskService) Complete(id int) error {
	st, err := svc.storage.load()
	if err != nil {
		return err
	}

	for i, t := range st.Tasks {
		if t.ID == id {
			if st.Tasks[i].Status == StatusDone {
				return fmt.Errorf("task %d is already done", id)
			}
			now := time.Now()
			st.Tasks[i].Status = StatusDone
			st.Tasks[i].CompletedAt = &now
			return svc.storage.save(st)
		}
	}
	return fmt.Errorf("task %d: %w", id, ErrTaskNotFound)
}

func (svc *TaskService) Delete(id int) error {
	st, err := svc.storage.load()
	if err != nil {
		return err
	}

	for i, t := range st.Tasks {
		if t.ID == id {
			st.Tasks = append(st.Tasks[:i], st.Tasks[i+1:]...)
			return svc.storage.save(st)
		}
	}
	return fmt.Errorf("task %d: %w", id, ErrTaskNotFound)
}

func (svc *TaskService) Clear() error {
	return svc.storage.save(&store{NextID: 1})
}

// === CLI ===

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: tasks <add|list|done|delete|clear> [args...]")
	}

	homeDir, _ := os.UserHomeDir()
	storage := NewStorage(filepath.Join(homeDir, ".tasks", "tasks.json"))
	svc := NewTaskService(storage)

	command := args[0]
	rest := args[1:]

	switch command {
	case "add":
		if len(rest) == 0 {
			return fmt.Errorf("add requires a title")
		}
		title := rest[0]
		priority := PriorityMedium
		for i, a := range rest[1:] {
			if a == "--priority" && i+2 < len(rest) {
				priority = Priority(rest[i+2])
			}
		}
		task, err := svc.Add(title, priority)
		if err != nil {
			return err
		}
		fmt.Printf("added: %s\n", task)

	case "list":
		var filter Status
		for i, a := range rest {
			if a == "--filter" && i+1 < len(rest) {
				filter = Status(rest[i+1])
			}
		}
		tasks, err := svc.List(filter)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			fmt.Println("no tasks")
			return nil
		}
		for _, t := range tasks {
			fmt.Println(t)
		}
		fmt.Printf("\ntotal: %d\n", len(tasks))

	case "done":
		if len(rest) == 0 {
			return fmt.Errorf("done requires a task ID")
		}
		id, err := strconv.Atoi(rest[0])
		if err != nil {
			return fmt.Errorf("invalid ID %q", rest[0])
		}
		if err := svc.Complete(id); err != nil {
			return err
		}
		fmt.Printf("marked task %d as done\n", id)

	case "delete":
		if len(rest) == 0 {
			return fmt.Errorf("delete requires a task ID")
		}
		id, err := strconv.Atoi(rest[0])
		if err != nil {
			return fmt.Errorf("invalid ID %q", rest[0])
		}
		if err := svc.Delete(id); err != nil {
			return err
		}
		fmt.Printf("deleted task %d\n", id)

	case "clear":
		if err := svc.Clear(); err != nil {
			return err
		}
		fmt.Println("all tasks cleared")

	default:
		return fmt.Errorf("unknown command %q (use: add, list, done, delete, clear)", command)
	}

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// === EXERCISES — Extend This Project ===
// 1. Add a 'due' flag: --due "2026-06-10"
//    Parse with time.Parse and store as *time.Time.
//    Show tasks overdue in the list with a warning marker.
//
// 2. Add tags: --tags "work,urgent"
//    Filter by tag: tasks list --tag work
//    Store as []string in the JSON.
//
// 3. Write tests in day25_test.go:
//    - TestAdd: verify title validation, default priority
//    - TestComplete: verify ErrTaskNotFound for invalid ID
//    - TestList: verify filter works
//    Use an in-memory storage (implement Storage interface) to avoid
//    file system dependency in tests.
//
// 4. Add goroutine-based auto-save:
//    Buffer writes and flush every 5 seconds, or immediately on shutdown.
//    Use a channel to receive write requests and a ticker for periodic flush.
//
// 5. Refactor: extract a Repository interface:
//    type Repository interface {
//        Load() (*store, error)
//        Save(*store) error
//    }
//    Implement JSONRepository (current) and InMemoryRepository (for tests).
//    Inject Repository into TaskService — proper dependency injection Go-style.
