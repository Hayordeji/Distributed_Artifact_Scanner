package main

import (
	"sync"
	"time"
)

type FileTask struct {
	Path string
	Size int64
}

type ScanResult struct {
	Path     string
	Size     int64
	Hash     string
	FileType string
	Error    string
}

type ScanMetrics struct {
	TotalFiles          int
	TotalBytes          int64
	FilesScanned        int
	FilesPending        int
	Duplicates          map[string][]string
	DuplicateFilesCount int
	TypeCount           map[string]int
	Errors              []FileError
	StartTime           time.Time
	EndTime             time.Time
}

type ScanConfig struct {
	Directories []string
	WorkerCount int
	MaxFileSize int64
}

type ServerContext struct {
	Metrics      *ScanMetrics
	MetricsMutex *sync.RWMutex
	CancelFunc   func()
}

type FileError struct {
	Path  string    `json:"path"`
	Error string    `json:"error"`
	Time  time.Time `json:"time"`
}
