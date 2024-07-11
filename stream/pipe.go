package stream

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// RunPipe chains streams together and panics on failure.
func RunPipe(src io.Reader, streams ...io.ReadWriter) {
	all := make([]any, len(streams)+1)
	all[0] = src
	for i, s := range streams {
		all[i+1] = s
	}
	printStreams(all...)
	if _, err := Copy(nopCloser{os.Stdout}, src, streams...); err != nil {
		panic(err)
	}
}

// CheckPipe chains streams together. It does not panic on failure.
// On failure, the failing stream's information will be returned.
func CheckPipe(src io.Reader, streams ...io.ReadWriter) *Result {
	all := make([]any, len(streams)+1)
	all[0] = src
	for i, s := range streams {
		all[i+1] = s
	}
	printStreams(all...)
	r := &Result{}
	if _, err := Copy(io.Discard, src, streams...); err != nil {
		if se := new(Error); errors.As(err, &se) {
			if se.Code > 0 {
				r.Stream = se.Stream
				r.Code = se.Code
				return r
			}
			panic(se)
		} else if ce := new(CopyError); errors.As(err, &ce) {
			if s, ok := ce.Reader.(Stream); ok {
				panic(NewError(err, s))
			} else {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	r.Ok = true
	return r
}

// GetPipe chains streams together and captures the output.
func GetPipe(src io.Reader, streams ...io.ReadWriter) *Result {
	all := make([]any, len(streams)+1)
	all[0] = src
	for i, s := range streams {
		all[i+1] = s
	}
	printStreams(all...)
	dst := new(bytes.Buffer)
	if _, err := Copy(dst, src, streams...); err != nil {
		panic(err)
	}
	r := &Result{
		Ok:     true,
		Output: strings.Trim(dst.String(), "\n"),
	}
	if s, ok := streams[len(streams)-1].(Stream); ok {
		r.Stream = s
	}
	return r
}

func printStreams(a ...any) {
	fmt.Fprintf(Trace, "+ ")
	for i, e := range a {
		if i > 0 {
			fmt.Fprintf(Trace, " | ")
		}
		if str, ok := e.(fmt.Stringer); ok {
			fmt.Fprintf(Trace, str.String())
		} else {
			fmt.Fprintf(Trace, "<stream>")
		}
	}
	fmt.Fprintln(Trace)
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }
