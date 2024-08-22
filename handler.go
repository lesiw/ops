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
	exit  = os.Exit
	flags = flag.NewSet(stderr, "op [-l] OPERATION")
	list  = flags.Bool("l,list", "list available ops and exit")

	posts []func()

	stderr io.Writer = os.Stderr
)

func Handle(a any) {
	var code int
	defer func() { exit(code) }()
	if err := opHandler(a, os.Args[1:]...); err != nil {
		if err.Error() != "" {
			fmt.Fprintln(stderr, err)
		}
		code = 1
	}
}

// PostHandle registers a func() to run after Handle has run successfully.
func PostHandle(post func()) {
	posts = append([]func(){post}, posts...)
}

func opHandler(a any, args ...string) (err error) {
	defer handleRecover(&err)
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
		return fmt.Errorf("bad op: no op provided")
	}
	val := reflect.ValueOf(a)
	var ran bool
	for _, method := range methodsByName(t, args[0]) {
		if method.Func.Type().In(0).Kind() == reflect.Ptr {
			if val.Kind() == reflect.Ptr {
				method.Func.Call([]reflect.Value{val})
			} else {
				ptr := reflect.New(val.Type())
				ptr.Elem().Set(val)
				method.Func.Call([]reflect.Value{ptr})
			}
		} else if val.Kind() == reflect.Ptr {
			method.Func.Call([]reflect.Value{reflect.ValueOf(a).Elem()})
		} else {
			method.Func.Call([]reflect.Value{reflect.ValueOf(a)})
		}
		ran = true
	}
	if !ran {
		return fmt.Errorf("bad op '%s'", args[0])
	}
	for _, post := range posts {
		post()
	}
	return nil
}

type errorPrinter interface {
	error
	Print(io.Writer)
}

func handleRecover(ret *error) {
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
		errp.Print(stderr)
	} else {
		fmt.Fprintln(stderr, err.Error())
	}
	if ret == nil || *ret == nil {
		*ret = errors.New("")
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
	var ptr reflect.Type
	if t.Kind() == reflect.Pointer {
		ptr = t
		t = t.Elem()
	} else {
		ptr = reflect.PointerTo(t)
	}
	for i := range ptr.NumMethod() {
		method := ptr.Method(i)
		methods[method.Name] = true
	}
	if t.Kind() != reflect.Struct {
		return
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
	var ptr reflect.Type
	if t.Kind() == reflect.Pointer {
		ptr = t
		t = t.Elem()
	} else {
		ptr = reflect.PointerTo(t)
	}
	for i := range ptr.NumMethod() {
		method := ptr.Method(i)
		if snakecase(method.Name) == name {
			methods = append(methods, method)
		}
	}
	if len(methods) == 0 && t.Kind() == reflect.Struct {
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
