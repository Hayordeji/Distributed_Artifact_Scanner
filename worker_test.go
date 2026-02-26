package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// TestProcessFile_Success tests successful file processing
func TestProcessFile_Success(t *testing.T) {
	// Setup: Create a temporary test file
	tempDir := t.TempDir() // Go automatically cleans this up
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("Hello, World!")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create task
	task := FileTask{
		Path: testFile,
		Size: int64(len(testContent)),
	}

	// Execute
	result := ProcessFiles(task)

	// Assert: Check result
	if result.Path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, result.Path)
	}

	if result.Size != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), result.Size)
	}

	if result.Error != "" {
		t.Errorf("Expected no error, got: %s", result.Error)
	}

	// Verify hash is correct
	expectedHash := sha256.Sum256(testContent)
	expectedHashStr := hex.EncodeToString(expectedHash[:])

	if result.Hash != expectedHashStr {
		t.Errorf("Expected hash %s, got %s", expectedHashStr, result.Hash)
	}

	// Verify file type
	if result.FileType != ".txt" {
		t.Errorf("Expected file type .txt, got %s", result.FileType)
	}
}

// TestProcessFile_NonExistent tests handling of missing files
func TestProcessFile_NonExistent(t *testing.T) {
	task := FileTask{
		Path: "/path/that/does/not/exist.txt",
		Size: 100,
	}

	result := ProcessFiles(task)

	// Should have an error
	if result.Error == "" {
		t.Error("Expected error for non-existent file, got none")
	}

	// Hash should be empty
	if result.Hash != "" {
		t.Errorf("Expected empty hash, got %s", result.Hash)
	}
}

// TestProcessFile_EmptyFile tests processing of empty files
func TestProcessFile_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "empty.txt")

	// Create empty file
	err := os.WriteFile(testFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	task := FileTask{
		Path: testFile,
		Size: 0,
	}

	result := ProcessFiles(task)

	// Should succeed
	if result.Error != "" {
		t.Errorf("Unexpected error: %s", result.Error)
	}

	// Empty file has a known hash
	expectedHash := sha256.Sum256([]byte{})
	expectedHashStr := hex.EncodeToString(expectedHash[:])

	if result.Hash != expectedHashStr {
		t.Errorf("Expected hash %s for empty file, got %s", expectedHashStr, result.Hash)
	}
}

// TestProcessFile_NoExtension tests files without extensions
func TestProcessFile_NoExtension(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "README")

	err := os.WriteFile(testFile, []byte("No extension"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	task := FileTask{
		Path: testFile,
		Size: 12,
	}

	result := ProcessFiles(task)

	if result.Error != "" {
		t.Errorf("Unexpected error: %s", result.Error)
	}

	// Files without extension should have empty FileType
	if result.FileType != "" {
		t.Errorf("Expected empty file type, got %s", result.FileType)
	}
}

// TestProcessFile_DifferentTypes tests various file extensions
func TestProcessFile_DifferentTypes(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantType string
	}{
		{"PDF file", "document.pdf", ".pdf"},
		{"Image file", "photo.jpg", ".jpg"},
		{"Excel file", "data.xlsx", ".xlsx"},
		{"Multiple dots", "file.tar.gz", ".gz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, tt.filename)

			err := os.WriteFile(testFile, []byte("test content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			task := FileTask{Path: testFile, Size: 12}
			result := ProcessFiles(task)

			if result.FileType != tt.wantType {
				t.Errorf("Expected file type %s, got %s", tt.wantType, result.FileType)
			}
		})
	}
}
