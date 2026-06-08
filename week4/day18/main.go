// Day 18: File I/O — os, io, bufio, filepath
// HOW TO RUN: go run week4/day18/main.go
//
// Java dev key shifts:
//   - os.Open returns *os.File + error (no checked exceptions)
//   - Always use defer f.Close() immediately after opening
//   - bufio.Scanner is your line-by-line reader (like BufferedReader + readLine)
//   - io.ReadAll reads everything at once (like Files.readAllBytes)
//   - os.WriteFile / os.ReadFile — one-liners for small files
//   - filepath.Join uses OS-appropriate separator (like Paths.get())
//   - io.Copy — efficient piping between reader and writer

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// === WRITE A FILE ===
	// os.WriteFile — simple, reads whole file into memory
	// Java: Files.writeString(Path.of("file.txt"), content)
	content := "Hello, Go!\nLine 2\nLine 3\nLine 4\n"
	err := os.WriteFile("/tmp/golearn_day18.txt", []byte(content), 0644)
	if err != nil {
		fmt.Println("write error:", err)
		return
	}
	fmt.Println("wrote /tmp/golearn_day18.txt")

	// === READ A FILE (all at once) ===
	// Java: Files.readString(Path.of("file.txt"))
	data, err := os.ReadFile("/tmp/golearn_day18.txt")
	if err != nil {
		fmt.Println("read error:", err)
		return
	}
	fmt.Printf("read %d bytes:\n%s", len(data), data)

	// === READ LINE BY LINE with bufio.Scanner ===
	// Java: try (BufferedReader br = new BufferedReader(new FileReader(f))) { ... }
	f, err := os.Open("/tmp/golearn_day18.txt")
	if err != nil {
		fmt.Println("open error:", err)
		return
	}
	defer f.Close() // ALWAYS defer close right after successful open

	fmt.Println("--- line by line ---")
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		fmt.Printf("  %d: %s\n", lineNum, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("scan error:", err)
	}

	// === WRITE LINE BY LINE with bufio.Writer ===
	// Buffered writes are more efficient for many small writes
	out, err := os.Create("/tmp/golearn_day18_out.txt")
	if err != nil {
		fmt.Println("create error:", err)
		return
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for i := 1; i <= 5; i++ {
		fmt.Fprintf(w, "buffered line %d\n", i)
	}
	w.Flush() // IMPORTANT: flush buffered data to file
	fmt.Println("wrote buffered file")

	// === APPEND TO A FILE ===
	appendF, err := os.OpenFile("/tmp/golearn_day18.txt",
		os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("append error:", err)
		return
	}
	defer appendF.Close()
	fmt.Fprintln(appendF, "Line 5 (appended)")

	// === IO.COPY — efficient piping ===
	// Java: IOUtils.copy() from Apache Commons or Files.copy()
	src, _ := os.Open("/tmp/golearn_day18.txt")
	defer src.Close()

	dst, _ := os.Create("/tmp/golearn_day18_copy.txt")
	defer dst.Close()

	n, err := io.Copy(dst, src)
	fmt.Printf("copied %d bytes\n", n)

	// === FILEPATH — OS-agnostic path handling ===
	// Java: Paths.get(), Path.of(), File.separator
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "myapp", "config.json")
	fmt.Println("config path:", configPath)

	dir := filepath.Dir(configPath)
	base := filepath.Base(configPath)
	ext := filepath.Ext(configPath)
	fmt.Printf("dir=%s base=%s ext=%s\n", dir, base, ext)

	// filepath.Walk — recursive directory traversal
	// Java: Files.walk(Path) stream
	fmt.Println("\n--- walking /tmp (first 5 entries) ---")
	count := 0
	err = filepath.Walk("/tmp", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if count >= 5 {
			return filepath.SkipDir // stop walking
		}
		if !info.IsDir() {
			fmt.Println(" ", path)
			count++
		}
		return nil
	})
	if err != nil {
		fmt.Println("walk error:", err)
	}

	// === FILE INFO ===
	info, err := os.Stat("/tmp/golearn_day18.txt")
	if err == nil {
		fmt.Printf("\nfile: %s size=%d perm=%v\n",
			info.Name(), info.Size(), info.Mode())
	}

	// === CHECK IF FILE EXISTS ===
	// Java: new File(path).exists()
	if _, err := os.Stat("/tmp/nonexistent"); os.IsNotExist(err) {
		fmt.Println("file does not exist")
	}

	// === STRING READER (in-memory file-like) ===
	// Java: new StringReader("content")
	sr := strings.NewReader("line1\nline2\nline3\n")
	sc := bufio.NewScanner(sr)
	for sc.Scan() {
		fmt.Println("in-memory:", sc.Text())
	}

	// Cleanup
	os.Remove("/tmp/golearn_day18.txt")
	os.Remove("/tmp/golearn_day18_out.txt")
	os.Remove("/tmp/golearn_day18_copy.txt")
	fmt.Println("\ncleanup done")
}

// === EXERCISES ===
// 1. Write a function countLinesWords(filename string) (lines, words int, err error)
//    that counts lines and words in a file using bufio.Scanner.
//
// 2. Write a CSV parser:
//    func parseCSV(filename string) ([][]string, error)
//    that returns rows as slices of strings.
//    (Bonus: use the standard "encoding/csv" package)
//
// 3. Write a function copyDir(src, dst string) error that recursively
//    copies a directory tree. Use filepath.Walk + os.MkdirAll.
//
// 4. Write a tail(filename string, n int) ([]string, error) function
//    that returns the last N lines of a file efficiently.
//    (Don't read the whole file into memory if you can avoid it)
//
// 5. Write a log rotator:
//    func rotate(logFile string, maxSize int64) error
//    If the file exceeds maxSize bytes, rename it to logFile.1 and create fresh logFile.
