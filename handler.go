//go:build !tinygo
// +build !tinygo

package ops

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"

	"lesiw.io/flag"
)

var (
	flags = flag.NewSet(os.Stderr, "op [-l] OPERATION")
	list  = flags.Bool("l,list", "list available ops and exit")

	posts []func()
)

func Handle(a any) {
	defer handleRecover()
	if err := opHandler(a, os.Args[1:]...); err != nil {
		if err.Error() != "" {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

// PostHandle registers a func() to run after Handle has run successfully.
func PostHandle(post func()) {
	posts = append([]func(){post}, posts...)
}

func opHandler(a any, args ...string) error {
	if err := flags.Parse(args...); err != nil {
		return errors.New("")
	}
	t := reflect.TypeOf(a)
	if *list {
		for _, name := range methodNames(t) {
			fmt.Println(snakecase(name))
		}
		return nil
	}
	if len(flags.Args) < 1 {
		return fmt.Errorf("bad action: no action provided")
	}
	var ran bool
	for _, method := range methodsByName(t, args[0]) {
		method.Func.Call([]reflect.Value{reflect.ValueOf(a)})
		ran = true
	}
	if !ran {
		return fmt.Errorf("bad action '%s'", args[0])
	}
	for _, post := range posts {
		post()
	}
	os.Exit(0)
	return nil
}

type errorPrinter interface {
	error
	Print(io.Writer)
}

func handleRecover() {
	r := recover()
	if r == nil {
		return
	}
	err, ok := r.(error)
	if !ok {
		panic(r)
	}
	var errp errorPrinter
	if errors.As(err, &errp) {
		errp.Print(os.Stderr)
	} else {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

func methodNames(t reflect.Type) (methods []string) {
	for m := range methodNameSet(t) {
		methods = append(methods, m)
	}
	slices.Sort(methods)
	return
}

func methodNameSet(t reflect.Type) (methods map[string]bool) {
	methods = make(map[string]bool)
	for i := range t.NumMethod() {
		method := t.Method(i)
		methods[method.Name] = true
	}
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.Anonymous {
			continue
		}
		for name := range methodNameSet(field.Type) {
			methods[name] = true
		}
	}
	return
}

func methodsByName(t reflect.Type, name string) (methods []reflect.Method) {
	for i := range t.NumMethod() {
		method := t.Method(i)
		if snakecase(method.Name) == name {
			methods = append(methods, method)
		}
	}
	if len(methods) == 0 {
		for i := range t.NumField() {
			field := t.Field(i)
			if !field.Anonymous {
				continue
			}
			methods = append(methods, methodsByName(field.Type, name)...)
		}
	}
	return
}
