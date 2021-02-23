// +build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
)

var GoBin string

// Default target to run when none is specified
// If not set, running mage will list available targets
//goland:noinspection GoUnusedGlobalVariable
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

func ldflags() string {
	kubeBuilderVersion, _ := sh.Output("git", "describe", "--tags", "--dirty", "--broken")
	gitCommit, _ := sh.Output("git", "rev-parse", "HEAD")
	buildDate := time.Now().Format(time.RFC3339)

	return fmt.Sprintf(`-X main.kubeBuilderVersion=%s `+
		`-X main.goos=%s `+
		`-X main.goarch=%s `+
		`-X main.gitCommit=%s `+
		`-X main.buildDate=%s`,
		kubeBuilderVersion,
		runtime.GOOS,
		runtime.GOARCH,
		gitCommit,
		buildDate)
}

func programName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// Build the project locally
func Build() error {
	mg.Deps(Clean)
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", programName("bin/kubebuilder"), "-ldflags="+ldflags(), "./cmd")
	return cmd.Run()
}

type Linter mg.Namespace

// Download and install golangci-lint
func (Linter) Install() error {
	fmt.Println("Installing golangci-lint...")
	return nil
}

// Run golangci-lint linter
func (Linter) Lint() error {
	mg.Deps(Linter.Install)
	fmt.Println("Linting using golangci-lint...")
	return nil
}

// Build and install the binary with the current source code. Use it to test your changes locally.
//goland:noinspection GoUnusedExportedFunction
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	src, _ := filepath.Abs(programName("bin/kubebuilder"))
	dst, _ := filepath.Abs(path.Join(GoBin, programName("kubebuilder")))
	return sh.Copy(dst, src)
}

// Remove output and intermediate files
func Clean() {
	fmt.Println("Cleaning...")
	_ = os.RemoveAll("bin")
}

//goland:noinspection GoUnusedExportedType
type Test mg.Namespace

// Run the unit tests
func (Test) Unit() error {
	results, err := sh.Output("go", "test", "-race", "-v", "./pkg/...")
	fmt.Println(results)
	return err
}

func rmGlob(s string) error {
	files, err := filepath.Glob(s)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

// Run unit tests creating the output to report coverage
func (Test) Coverage() error {
	if err := rmGlob("*.out"); err != nil {
		return err
	}

	results, err := sh.Output("go", "test", "-race", "-failfast", "-tags=integration", "-coverprofile=coverage-all.out", `-coverpkg="./pkg/cli/...,./pkg/config/...,./pkg/internal/...,./pkg/model/...,./pkg/plugin/...,./pkg/plugins/golang,./pkg/plugins/internal/...`, "./pkg/...")
	fmt.Println(results)
	return err
}

// Run the integration tests
func (Test) Integration() error {
	if runtime.GOOS == "windows" {
		return errors.New("integration tests are not available on windows yet")
	}

	results, err := sh.Output("./test/integration.sh")
	if err != nil {
		return err
	}

	fmt.Println(results)
	return nil
}

// Run the end-to-end tests (used in the CI)
func (Test) E2e() error {
	if runtime.GOOS == "windows" {
		return errors.New("end to end tests are not available on windows yet")
	}

	results, err := sh.Output("./test/e2e/ci.sh")
	if err != nil {
		return err
	}

	fmt.Println(results)
	return nil

}
