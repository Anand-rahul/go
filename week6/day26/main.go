// Day 26: database/sql — SQLite, Queries, Transactions
// HOW TO RUN: go run week6/day26/main.go
// REQUIRES:   go get modernc.org/sqlite  (pure-Go SQLite, no cgo needed)
//
// Where you see this in graph-harness:
//   internal/code_core/adjacency.go     — entity graph stored in SQLite
//   internal/facts/facts.go             — event log backed by SQLite
//
// Java dev key shifts:
//   - database/sql is the stdlib interface — drivers plug in separately
//   - No ORM by default — raw SQL, but clean and explicit
//   - sql.DB is a connection pool, NOT a single connection
//   - sql.Tx is a transaction — commit or rollback
//   - Rows must be closed (defer rows.Close()) — like ResultSet in JDBC
//   - Scan() maps columns to Go variables — like ResultSet.getString()
//   - db.QueryRow().Scan() for single-row queries (no need to check rows.Next)
//   - errors.Is(err, sql.ErrNoRows) — the "not found" sentinel
//   - Use sql.Conn for connection-scoped operations (graph-harness uses this)

package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "modernc.org/sqlite" // blank import — registers the "sqlite" driver via init()
)

// === DOMAIN TYPES ===
// graph-harness stores entities like this:
type Entity struct {
	ID   int64
	Kind string
	Name string
	Path string
}

