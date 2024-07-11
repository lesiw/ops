package ns

import (
	"context"

	"lesiw.io/ci/stream"
)

type NS struct {
	Namespace
}

func New(ns Namespace) *NS {
	return &NS{ns}
}

func (n *NS) Stream(args ...string) stream.Stream {
	return n.StreamContext(context.Background(), args...)
}

// Run runs a command and panics on failure.
func (n *NS) Run(args ...string) {
	if err := n.Try(args...); err != nil {
		panic(err)
	}
}

// Try runs a command and returns an error on failure.
func (n *NS) Try(args ...string) error {
	return n.RunContext(context.Background(), args...)
}

// Check runs a command, capturing its output.
// It will not panic regardless of exit status.
// If the command to run, however, it will panic.
func (n *NS) Check(args ...string) *stream.Result {
	return stream.Check(n.Stream(args...))
}

// Get runs a command and captures its output, panicking on failure.
func (n *NS) Get(args ...string) *stream.Result {
	return stream.Get(n.Stream(args...))
}
