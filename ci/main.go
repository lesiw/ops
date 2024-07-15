package main

import (
	"os"

	"lesiw.io/ci"
	"lesiw.io/cmdio"
	"lesiw.io/cmdio/cmd"
)

type target struct {
	goos   string
	goarch string
	unames string
	unamer string
}

var targets = []target{
	{"linux", "386", "linux", "i386"},
	{"linux", "amd64", "linux", "x86_64"},
	{"linux", "arm", "linux", "armv7l"},
	{"linux", "arm64", "linux", "aarch64"},
	{"darwin", "amd64", "darwin", "x86_64"},
	{"darwin", "arm64", "darwin", "arm64"},
	{"windows", "386", "", ""},
	{"windows", "arm", "", ""},
	{"windows", "amd64", "", ""},
	{"plan9", "386", "", ""},
	{"plan9", "arm", "", ""},
	{"plan9", "amd64", "", ""},
}

type actions struct{}

var name = "ci"

func main() {
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "build")
	}
	ci.Handle(actions{})
}

func (a actions) Build() {
	a.Clean()
	a.Lint()
	a.Test()
	a.Race()
	for _, t := range targets {
		box := cmd.Env(map[string]string{
			"CGO_ENABLED": "0",
			"GOOS":        t.goos,
			"GOARCH":      t.goarch,
		})
		box.MustRun("go", "build", "-o", "/dev/null")
		if t.unames != "" && t.unamer != "" {
			box.MustRun("go", "build", "-ldflags=-s -w", "-o",
				"out/"+name+"-"+t.unames+"-"+t.unamer,
				"./cmd/ci",
			)
		}
	}
}

func (a actions) Clean() {
	cmd.MustRun("rm", "-rf", "out")
	cmd.MustRun("mkdir", "out")
}

func (a actions) Lint() {
	ensureGolangci()
	cmd.MustRun("golangci-lint", "run")

	cmd.MustRun("go", "run", "github.com/bobg/mingo/cmd/mingo@latest", "-check")
}

func ensureGolangci() {
	if _, err := cmd.Get("which", "golangci-lint"); err == nil {
		return
	}
	gopath := cmd.MustGet("go", "env", "GOPATH")
	cmdio.MustPipe(
		cmd.Command("curl", "-sSfL",
			"https://raw.githubusercontent.com/golangci"+
				"/golangci-lint/master/install.sh"),
		cmd.Command("sh", "-s", "--", "-b", gopath.Output+"/bin"),
	)
}

func (a actions) Test() {
	ensureGoTestSum()
	cmd.MustRun("gotestsum", "./...")
}

func ensureGoTestSum() {
	if _, err := cmd.Get("which", "gotestsum"); err == nil {
		return
	}
	cmd.MustRun("go", "install", "gotest.tools/gotestsum@latest")
}

func (a actions) Race() {
	cmd.MustRun("go", "build", "-race", "-o", "/dev/null")
}

func (a actions) Bump() {
	versionfile := "cmd/ci/version.txt"
	bump := cmdio.MustGetPipe(
		cmd.Command("curl", "lesiw.io/bump"),
		cmd.Command("sh"),
	).Output
	version := cmdio.MustGetPipe(
		cmd.Command("cat", versionfile),
		cmd.Command(bump, "-s", "1"),
		cmd.Command("tee", versionfile),
	).Output
	cmd.MustRun("git", "add", versionfile)
	cmd.MustRun("git", "commit", "-m", version)
	cmd.MustRun("git", "tag", version)
	cmd.MustRun("git", "push")
	cmd.MustRun("git", "push", "--tags")
}
