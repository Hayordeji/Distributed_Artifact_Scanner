package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

func main() {

	//GET CONFIG FROM CMD ARGUMENTS
	var (
		dirFlag     = flag.String("dir", ".", "Directory containing files")
		workersFlag = flag.Int("workers", 4, "Number of concurrent workers")
		maxSizeFlag = flag.Int64("max-size", 100*1024*1024, "Maximum amount of files to scan")
	)

	flag.Parse()

	config := ScanConfig{
		Directories: []string{*dirFlag},
		WorkerCount: *workersFlag,
		MaxFileSize: *maxSizeFlag,
	}

	//INITIALIZE CHANNELS
	tasksChannel := make(chan FileTask, 100)
	resultsChannel := make(chan ScanResult, 100)
	doneChannel := make(chan struct{})

	metrics := &ScanMetrics{
		StartTime:  time.Now(),
		Duplicates: make(map[string][]string),
		TypeCount:  make(map[string]int),
		Errors:     make([]string, 0),
	}

	metricsMutex := &sync.RWMutex{}

	//	CREATE NEW SERVER
	server := NewServer(metrics, doneChannel, metricsMutex)
	server.Start()

	//GOROUTINE FOR DISCOVERING FILES
	go DiscoverFiles(config, tasksChannel, doneChannel)

	var workerWaitGroup sync.WaitGroup
	workerWaitGroup.Add(config.WorkerCount)

	for i := 0; i < config.WorkerCount; i++ {
		go func(id int) {
			Worker(id, tasksChannel, resultsChannel, doneChannel)
			workerWaitGroup.Done()

		}(i)
	}

	collectorDone := make(chan struct{})
	go func() {
		CollectResults(resultsChannel, doneChannel, metrics, metricsMutex)
		close(collectorDone)
	}()

	go func() {
		workerWaitGroup.Wait()
		close(resultsChannel)

	}()

	select {
	case <-collectorDone:
		fmt.Println("Scan Completed Successfully")

	case <-doneChannel:
		fmt.Println("Scan cancelled by user")
	}

	metricsMutex.Lock()
	metrics.EndTime = time.Now()
	metricsMutex.Unlock()

	server.Stop()
	metricsMutex.RLock()
	fmt.Printf("Files scanned: %d \n", metrics.FilesScanned)
	fmt.Printf("Duplicates: %d \n", countDuplicates(metrics))
	fmt.Printf("Total bytes: %d\n", metrics.TotalBytes)
	fmt.Printf("Errors: %d \n", len(metrics.Errors))
	metricsMutex.RUnlock()
}

func countDuplicates(metrics *ScanMetrics) int {
	count := 0
	for _, paths := range metrics.Duplicates {
		if len(paths) > 1 { // Only count actual duplicates
			count += len(paths) - 1 // Extra copies
		}
	}
	return count
}
