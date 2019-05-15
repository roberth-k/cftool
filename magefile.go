// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"os"
	"path/filepath"
)

const progName = "cfn-tool"

var Default = Install

func Install() error {
	return sh.Run("go", "install")
}

type Build mg.Namespace

var targets = []struct {
	GOOS   string
	GOARCH string
}{
	{"windows", "amd64"},
	{"darwin", "amd64"},
	{"linux", "amd64"},
}

func (Build) All() error {
	for _, target := range targets {
		err := build(target.GOOS, target.GOARCH)

		if err != nil {
			return err
		}
	}

	return nil
}

func build(goos string, goarch string) error {
	env := map[string]string{
		"GOOS":   goos,
		"GOARCH": goarch,
	}

	dir := filepath.Join(".build", goos+"-"+goarch)
	err := os.MkdirAll(dir, os.ModeDir|0751)

	if err != nil {
		return err
	}

	out := filepath.Join(dir, progName)

	if goos == "windows" {
		out += ".exe"
	}

	return sh.RunWith(env, "go", "build", "-o", out)
}
