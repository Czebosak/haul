package main

import (
	"os"
	"path/filepath"
)

func getFilesFromDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			ext := filepath.Ext(entry.Name())
			if ext == ".cpp" || ext == ".c" || ext == ".o" {
				paths = append(paths, path + "/" + entry.Name())
			}
		}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := path + "/" + entry.Name()
			newPaths, err := getFilesFromDir(dirPath)
			if err != nil {
				return nil, err
			}

			paths = append(paths, newPaths...)
		}
	}

	return paths, nil
}
