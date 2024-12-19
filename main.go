package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/pelletier/go-toml"
)

type (
	Config struct {
		Build BuildConfig
		Run RunConfig
	}

	BuildConfig struct {
		SourcePath string
		BuildPath string
		Compiler string
		Linker string
	}

	RunConfig struct {
		Arguments string
	}

	Output struct {
		command string
		path string
		data string
	}
)

const ConfigPath = "haul.toml"

var configLoaded bool = false
var config Config

func loadConfig() {
	if configLoaded {
		return
	}
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		panic(err)
	}

	err = toml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	
	configLoaded = true
}

func getFilesFromDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			paths = append(paths, path + "/" + entry.Name())
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

func compileFile(compilerPath string, objectDirectoryPath string, path string) Output {
	name := filepath.Base((path))
	nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))
	objectFilePath := filepath.Join(objectDirectoryPath, nameWithoutExt) + ".o"

	command := exec.Command(compilerPath, "-c", path, "-o", objectFilePath)
	
	stdout, _ := command.Output()

	output := Output {
		command: command.String(),
		path: path,
		data: string(stdout),
	}

	return output
}

func linkFiles(linkerPath string, config Config, paths []string) Output {
	command_args := paths
	command_args = append(command_args, []string{"-o", filepath.Join(config.Build.BuildPath, "main")}...)
	command := exec.Command(linkerPath, command_args...)
	
	stdout, _ := command.Output()

	output := Output {
		command: command.String(),
		data: string(stdout),
	}

	return output
}

func build() {
	loadConfig()

	objectDirectoryPath := filepath.Join(config.Build.BuildPath, "obj")

	err := os.MkdirAll(objectDirectoryPath, os.ModePerm)
	if err != nil {
		entries, err := os.ReadDir(config.Build.BuildPath)
		if err != nil {
			panic(err)
		}

		for _, entry := range entries {
			os.RemoveAll(config.Build.BuildPath + entry.Name())
		}
	}

	paths, err := getFilesFromDir(config.Build.SourcePath)
	if err != nil {
		panic(err)
	}

	compilerPath, err := exec.LookPath(config.Build.Compiler)
	if err != nil {
		panic(err)
	}

	var outputs []Output

	for _, path := range paths {
		output := compileFile(compilerPath, objectDirectoryPath, path)
		outputs = append(outputs, output)
	}

	for _, output := range outputs {
		fmt.Println(output.command)
		fmt.Print(output.data)
	}

	paths, err = getFilesFromDir(objectDirectoryPath)
	if err != nil {
		panic(err)
	}
	
	linkerPath, err := exec.LookPath(config.Build.Linker)
	if err != nil {
		panic(err)
	}

	output := linkFiles(linkerPath, config, paths)
	fmt.Println(output.command)
	fmt.Print(output.data)
}

func run() {
	build()

	path := "build/main"

	command := exec.Command(path)
	stdout, _ := command.Output()

	output := Output {
		command: command.String(),
		path: path,
		data: string(stdout),
	}

	fmt.Println(output.command)
	fmt.Print(output.data)
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "build": build()
		case "run": run()
		}	
	}
}
