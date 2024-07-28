package main

import (
	"os"
	"strings"

	"lesiw.io/cmdio"
	"lesiw.io/cmdio/sys"
	"lesiw.io/ops"
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

type Ops struct{}

var binname = "op"

func main() {
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "build")
	}
	ops.Handle(Ops{})
}

func (op Ops) Build() {
	op.Clean()
	op.Lint()
	op.Test()
	op.Race()
	for _, t := range targets {
		box := sys.Env(map[string]string{
			"CGO_ENABLED": "0",
			"GOOS":        t.goos,
			"GOARCH":      t.goarch,
		})
		box.MustRun("go", "build", "-o", "/dev/null")
		if t.unames != "" && t.unamer != "" {
			box.MustRun("go", "build", "-ldflags=-s -w", "-o",
				"out/"+binname+"-"+t.unames+"-"+t.unamer,
				"./op",
			)
		}
	}
}

func (op Ops) Clean() {
	sys.MustRun("rm", "-rf", "out")
	sys.MustRun("mkdir", "out")
}

func (op Ops) Lint() {
	ensureGolangci()
	sys.MustRun("golangci-lint", "run")

	sys.MustRun("go", "run", "github.com/bobg/mingo/cmd/mingo@latest", "-check")
}

func ensureGolangci() {
	if _, err := sys.Get("which", "golangci-lint"); err == nil {
		return
	}
	gopath := sys.MustGet("go", "env", "GOPATH")
	cmdio.MustPipe(
		sys.Command("curl", "-sSfL",
			"https://raw.githubusercontent.com/golangci"+
				"/golangci-lint/master/install.sh"),
		sys.Command("sh", "-s", "--", "-b", gopath.Output+"/bin"),
	)
}

func (op Ops) Test() {
	ensureGoTestSum()
	sys.MustRun("gotestsum", "./...")
}

func ensureGoTestSum() {
	if _, err := sys.Get("which", "gotestsum"); err == nil {
		return
	}
	sys.MustRun("go", "install", "gotest.tools/gotestsum@latest")
}

func (op Ops) Race() {
	sys.MustRun("go", "build", "-race", "-o", "/dev/null")
}

func (op Ops) BumpApp() {
	versionfile := "op/version.txt"
	bump := cmdio.MustGetPipe(
		sys.Command("curl", "lesiw.io/bump"),
		sys.Command("sh"),
	).Output
	curVersion := sys.MustGet("cat", versionfile).Output
	version := cmdio.MustGetPipe(
		strings.NewReader(curVersion+"\n"),
		sys.Command(bump, "-s", "1"),
		sys.Command("tee", versionfile),
	).Output
	sys.MustRun("git", "add", versionfile)
	sys.MustRun("git", "commit", "-m", version)
	sys.MustRun("git", "push")
}

func (op Ops) BumpLib() {
	bump := cmdio.MustGetPipe(
		sys.Command("curl", "lesiw.io/bump"),
		sys.Command("sh"),
	).Output
	version := cmdio.MustGetPipe(
		sys.Command("git", "describe", "--abbrev=0", "--tags"),
		sys.Command(bump, "-s", "1"),
	).Output
	sys.MustRun("git", "tag", version)
	sys.MustRun("git", "push")
	sys.MustRun("git", "push", "--tags")
}
