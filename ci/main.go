package main

import (
	"os"

	"lesiw.io/ci"
	"lesiw.io/cmdio"
	"lesiw.io/cmdio/cmd"
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
	defer cmdio.Recover(os.Stderr)
	args := os.Args[1:]
	if len(args) == 0 {
		args = append(args, "build")
	}
	ci.ActionHandler(project, args...)
}

func (a *actions) Build() {
	a.Lint()
	for _, target := range targets {
		cmd.Env(map[string]string{
			"CGO_ENABLED": "0",
			"GOOS":        target[0],
			"GOARCH":      target[1],
		}).MustRun("go", "build", "-o", "/dev/null")
	}
}

func (a *actions) Lint() {
	cmd.MustRun("go", "build", "-race", "-o", "/dev/null")

	EnsureGolangci()
	cmd.MustRun("golangci-lint", "run")

	cmd.MustRun("go", "run", "github.com/bobg/mingo/cmd/mingo@latest", "-check")
}

func EnsureGolangci() {
	if cmd.MustCheck("which", "golangci-lint").Ok {
		return
	}
	gopath := cmd.MustGet("go", "env", "GOPATH")
	cmdio.MustRunPipe(
		cmd.Command("curl", "-sSfL",
			"https://raw.githubusercontent.com/golangci"+
				"/golangci-lint/master/install.sh"),
		cmd.Command("sh", "-s", "--", "-b", gopath.Output+"/bin"),
	)
}
