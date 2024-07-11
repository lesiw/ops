package stream

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Trace is an [io.Writer] to which stream tracing information is written.
// To disable, set this variable to [io.Discard].
var Trace io.Writer = os.Stderr

type Stream interface {
	io.ReadWriteCloser
}

// Check executes a stream, capturing its output.
// It will not panic regardless of exit status.
// If the command to run, however, it will panic.
func Check(s Stream) *Result {
	fmt.Fprintln(Trace, "+", s)
	buf, err := io.ReadAll(s)
	r := &Result{
		Stream: s,
		Output: strings.Trim(string(buf), "\n"),
	}
	if err != nil {
		se := new(Error)
		if errors.As(err, &se) {
			if se.Code > 0 {
				r.Code = se.Code
				return r
			}
			panic(se)
		} else {
			panic(NewError(err, s))
		}
	}
	r.Ok = true
	return r
}

// Get executes a stream and captures its output, panicking on failure.
func Get(s Stream) *Result {
	fmt.Fprintln(Trace, "+", s)
	buf, err := io.ReadAll(s)
	if err != nil {
		panic(err)
	}
	r := &Result{
		Ok:     true,
		Stream: s,
		Output: strings.Trim(string(buf), "\n"),
	}
	return r
}
