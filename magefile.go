// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
)

var GoBin string

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build

func buildGoBin() string {
	if goBin := os.Getenv("GOBIN"); goBin != "" {
		return goBin
	}

	if goPath := os.Getenv("GOPATH"); goPath != "" {
		return path.Join(goPath, "bin")
	}

	panic("GOBIN or GOPATH environment variables should be set")
}

func init() {
	GoBin = buildGoBin()
}

func programName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	mg.Deps(Clean)
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", programName("bin/kubebuilder"), "./cmd")
	return cmd.Run()
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	src, _ := filepath.Abs(programName("bin/kubebuilder"))
	dst, _ := filepath.Abs(path.Join(GoBin, programName("kubebuilder")))
	fmt.Printf("src: %s, dst: %s", src, dst)
	return sh.Copy(dst, src)
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("bin")
}
