package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Response struct {
	FilesScanned int   `json:"files_scanned"`
	FilesPending int   `json:"files_pending"`
	TotalBytes   int64 `json:"total_bytes"`
	ErrorsCount  int   `json:"errors_count"`
	Running      bool  `json:"running"`
}

type Server struct {
	metrics      *ScanMetrics
	metricsMutex *sync.RWMutex
	doneChannel  chan struct{}
	httpServer   *http.Server
}

func NewServer(metrics *ScanMetrics, doneChannel chan struct{}, metricsMutex *sync.RWMutex) *Server {
	//CREATE NEW SERVER OBJECT
	server := &Server{
		metrics:      metrics,
		doneChannel:  doneChannel,
		metricsMutex: metricsMutex,
	}

	mux := http.NewServeMux()

	//REGISTER ENDPOINTS
	mux.HandleFunc("/status", server.handleStatus)
	mux.HandleFunc("/metrics", server.handleMetrics)
	mux.HandleFunc("/cancel", server.handleCancel)

	//ADD HTTP SERVER INSTANCE
	server.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return server
}

func (s *Server) Start() {
	go func() {
		fmt.Println("HTTP server starting on :8080")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		fmt.Printf("HTTP server shutdown error: %v\n", err)
	} else {
		fmt.Println("HTTP server stopped gracefully")
	}

}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//MUTEX LOCK
	s.metricsMutex.RLock()
	response := Response{
		FilesScanned: s.metrics.FilesScanned,
		FilesPending: s.metrics.FilesPending,
		TotalBytes:   s.metrics.TotalBytes,
		ErrorsCount:  len(s.metrics.Errors),
		Running:      s.metrics.EndTime.IsZero(),
	}
	s.metricsMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.metricsMutex.RLock()

	//CREATE A NEW METRICS
	metricsCopy := ScanMetrics{
		TotalFiles:   s.metrics.TotalFiles,
		TotalBytes:   s.metrics.TotalBytes,
		FilesScanned: s.metrics.FilesScanned,
		FilesPending: s.metrics.FilesPending,
		StartTime:    s.metrics.StartTime,
		EndTime:      s.metrics.EndTime,
		Duplicates:   make(map[string][]string, len(s.metrics.Duplicates)),
		TypeCount:    make(map[string]int, len(s.metrics.TypeCount)),

		Errors: make([]string, len(s.metrics.Errors)),
	}

	//COPY ALL METRICS FROM EXISTING SERVER METRICS TO NEW METRICS
	for hash, paths := range s.metrics.Duplicates {
		pathsCopy := make([]string, len(paths))
		copy(pathsCopy, paths)
		metricsCopy.Duplicates[hash] = pathsCopy
	}

	//COPY ALL TYPE COUNT FROM EXISTING SERVER TYPE COUNTS TO NEW TYPE COUNTS
	for ext, count := range s.metrics.TypeCount {
		metricsCopy.TypeCount[ext] = count
	}
	copy(metricsCopy.Errors, s.metrics.Errors)
	s.metricsMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metricsCopy)
}

// handleCancel triggers graceful scan cancellation
func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Try to send cancellation signal
	select {
	case s.doneChannel <- struct{}{}:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Scan cancellation initiated\n"))
	default:
		// Channel already closed or full
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Scan already stopped\n"))
	}
}
