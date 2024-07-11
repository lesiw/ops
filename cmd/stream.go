package cmd

import (
	"cmp"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"

	"lesiw.io/ci/stream"
)

func Stream(args ...string) stream.Stream {
	return defaultNS.Stream(args...)
}

func StreamContext(ctx context.Context, args ...string) stream.Stream {
	return defaultNS.StreamContext(ctx, args...)
}

type cmdStream struct {
	ctx context.Context
	cmd *exec.Cmd
	log []byte
	env map[string]string

	initonce sync.Once
	waitonce sync.Once

	writer io.WriteCloser
	reader io.ReadCloser
	logger io.ReadCloser
}

func (s *cmdStream) init() {
	s.cmd.Env = os.Environ()
	for k, v := range s.env {
		s.cmd.Env = append(s.cmd.Env, k+"="+v)
	}
	if s.cmd.Stdin == nil {
		s.writer = must1(s.cmd.StdinPipe())
	}
	if s.cmd.Stdout == nil {
		s.reader = must1(s.cmd.StdoutPipe())
	}
	if s.cmd.Stderr == nil {
		s.logger = must1(s.cmd.StderrPipe())
	}
	must(stream.NewError(s.cmd.Start(), s))
}

func (s *cmdStream) Write(bytes []byte) (int, error) {
	s.initonce.Do(s.init)
	if s.writer == nil {
		return 0, nil
	}
	return s.writer.Write(bytes)
}

func (s *cmdStream) Close() error {
	s.initonce.Do(s.init)
	if s.writer == nil {
		return nil
	}
	return s.writer.Close()
}

func (s *cmdStream) Read(bytes []byte) (int, error) {
	s.initonce.Do(s.init)
	ch := make(chan ioret)
	var n int
	var err error
	if s.reader == nil {
		goto nilreader
	}

	go func() {
		n, err := s.reader.Read(bytes)
		ch <- ioret{n, err}
	}()
	select {
	case <-s.ctx.Done():
		n = 0
		err = io.EOF
	case ret := <-ch:
		n = ret.n
		err = ret.err
	}

nilreader:
	if err != nil || n == 0 {
		s.waitonce.Do(func() {
			if err1 := s.wait(); err1 != nil {
				err = err1
			}
		})
	}
	return n, err
}

func (s *cmdStream) wait() error {
	var log []byte
	if s.logger != nil {
		log = must1(io.ReadAll(s.logger))
	}
	err := s.cmd.Wait() // Closes pipes.
	if err == nil {
		return nil
	}
	ee := new(exec.ExitError)
	se := stream.NewError(err, s).(*stream.Error)
	if errors.As(err, &ee) {
		se.Code = ee.ExitCode()
	}
	se.Log = strings.TrimRight(string(log), "\n")
	return se
}

func (s *cmdStream) String() string {
	ret := new(strings.Builder)
	for _, k := range sortkeys(s.env) {
		ret.WriteString(k + "=" + s.env[k] + " ")
	}
	ret.WriteString(shJoin(s.cmd.Args))
	return ret.String()
}

type ioret struct {
	n   int
	err error
}

func sortkeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	var i int
	for k := range m {
		keys[i] = k
		i++
	}
	slices.Sort(keys)
	return keys
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func must1[T any](r0 T, err error) T {
	if err != nil {
		panic(err)
	}
	return r0
}
