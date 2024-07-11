package main

import (
	"os"

	"lesiw.io/ci"
	"lesiw.io/ci/cmd"
)

var targets = [][]string{
	{"linux", "386"},
	{"linux", "amd64"},
	{"linux", "arm"},
	{"linux", "arm64"},
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"windows", "386"},
	{"windows", "arm"},
	{"windows", "amd64"},
	{"plan9", "386"},
	{"plan9", "arm"},
	{"plan9", "amd64"},
}

type actions struct{}

var project = new(actions)

func main() {
	os.Setenv("CGO_ENABLED", "0")

	defer ci.Handler()
	ci.ActionHandler(project, os.Args[1:]...)
}

func (a *actions) Build() {
	for _, target := range targets {
		cmd.Env(map[string]string{
			"GOOS":   target[0],
			"GOARCH": target[1],
		}).Run("go", "build", "-o", "/dev/null")
	}
}
