package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type (
	Config struct {
		Name string `toml:"name" comment:"Name of the library"`
		Build BuildConfig `toml:"build"`
		Run RunConfig `toml:"run"`
		LibraryPath string `toml:"library_path"`
		Libraries []Library `toml:"libraries"`
	}

	BuildConfig struct {
		SourcePath string `toml:"source_path"`
		BuildPath string `toml:"build_path"`
		Compiler string `toml:"compiler"`
		Linker string `toml:"linker"`
        IncludeSourceDirectory bool `toml:"include_source_directory"`
	}

	RunConfig struct {
		Arguments string `toml:"arguments"`
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

var defaultConfig = Config {
	Name: "example-project",
	Build: BuildConfig {
		Compiler: "g++",
		Linker: "g++",
	},
}

const ConfigPath = "haul.toml"

func LoadConfig() Config {
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

func getDefaultConfigPath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userConfigDir, "haul", "default.toml"), nil
}

func createDefaultConfig() error {
	configPath, err := getDefaultConfigPath()
	if err != nil {
		return err
	}

	data, err := toml.Marshal(defaultConfig)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(configPath)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	f.Close()

	return err
}

func GetOrCreateDefaultConfig() ([]byte, error) {
	configPath, err := getDefaultConfigPath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = createDefaultConfig()
			if err != nil {
				return nil, err
			}

			f, err = os.Open(configPath)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}
