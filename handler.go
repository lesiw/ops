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
	"strings"

	"lesiw.io/defers"
	"lesiw.io/flag"
)

var (
	exit  = defers.Exit
	flags = flag.NewSet(stderr, "op [-l] OPERATION")
	list  = flags.Bool("l,list", "list available ops and exit")

	afters []func()

	stderr io.Writer = os.Stderr
)

func Handle(a any) {
	var code int
	defer func() { exit(code) }()
	defer defers.Recover()
	if err := opHandler(a, os.Args[1:]...); err != nil {
		if err.Error() != "" {
			fmt.Fprintln(stderr, err)
		}
		code = 1
	}
}

// After registers a func() to run after Handle has run successfully.
func After(f func()) {
	afters = append([]func(){f}, afters...)
}

// Defer registers a func() to run at the end of the program.
func Defer(f func()) {
	defers.Add(f)
}

func opHandler(a any, args ...string) (err error) {
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
	for _, method := range methods(t) {
		if err := validate(method); err != nil {
			return err
		}
	}
	val := reflect.ValueOf(a)
	var ran bool
	for _, method := range methodsByName(t, args[0]) {
		if err := exec(val, method); err != nil {
			return err
		}
		ran = true
	}
	if !ran {
		return fmt.Errorf("bad op '%s'", args[0])
	}
	for _, after := range afters {
		after()
	}
	return nil
}

func validate(fn reflect.Method) error {
	typ := fn.Type
	if typ.NumOut() == 0 {
		return nil
	}
	if typ.NumOut() > 1 || typ.Out(0).Name() != "error" {
		return fmt.Errorf("bad op: bad signature: %s", signature(fn))
	}
	return nil
}

func exec(val reflect.Value, fn reflect.Method) (err error) {
	var catch bool
	defer func() {
		if !catch {
			return
		}
		r := recover()
		if r == nil {
			return
		}
		switch v := r.(type) {
		case fmt.Stringer:
			err = errors.New(v.String())
		case error:
			err = errors.New(v.Error())
		default:
			err = fmt.Errorf("%v", v)
		}
	}()

	typ := fn.Type
	if typ.NumOut() == 0 {
		catch = true
	}
	var ret []reflect.Value
	if typ.In(0).Kind() == reflect.Ptr {
		if val.Kind() == reflect.Ptr {
			ret = call(val, fn)
		} else {
			ptr := reflect.New(val.Type())
			ptr.Elem().Set(val)
			ret = call(ptr, fn)
		}
	} else if val.Kind() == reflect.Ptr {
		ret = call(val.Elem(), fn)
	} else {
		ret = call(val, fn)
	}
	if len(ret) > 0 {
		if errv := ret[0]; !errv.IsNil() {
			return errv.Interface().(error)
		}
	}
	return
}

func call(rcvr reflect.Value, fn reflect.Method) []reflect.Value {
	t := fn.Type
	in := make([]reflect.Value, t.NumIn())
	for i := range t.NumIn() {
		if i == 0 {
			in[i] = rcvr
		} else {
			in[i] = reflect.Zero(t.In(i))
		}
	}
	return fn.Func.Call(in)
}

func methods(t reflect.Type) (methods []reflect.Method) {
	var ptr reflect.Type
	if t.Kind() == reflect.Pointer {
		ptr = t
	} else {
		ptr = reflect.PointerTo(t)
	}
	for i := range ptr.NumMethod() {
		method := ptr.Method(i)
		methods = append(methods, method)
	}
	return
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

func signature(m reflect.Method) string {
	t := m.Type
	var buf strings.Builder
	buf.WriteString("func (")
	for i := 0; i < t.NumIn(); i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(t.In(i).String())
	}
	buf.WriteString(")")
	if t.NumOut() > 0 {
		if t.NumOut() > 1 {
			buf.WriteString(" (")
		} else {
			buf.WriteString(" ")
		}
		for i := 0; i < t.NumOut(); i++ {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(t.Out(i).String())
		}
		if t.NumOut() > 1 {
			buf.WriteString(")")
		}
	}
	return buf.String()
}
