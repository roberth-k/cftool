// +build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"log"
	"os"
	"path/filepath"
)

const progName = "cfn-tool"

var Default = Install

func Install() error {
	return sh.Run("go", "install", "-ldflags", ldflags)
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

var (
	ldflags string
)

func init() {
	gitVersion, err := sh.Output(
		"git", "describe",
		"--tags", "--always",
		"--match", "v*")

	if err != nil {
		log.Panic("git describe: %v", err)
	}

	gitVersion = gitVersion[1:] // Strip 'v' prefix.

	gitCommit, err := sh.Output("git", "rev-parse", "--short", "HEAD")

	if err != nil {
		log.Panic("git rev-parse: %v", err)
	}

	ldflags = fmt.Sprintf(
		"-X main.gitVersion=%s -X main.gitCommit=%s -X main.progName=%s",
		gitVersion, gitCommit, progName)
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

	return sh.RunWith(
		env, "go", "build",
		"-ldflags", ldflags,
		"-o", out)
}
