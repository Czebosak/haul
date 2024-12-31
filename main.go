package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/pelletier/go-toml/v2"
)

const Prefix = "[Haul] "
var PrefixColor = color.New(color.FgGreen, color.Bold)

func PrefixPrint(s string, a ...any) {
	PrefixColor.Printf(Prefix + s + "\n", a...)
}

func compileFile(compilerPath string, includes []string, includeSourceDirectory bool, sourceDirectory string, objectDirectoryPath string, path string) Output {
	name := filepath.Base((path))
	nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))
	objectFilePath := filepath.Join(objectDirectoryPath, nameWithoutExt) + ".o"

	commandArgs := []string{"-c", path, "-o", objectFilePath}

    if includeSourceDirectory {
        commandArgs = append(commandArgs, "-I" + sourceDirectory)
    }
	
	for _, include := range includes {
		commandArgs = append(commandArgs, "-I" + include)
	}

	command := exec.Command(compilerPath, commandArgs...)
	
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 0 {
				PrefixPrint("Failed to compile file %s", path)
				os.Exit(1)
			}
		} else {
			panic(err)
		}
	}

	output := Output {
		command: command.String(),
		path: path,
	}

	return output
}

func linkFiles(linkerPath string, includes []string, config Config, paths []string) Output {
	commandArgs := paths
	commandArgs = append(commandArgs, []string{"-o", filepath.Join(config.Build.BuildPath, config.Name)}...)

	for _, include := range includes {
		commandArgs = append(commandArgs, "-L" + include)
	}

	for _, library := range config.Libraries {
		commandArgs = append(commandArgs, "-l" + library.Name())
	}

	command := exec.Command(linkerPath, commandArgs...)
	fmt.Println(command.String())
	
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 0 {
				PrefixPrint("Failed to link files")
				os.Exit(1)
			}
		} else {
			panic(err)
		}
	}

	output := Output {
		command: command.String(),
	}

	return output
}

func installLibraries(config Config) {
	if len(config.Libraries) == 0 {
		return
	}

	PrefixPrint("Checking libraries")

	err := os.Mkdir(config.LibraryPath, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}

	for _, library := range config.Libraries {
		if !library.Installed(config.LibraryPath) {
			PrefixPrint("Installing %s", library.Name())
			err := library.Install(config.LibraryPath)
			if err != nil {
				panic(err)
			}
		}
	}
}

func build() {
	config := LoadConfig()

	installLibraries(config)

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

	var includes []string
	for _, library := range config.Libraries {
		if library.Include != nil {
			for _, include := range library.Include {
				includes = append(includes, filepath.Join(config.LibraryPath, library.Name(), include))
			}
		} else {
			includes = append(includes, filepath.Join(config.LibraryPath, library.Name()))
		}
	}

	var outputs []Output

	for _, path := range paths {
		output := compileFile(compilerPath, includes, config.Build.IncludeSourceDirectory, config.Build.SourcePath, objectDirectoryPath, path)
		outputs = append(outputs, output)
	}

	for _, output := range outputs {
		fmt.Println(output.command)
	}

	paths, err = getFilesFromDir(objectDirectoryPath)
	if err != nil {
		panic(err)
	}
	
	linkerPath, err := exec.LookPath(config.Build.Linker)
	if err != nil {
		panic(err)
	}

	output := linkFiles(linkerPath, includes, config, paths)
	fmt.Println(output.command)
	fmt.Print(output.data)
}

func run() {
	build()

	config := LoadConfig()

	command := exec.Command(filepath.Join(config.Build.BuildPath, config.Name))
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		panic(err)
	}

	output := Output {
		command: command.String(),
	}

	fmt.Println(output.command)
	fmt.Print(output.data)
}

func init_command() {
	if len(os.Args) < 3 {
		color.Red("Provide the name of the project.")
		os.Exit(1)
	}

	data, err := GetOrCreateDefaultConfig()
	if err != nil {
		panic(err)
	}

	var config Config
	toml.Unmarshal(data, &config)

	config.Name = os.Args[2]

	data, err = toml.Marshal(config)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(ConfigPath)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(data)
	if err != nil {
		panic(err)
	}

	f.Close()
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "build": build()
		case "run": run()
		case "init": init_command()
		}
	}
}
