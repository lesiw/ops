package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"lesiw.io/ci/ns"
	"lesiw.io/ci/stream"
)

var defaultNS = ns.New(&cmdNS{})

type cmdNS struct {
	env map[string]string
}

func (ns *cmdNS) StreamContext(
	ctx context.Context, args ...string,
) stream.Stream {
	s := &cmdStream{
		ctx: ctx,
		cmd: exec.CommandContext(ctx, args[0], args[1:]...),
		env: ns.env,
	}
	return s
}

func (ns *cmdNS) RunContext(ctx context.Context, args ...string) error {
	s := ns.StreamContext(ctx, args...).(*cmdStream)
	s.cmd.Stdin = os.Stdin
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr
	s.Close() // init.
	fmt.Fprintln(stream.Trace, "+", s)
	_, err := s.Read(nil) // wait.
	return err
}

func (ns *cmdNS) Env(k string) string {
	if ns.env == nil {
		return ""
	}
	return ns.env[k]
}

func (ns *cmdNS) Setenv(k, v string) {
	if ns.env == nil {
		ns.env = make(map[string]string)
	}
	ns.env[k] = v
}

func Env(env map[string]string) *ns.NS {
	n := new(cmdNS)
	for k, v := range env {
		n.Setenv(k, v)
	}
	return ns.New(n)
}

func Run(args ...string) {
	defaultNS.Run(args...)
}

func Check(args ...string) *stream.Result {
	return defaultNS.Check(args...)
}

func Get(args ...string) *stream.Result {
	return defaultNS.Get(args...)
}
