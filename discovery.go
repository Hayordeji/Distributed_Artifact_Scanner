package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func DiscoverFiles(config ScanConfig, tasksChannel chan FileTask, doneChannel chan struct{}) {
	for _, dir := range config.Directories {
		//CHECK IF PATH IS A DIRECTORY
		dirInfo, err := os.Stat(dir)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		if !dirInfo.IsDir() {
			fmt.Printf("%s is not a directory\n", dir)
			continue
		}

		//WALK DIRECTORY
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {

				fmt.Printf("Error accessing %s: %v\n", path, err)
				return nil
			}

			//CHECK IF IT NOT A SYMLINK OR ANY KIND OF SPECIAL FILE
			if !info.Mode().IsRegular() {
				return nil
			}

			//SKIP IF FILE SIZE IS ABOVE MAX FILE SIZE
			if info.Size() > config.MaxFileSize {
				println("file too big")
				return nil
			}

			//ADD FILE TASK TO CHANNEL FOR PROCESSING
			task := FileTask{Path: path, Size: info.Size()}

			select {
			case tasksChannel <- task:

			case <-doneChannel:
				return nil
			}
			return nil
		})
	}
	close(tasksChannel)

}
