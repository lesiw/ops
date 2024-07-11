package ci

import (
	"errors"
	"fmt"
	"os"

	"lesiw.io/ci/stream"
)

func Handler() {
	if r := recover(); r != nil {
		err, ok := r.(error)
		if !ok {
			panic(r)
		}
		if se := new(stream.Error); errors.As(err, &se) {
			fmt.Fprintf(os.Stderr,
				"exec failed: %v: %s\n", se.Stream, se.Error())
			if se.Log != "" {
				fmt.Fprintf(os.Stderr, "\nstderr:\n---\n%s\n---\n", se.Log)
			}
		} else {
			panic(err)
		}
		os.Exit(1)
	}
}
