package main

import (
	"fmt"
	"sync"
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
			metrics.TotalBytes += result.Size

			if result.Error != "" {
				metrics.Errors = append(metrics.Errors, result.Error)
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
