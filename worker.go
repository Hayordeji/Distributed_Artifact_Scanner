package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func WorkerProcessFiles(id int, taskChannel chan FileTask, resultsChannel chan ScanResult, doneChannel chan struct{}) {

	for {
		select {
		case task, ok := <-taskChannel:
			if !ok {
				fmt.Printf("Task channel closed.\n")
				return
			}
			result := ProcessFiles(task)
			select {
			case resultsChannel <- result:

			case <-doneChannel:
				return
			}

		case <-doneChannel:
			return
		}

	}
}

func ProcessFiles(task FileTask) ScanResult {
	result := ScanResult{
		Path: task.Path,
		Size: task.Size,
	}

	//GET FILE
	file, err := os.Open(task.Path)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	defer file.Close()

	//HASH FILE
	hasher := sha256.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		fmt.Printf("Error copying %s: %v\n", task.Path, err)
		result.Error = err.Error()
		return result
	}

	hashbyte := hasher.Sum(nil)
	hashstring := hex.EncodeToString(hashbyte)
	result.Hash = hashstring

	//GET EXTENSION AND RETURN
	extension := filepath.Ext(task.Path)
	result.FileType = extension
	return result

}
