package main

import (
	"fmt"
	"sync"
	"time"
)

func CollectResults(resultsChannel chan ScanResult, doneChannel chan struct{}, metrics *ScanMetrics, metricsMutex *sync.RWMutex) {

	metrics.Duplicates = make(map[string][]string)
	metrics.TypeCount = make(map[string]int)
	for {
		select {
		case result, ok := <-resultsChannel:
			if !ok {
				fmt.Printf("CollectResults channel closed.\n")
				return
			}

			metricsMutex.Lock()

			metrics.FilesScanned++
			metrics.FilesPending = metrics.TotalFiles - metrics.FilesScanned
			metrics.TotalBytes += result.Size

			if result.Error != "" {
				newError := FileError{
					Path:  result.Path,
					Error: result.Error,
					Time:  time.Now(),
				}
				metrics.Errors = append(metrics.Errors, newError)
				metricsMutex.Unlock()
				continue
			}

			//CHECK IF HASH EXISTS IN DUPLICATE
			if existingPaths, exists := metrics.Duplicates[result.Hash]; exists {
				metrics.Duplicates[result.Hash] = append(existingPaths, result.Path)
			} else {
				metrics.Duplicates[result.Hash] = []string{result.Path}
			}

			metrics.TypeCount[result.FileType]++
			metricsMutex.Unlock()

		case <-doneChannel:
			return
		}
	}

}

func CollectRealMetrics(metrics *ScanMetrics) ScanMetrics {
	//CREATE A NEW METRICS
	metricsCopy := ScanMetrics{
		TotalFiles:   metrics.TotalFiles,
		TotalBytes:   metrics.TotalBytes,
		FilesScanned: metrics.FilesScanned,
		FilesPending: metrics.FilesPending,
		StartTime:    metrics.StartTime,
		EndTime:      metrics.EndTime,
		Duplicates:   make(map[string][]string, len(metrics.Duplicates)),
		TypeCount:    make(map[string]int, len(metrics.TypeCount)),

		Errors: make([]FileError, len(metrics.Errors)),
	}

	actualDuplicates := make(map[string][]string)
	duplicateFiles := 0

	//COPY DUPLICATES DATA TO THE NEW METRICS
	for hash, paths := range metrics.Duplicates {
		//pathsCopy := make([]string, len(paths))
		if len(paths) >= 2 {
			actualDuplicates[hash] = paths
			duplicateFiles += len(paths) - 1
		}
	}

	//ADD DUPLICATES AND DUPLICATES FILE COUNT
	metricsCopy.Duplicates = actualDuplicates
	metricsCopy.DuplicateFilesCount = duplicateFiles

	//COPY ALL TYPE COUNT FROM EXISTING SERVER TYPE COUNTS TO NEW TYPE COUNTS
	for ext, count := range metrics.TypeCount {
		metricsCopy.TypeCount[ext] = count
	}
	copy(metricsCopy.Errors, metrics.Errors)

	//RECORD END TIME
	if !metrics.EndTime.IsZero() {
		metricsCopy.EndTime = metrics.EndTime
	}

	return metricsCopy
}
