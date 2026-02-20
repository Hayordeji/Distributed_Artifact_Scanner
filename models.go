package main

import "time"

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
	TotalFiles   int
	TotalBytes   int64
	FilesScanned int
	FilesPending int
	Duplicates   map[string][]string
	TypeCount    map[string]int
	Errors       []string
	StartTime    time.Time
	EndTime      time.Time
}

type ScanConfig struct {
	Directories []string
	WorkerCount int
	MaxFileSize int64
}