// === OPEN A DATABASE ===
// sql.Open does NOT connect yet — it just validates the driver name
// The actual connection happens on first use
func openDB() *sql.DB {
	// ":memory:" = in-memory database (gone when process exits)
	// For files: sql.Open("sqlite", "/path/to/file.db")
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	// db.Ping() actually connects and checks the connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

// === DDL: CREATE TABLE ===
func createSchema(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS entities (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			kind TEXT    NOT NULL,
			name TEXT    NOT NULL,
			path TEXT    NOT NULL
		)
	`)
	if err != nil {
		log.Fatal("create table:", err)
	}
	fmt.Println("Schema created")
}

// === INSERT ===
// db.Exec returns (sql.Result, error)
// sql.Result has LastInsertId() and RowsAffected()
func insertEntity(db *sql.DB, kind, name, path string) (int64, error) {
	result, err := db.Exec(
		`INSERT INTO entities (kind, name, path) VALUES (?, ?, ?)`,
		kind, name, path, // positional args — no SQL injection risk
	)
	if err != nil {
		return 0, fmt.Errorf("insert entity: %w", err)
	}
	return result.LastInsertId()
}

// === QUERY ONE ROW ===
// db.QueryRow + Scan — for "get by ID" style lookups
// Java: preparedStatement.executeQuery() → resultSet.next() → resultSet.getX()
func getEntity(db *sql.DB, id int64) (Entity, error) {
	var e Entity
	err := db.QueryRow(
		`SELECT id, kind, name, path FROM entities WHERE id = ?`, id,
	).Scan(&e.ID, &e.Kind, &e.Name, &e.Path)

	if errors.Is(err, sql.ErrNoRows) {
		return Entity{}, fmt.Errorf("entity %d not found", id)
	}
	if err != nil {
		return Entity{}, fmt.Errorf("get entity: %w", err)
	}
	return e, nil
}

// === QUERY MULTIPLE ROWS ===
// db.Query returns *sql.Rows — must be closed
// rows.Next() advances the cursor — like ResultSet.next()
func listByKind(db *sql.DB, kind string) ([]Entity, error) {
	rows, err := db.Query(
		`SELECT id, kind, name, path FROM entities WHERE kind = ? ORDER BY id`,
		kind,
	)
	if err != nil {
		return nil, fmt.Errorf("list entities: %w", err)
	}
	defer rows.Close() // ALWAYS defer close — leaks connection if not closed

	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Kind, &e.Name, &e.Path); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		entities = append(entities, e)
	}
	// rows.Err() captures any error that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return entities, nil
}

// === TRANSACTIONS ===
// db.Begin() returns a *sql.Tx
// Tx has the same Exec/Query/QueryRow methods as *sql.DB
// ALWAYS either Commit or Rollback — use defer for safety
//
// graph-harness uses db.BeginTx(ctx, nil) — same but with context
func insertBatch(db *sql.DB, entities []Entity) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() // no-op if already committed

	stmt, err := tx.Prepare(
		`INSERT INTO entities (kind, name, path) VALUES (?, ?, ?)`,
	)
	if err != nil {
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, e := range entities {
		if _, err := stmt.Exec(e.Kind, e.Name, e.Path); err != nil {
			return fmt.Errorf("insert %s: %w", e.Name, err)
		}
	}

	return tx.Commit()
}

// === sql.Conn — connection-scoped operations ===
// graph-harness uses sql.Conn to run multiple statements on the SAME connection
// (important for SQLite: PRAGMA, WAL mode, temp tables are connection-local)
func demonstrateConn(db *sql.DB) {
	conn, err := db.Conn(nil) // nil context → context.Background()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// SQLite PRAGMA — must run on same connection as subsequent queries
	conn.ExecContext(nil, `PRAGMA journal_mode = WAL`)
	fmt.Println("PRAGMA set on dedicated connection")
}

func main() {
	db := openDB()
	defer db.Close()

	createSchema(db)

	// Single insert
	id1, _ := insertEntity(db, "route", "GET /users", "internal/api/users.go")
	id2, _ := insertEntity(db, "route", "POST /users", "internal/api/users.go")
	id3, _ := insertEntity(db, "event", "UserCreated", "internal/events/user.go")

	fmt.Printf("Inserted IDs: %d, %d, %d\n", id1, id2, id3)

	// Single row lookup
	e, err := getEntity(db, id1)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Entity: %+v\n", e)
	}

	// Not found — errors.Is(err, sql.ErrNoRows) would be true inside getEntity
	_, err = getEntity(db, 999)
	fmt.Println("Not found:", err)

	// Query many
	routes, _ := listByKind(db, "route")
	fmt.Printf("Routes: %d found\n", len(routes))
	for _, r := range routes {
		fmt.Printf("  [%d] %s → %s\n", r.ID, r.Kind, r.Name)
	}

	// Batch insert in a transaction
	batch := []Entity{
		{Kind: "event", Name: "UserDeleted", Path: "internal/events/user.go"},
		{Kind: "event", Name: "OrderPlaced", Path: "internal/events/order.go"},
	}
	if err := insertBatch(db, batch); err != nil {
		fmt.Println("Batch error:", err)
	} else {
		fmt.Println("Batch inserted OK")
	}

	events, _ := listByKind(db, "event")
	fmt.Printf("Events: %d found\n", len(events))
}

// ============================================================
// EXERCISES
// ============================================================
//
// Exercise 1: UpdateEntity
//   Write func updatePath(db *sql.DB, id int64, newPath string) error
//   Use db.Exec with UPDATE ... WHERE id = ?
//   After updating, verify with getEntity that the path changed.
//
// Exercise 2: DeleteEntity
//   Write func deleteEntity(db *sql.DB, id int64) error
//   Check RowsAffected() — return an error if nothing was deleted
//   (This is how you distinguish "delete succeeded" from "ID didn't exist")
//
// Exercise 3: Count query
//   Write func countByKind(db *sql.DB, kind string) (int, error)
//   Use `SELECT COUNT(*) FROM entities WHERE kind = ?`
//   Use db.QueryRow(...).Scan(&count) — single integer result
//
// Exercise 4: Transaction rollback
//   Write a function that inserts 3 entities in a transaction,
//   but returns an error on the 3rd insert (simulate: if i == 2 { return error }).
//   The Rollback in defer should undo inserts 1 and 2.
//   Verify by counting rows before and after — count should be unchanged.
//
// Exercise 5: Prepared statement reuse
//   In graph-harness, prepared statements are stored on the struct and reused
//   across many calls (see adjacency.go — stmts map[string]*sql.Stmt).
//   Write a type EntityStore struct { db *sql.DB; insertStmt *sql.Stmt }
//   with a New(db) constructor that prepares the INSERT statement once,
//   and an Insert(kind, name, path) method that reuses it.
//   Why: preparing once avoids re-parsing SQL on every call.
