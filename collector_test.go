package main

import (
	"sync"
	"testing"
	"time"
)

func TestCollectResults_BasicAggregation(t *testing.T) {
	// Setup
	resultsChannel := make(chan ScanResult, 10)
	doneChannel := make(chan struct{})
	metrics := &ScanMetrics{
		StartTime:  time.Now(),
		Duplicates: make(map[string][]string),
		TypeCount:  make(map[string]int),
		Errors:     make([]FileError, 0),
	}
	metricsMutex := &sync.RWMutex{}

	// Start collector
	collectorDone := make(chan struct{})
	go func() {
		CollectResults(resultsChannel, doneChannel, metrics, metricsMutex)
		close(collectorDone)
	}()

	// Send test results
	resultsChannel <- ScanResult{
		Path:     "/file1.txt",
		Hash:     "abc123",
		FileType: ".txt",
		Size:     100,
	}
	resultsChannel <- ScanResult{
		Path:     "/file2.txt",
		Hash:     "def456",
		FileType: ".txt",
		Size:     200,
	}
	resultsChannel <- ScanResult{
		Path:     "/file3.pdf",
		Hash:     "xyz789",
		FileType: ".pdf",
		Size:     300,
	}

	// Close channel and wait for collector
	close(resultsChannel)
	<-collectorDone

	// Assert
	metricsMutex.RLock()
	defer metricsMutex.RUnlock()

	if metrics.FilesScanned != 3 {
		t.Errorf("Expected 3 files scanned, got %d", metrics.FilesScanned)
	}

	if metrics.TotalBytes != 600 {
		t.Errorf("Expected 600 total bytes, got %d", metrics.TotalBytes)
	}

	if len(metrics.Duplicates) != 3 {
		t.Errorf("Expected 3 unique hashes, got %d", len(metrics.Duplicates))
	}

	if metrics.TypeCount[".txt"] != 2 {
		t.Errorf("Expected 2 .txt files, got %d", metrics.TypeCount[".txt"])
	}

	if metrics.TypeCount[".pdf"] != 1 {
		t.Errorf("Expected 1 .pdf file, got %d", metrics.TypeCount[".pdf"])
	}
}

// TestCollectResults_DuplicateDetection tests duplicate file detection
func TestCollectResults_DuplicateDetection(t *testing.T) {
	resultsChannel := make(chan ScanResult, 10)
	doneChannel := make(chan struct{})
	metrics := &ScanMetrics{
		StartTime:  time.Now(),
		Duplicates: make(map[string][]string),
		TypeCount:  make(map[string]int),
		Errors:     make([]FileError, 0),
	}
	metricsMutex := &sync.RWMutex{}

	collectorDone := make(chan struct{})
	go func() {
		CollectResults(resultsChannel, doneChannel, metrics, metricsMutex)
		close(collectorDone)
	}()

	// Send files with same hash (duplicates)
	sameHash := "duplicate-hash-123"
	resultsChannel <- ScanResult{Path: "/file1.txt", Hash: sameHash, FileType: ".txt", Size: 100}
	resultsChannel <- ScanResult{Path: "/file2.txt", Hash: sameHash, FileType: ".txt", Size: 100}
	resultsChannel <- ScanResult{Path: "/subdir/file3.txt", Hash: sameHash, FileType: ".txt", Size: 100}

	close(resultsChannel)
	<-collectorDone

	metricsMutex.RLock()
	defer metricsMutex.RUnlock()

	// Check duplicates were detected
	duplicatePaths := metrics.Duplicates[sameHash]
	if len(duplicatePaths) != 3 {
		t.Errorf("Expected 3 duplicate files, got %d", len(duplicatePaths))
	}

	expectedPaths := map[string]bool{
		"/file1.txt":        true,
		"/file2.txt":        true,
		"/subdir/file3.txt": true,
	}

	for _, path := range duplicatePaths {
		if !expectedPaths[path] {
			t.Errorf("Unexpected path in duplicates: %s", path)
		}
	}
}

// TestCollectResults_ErrorHandling tests error collection
func TestCollectResults_ErrorHandling(t *testing.T) {
	resultsChannel := make(chan ScanResult, 10)
	doneChannel := make(chan struct{})
	metrics := &ScanMetrics{
		StartTime:  time.Now(),
		Duplicates: make(map[string][]string),
		TypeCount:  make(map[string]int),
		Errors:     make([]FileError, 0),
	}
	metricsMutex := &sync.RWMutex{}

	collectorDone := make(chan struct{})
	go func() {
		CollectResults(resultsChannel, doneChannel, metrics, metricsMutex)
		close(collectorDone)

	}()

	// Send result with error
	resultsChannel <- ScanResult{
		Path:  "/bad-file.txt",
		Size:  100,
		Error: "permission denied",
	}

	// Send successful result
	resultsChannel <- ScanResult{
		Path:     "/good-file.txt",
		Hash:     "abc123",
		FileType: ".txt",
		Size:     200,
	}

	close(resultsChannel)
	<-collectorDone

	metricsMutex.RLock()
	defer metricsMutex.RUnlock()

	// Both should be counted as "scanned" (attempted)
	if metrics.FilesScanned != 2 {
		t.Errorf("Expected 2 files scanned, got %d", metrics.FilesScanned)
	}

	// Error should be recorded
	if len(metrics.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(metrics.Errors))
	}

	if metrics.Errors[0].Path != "/bad-file.txt" {
		t.Errorf("Expected error for /bad-file.txt, got %s", metrics.Errors[0].Path)
	}

	// Only successful file should be in duplicates
	if len(metrics.Duplicates) != 1 {
		t.Errorf("Expected 1 unique hash, got %d", len(metrics.Duplicates))
	}
}

// TestCollectResults_Cancellation tests graceful cancellation
func TestCollectResults_Cancellation(t *testing.T) {
	resultsChannel := make(chan ScanResult, 10)
	doneChannel := make(chan struct{})
	metrics := &ScanMetrics{
		StartTime:  time.Now(),
		Duplicates: make(map[string][]string),
		TypeCount:  make(map[string]int),
		Errors:     make([]FileError, 0),
	}
	metricsMutex := &sync.RWMutex{}

	collectorDone := make(chan struct{})
	go func() {
		CollectResults(resultsChannel, doneChannel, metrics, metricsMutex)
		close(collectorDone)
	}()

	// Send a result
	resultsChannel <- ScanResult{Path: "/file1.txt", Hash: "abc", FileType: ".txt", Size: 100}

	// Trigger cancellation
	close(doneChannel)

	// Collector should exit quickly
	select {
	case <-collectorDone:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Collector did not exit after cancellation")
	}

	metricsMutex.RLock()
	if metrics.FilesScanned != 1 {
		t.Errorf("Expected 1 file scanned before cancel, got %d", metrics.FilesScanned)
	}
	metricsMutex.RUnlock()
}
