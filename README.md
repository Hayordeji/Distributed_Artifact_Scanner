# Distributed Artifact Scanner

A high-performance, concurrent file scanner built in Go that detects duplicate files across multiple directories using SHA-256 hashing.

## Features

- **Concurrent Processing** - Multiple workers process files in parallel
- **Duplicate Detection** - Finds files with identical content using SHA-256 hashing
- **Real-time Monitoring** - HTTP API for live progress tracking
- **Graceful Cancellation** - Stop scans mid-run without data loss
- **File Classification** - Counts files by type (.txt, .pdf, .jpg, etc.)
- **Export Results** - Save scan results to JSON
- **Thread-Safe** - Race-condition free using mutexes and channels
- **Configurable** - Adjust worker count, file size limits, and target directories

**Architecture:**
1. **File Discovery** - Walks directories recursively, finds all files
2. **Worker Pool** - N goroutines process files concurrently
3. **Hashing** - Computes SHA-256 hash for each file
4. **Collector** - Aggregates results, detects duplicates
5. **HTTP Server** - Exposes real-time progress via REST API

## Installation

### Prerequisites

- Go 1.19 or higher
- GCC (for race detector, optional)

### Build

```bash
# Clone the repository
git clone https://github.com/Hayordeji/Distributed_Artifact_Scanner.git

# Run
go run .
```

## Usage

### Basic Scan

```bash
# Scan current directory
go run .

# Scan specific directory
go run . -dir=/path/to/scan

# Scan with 8 workers
go run . -dir=/path/to/scan -workers=8

# Set maximum file size (in bytes)
go run . -dir=/path/to/scan -max-size=104857600
```

### Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-dir` | `.` | Directory to scan |
| `-workers` | `4` | Number of concurrent workers |
| `-max-size` | `104857600` | Maximum file size to scan (bytes, default 100MB) |

## API Endpoints

The scanner starts an HTTP server on `http://localhost:8080` with the following endpoints:

### `GET /status`

Returns current scan progress.

**Response:**
```json
{
  "files_scanned": 1523,
  "files_pending": 477,
  "total_bytes": 45231891,
  "errors_count": 3,
  "running": true
}
```

### `GET /metrics`

Returns full scan statistics including duplicates.

**Response:**
```json
{
  "total_files": 2000,
  "total_bytes": 52341678,
  "files_scanned": 2000,
  "unique_files": 1847,
  "duplicate_groups": 12,
  "duplicate_files": 153,
  "duplicates": {
    "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3": [
      "/path/to/file1.txt",
      "/path/to/file2.txt"
    ]
  },
  "type_count": {
    ".txt": 1245,
    ".pdf": 532,
    ".jpg": 223
  },
  "errors": [],
  "start_time": "2026-02-24T10:00:00Z",
  "end_time": "2026-02-24T10:00:12Z",
  "duration": "12.345s"
}
```

### `POST /cancel`

Gracefully stops the current scan.

**Response:**
```
Scan cancellation initiated
```

## Testing

```bash
# Run all tests
go test -v

# Run with race detector (requires CGO)
go test -race -v

# Run specific test
go test -v -run TestProcessFile_Success

# Generate coverage report
go test -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**⭐ If you found this useful, consider starring the repo!**
