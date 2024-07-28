package main

import (
	"bytes"
	"crypto/sha1"
	_ "embed"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed completion/completion.bash
var bashcomp []byte

//go:embed completion/completion.zsh
var zshcomp []byte

func installComp() error {
	if os.Geteuid() != 0 {
		return errors.New("--install-completions must be run as root.")
	}
	for _, fn := range []func() error{installBashComp, installZshComp} {
		if err := fn(); err != nil {
			return err
		}
	}
	fmt.Println(`BASH: Install your system's bash-completion package,`,
		`then run "exec bash".`)
	fmt.Println(`ZSH: Run "rm -f ~/.zcompdump && compinit".`,
		`(If compinit fails with a "command not found" error, add`,
		`"autoload -Uz compinit && compinit" to your ~/.zshrc,`,
		`then run "exec zsh".)`)
	return nil
}

func installBashComp() error {
	var path string
	switch runtime.GOOS {
	case "darwin":
		path = "/usr/local/share/bash-completion/completions/op"
	default:
		path = "/usr/share/bash-completion/completions/op"
	}
	changed, err := installFile(bashcomp, path)
	if err != nil {
		return fmt.Errorf("Error installing bash completion script: %v\n", err)
	}
	if changed {
		fmt.Println("Bash completion script updated.")
	}
	return nil
}

func installZshComp() error {
	changed, err := installFile(zshcomp,
		"/usr/local/share/zsh/site-functions/_op")
	if err != nil {
		return fmt.Errorf("Error installing zsh completion script: %v\n", err)
	}
	if changed {
		fmt.Println("Zsh completion script updated.")
	}
	return nil
}

func installFile(content []byte, path string) (bool, error) {
	var fileHash, contentHash hash.Hash
	file, err := os.Open(path)
	if err != nil {
		goto install
	}
	defer file.Close()

	fileHash = sha1.New()
	_, err = io.Copy(fileHash, file)
	if err != nil {
		goto install
	}

	contentHash = sha1.New()
	_, err = io.Copy(contentHash, bytes.NewReader(content))
	if err != nil {
		goto install
	}

	if bytes.Equal(fileHash.Sum(nil), contentHash.Sum(nil)) {
		return false, nil
	}

install:
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, err
	}
	return true, os.WriteFile(path, content, 0644)
}
