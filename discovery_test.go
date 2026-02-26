package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDiscoverFiles_BasicDiscovery(t *testing.T) {
	// Setup: Create test directory structure
	tempDir := t.TempDir()

	// Create test files
	files := []string{
		"file1.txt",
		"file2.pdf",
		"subdir/file3.txt",
	}

	for _, f := range files {
		fullPath := filepath.Join(tempDir, f)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte("test"), 0644)
	}

	// Setup channels
	tasksChannel := make(chan FileTask, 10)
	doneChannel := make(chan struct{})
	metrics := &ScanMetrics{
		Duplicates: make(map[string][]string),
		TypeCount:  make(map[string]int),
	}
	metricsMutex := &sync.RWMutex{}

	// Run discovery
	config := ScanConfig{
		Directories: []string{tempDir},
		MaxFileSize: 1024 * 1024,
	}

	go DiscoverFiles(config, tasksChannel, doneChannel, metrics, metricsMutex)

	// Collect discovered tasks
	var discoveredTasks []FileTask
	for task := range tasksChannel {
		discoveredTasks = append(discoveredTasks, task)
	}

	// Assert: Should find 3 files
	if len(discoveredTasks) != 3 {
		t.Errorf("Expected 3 files discovered, got %d", len(discoveredTasks))
	}

	// Check TotalFiles was updated
	metricsMutex.RLock()
	if metrics.TotalFiles != 3 {
		t.Errorf("Expected TotalFiles = 3, got %d", metrics.TotalFiles)
	}
	metricsMutex.RUnlock()
}

// TestDiscoverFiles_MaxSizeFilter tests file size filtering
func TestDiscoverFiles_MaxSizeFilter(t *testing.T) {
	tempDir := t.TempDir()

	// Create small and large files
	smallFile := filepath.Join(tempDir, "small.txt")
	largeFile := filepath.Join(tempDir, "large.txt")

	os.WriteFile(smallFile, []byte("small"), 0644)         // 5 bytes
	os.WriteFile(largeFile, make([]byte, 1024*1024), 0644) // 1MB

	tasksChannel := make(chan FileTask, 10)
	doneChannel := make(chan struct{})
	metrics := &ScanMetrics{
		Duplicates: make(map[string][]string),
		TypeCount:  make(map[string]int),
	}
	metricsMutex := &sync.RWMutex{}

	config := ScanConfig{
		Directories: []string{tempDir},
		MaxFileSize: 1000, // Only 1000 bytes max
	}

	go DiscoverFiles(config, tasksChannel, doneChannel, metrics, metricsMutex)

	var discoveredTasks []FileTask
	for task := range tasksChannel {
		discoveredTasks = append(discoveredTasks, task)
	}

	// Should only find small file
	if len(discoveredTasks) != 1 {
		t.Errorf("Expected 1 file (small), got %d", len(discoveredTasks))
	}

	if len(discoveredTasks) > 0 && discoveredTasks[0].Path != smallFile {
		t.Errorf("Expected small file, got %s", discoveredTasks[0].Path)
	}
}

// TestDiscoverFiles_Cancellation tests discovery cancellation
func TestDiscoverFiles_Cancellation(t *testing.T) {
	tempDir := t.TempDir()

	// Create many files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("file%d.txt", i))
		os.WriteFile(filename, []byte("test"), 0644)
	}

	tasksChannel := make(chan FileTask, 10)
	doneChannel := make(chan struct{})
	metrics := &ScanMetrics{
		Duplicates: make(map[string][]string),
		TypeCount:  make(map[string]int),
	}
	metricsMutex := &sync.RWMutex{}

	config := ScanConfig{
		Directories: []string{tempDir},
		MaxFileSize: 1024 * 1024,
	}

	discoveryDone := make(chan struct{})
	go func() {
		DiscoverFiles(config, tasksChannel, doneChannel, metrics, metricsMutex)
		close(discoveryDone)
	}()

	// Cancel immediately
	close(doneChannel)

	// Discovery should exit quickly
	select {
	case <-discoveryDone:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Discovery did not exit after cancellation")
	}
}
