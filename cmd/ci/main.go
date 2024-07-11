package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"lesiw.io/cmdio"
	"lesiw.io/cmdio/cmd"
	"lesiw.io/defers"
	"lesiw.io/flag"
)

var (
	flags    = flag.NewSet(os.Stderr, "ci COMMAND")
	_        = flags.Bool("l", "list commands")
	install  = flags.Bool("install-completions", "install completion scripts")
	printver = flags.Bool("V,version", "print version")

	//go:embed version.txt
	versionfile string
	version     string
)

func init() {
	version = strings.TrimRight(versionfile, "\n")
}

func main() {
	cmdio.Trace = io.Discard
	if err := run(); err != nil {
		if err.Error() != "" {
			fmt.Fprintln(os.Stderr, err)
		}
		defers.Exit(1)
	}
	defers.Exit(0)
}

func run() error {
	if err := flags.Parse(os.Args[1:]...); err != nil {
		return errors.New("")
	}
	if *printver {
		fmt.Println(version)
		return nil
	} else if *install {
		return installComp()
	}
	if err := changeToGitRoot(); err != nil {
		return err
	}
	workdir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	if stat, err := os.Stat("ci"); os.IsNotExist(err) || !stat.IsDir() {
		return fmt.Errorf("no 'ci' directory found: %w", err)
	}
	bindir, err := os.MkdirTemp("", "ci")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defers.Add(func() { _ = os.RemoveAll(bindir) })
	cidir := filepath.Join(workdir, "ci")
	if err := os.Chdir(cidir); err != nil {
		return fmt.Errorf("failed to chdir to '%s': %w", cidir, err)
	}
	cibin := filepath.Join(bindir, "ci")
	if err := cmd.Run("go", "build", "-o", cibin, "."); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	if err := os.Chdir(workdir); err != nil {
		return fmt.Errorf("failed to chdir to '%s': %w", workdir, err)
	}
	if err := cmd.Run(append([]string{cibin}, os.Args[1:]...)...); err != nil {
		return errors.New("")
	}
	return nil
}

func changeToGitRoot() error {
	for {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		fileinfo, err := os.Stat(".git")
		if err == nil && fileinfo.IsDir() {
			return nil
		}
		reachedRoot := (cwd == "/" || cwd == (filepath.VolumeName(cwd)+"\\"))
		if reachedRoot || os.Chdir("..") != nil {
			return fmt.Errorf("No .git directory was found.")
		}
	}
}
