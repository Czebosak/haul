package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Library struct {
	Source string `toml:"source"`
	InstallCommand string `toml:"install_command"`
	Include []string `toml:"include"`
}

func (library Library) Name() string {
	return strings.TrimSuffix(filepath.Base(library.Source), ".git")
}

func (library Library) Installed(libraryDirPath string) bool {
	_, err := os.Stat(filepath.Join(libraryDirPath, library.Name()))
	return !errors.Is(err, os.ErrNotExist)
}

func (library Library) Install(libraryDirPath string) error {
	name := library.Name()

	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		gitPath, err := exec.LookPath("git")
		if err != nil {
			return err
		}

		command := exec.Command(gitPath, "-C", libraryDirPath, "clone", "--depth=1", library.Source)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		err = command.Run()

		if err != nil {
			return err
		}

		if library.InstallCommand != "" {
			split_command := strings.Split(library.InstallCommand, " ")
			executablePath, err := exec.LookPath(split_command[0])
			if err != nil {
				return err
			}

			exec.Command(executablePath, split_command[1:]...)
		}
	}

	return nil
}
