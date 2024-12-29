package main

import (
	"os"
	"github.com/pelletier/go-toml/v2"
)

type (
	Config struct {
		Name string
		Build BuildConfig
		Run RunConfig
		LibraryPath string
		Libraries []Library
	}

	BuildConfig struct {
		SourcePath string
		BuildPath string
		Compiler string
		Linker string
        IncludeSourceDirectory bool
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

func default_string(a *string, b string) {
	if *a == "" {
		*a = b
	}
}

func loadConfig() Config {
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		panic(err)
	}

	var config Config

	err = toml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}

	default_string(&config.LibraryPath, "external")
	default_string(&config.Build.SourcePath, "src")
	default_string(&config.Build.BuildPath, "build")
	
	return config
}
