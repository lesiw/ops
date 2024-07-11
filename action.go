//go:build !tinygo
// +build !tinygo

package ci

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"lesiw.io/flag"
)

var (
	flags = flag.NewSet(os.Stderr, "ci [-l] ACTION")
	list  = flags.Bool("l,list", "list available actions and exit")
)

func ActionHandler(a any, args ...string) {
	if err := actionHandler(a, args...); err != nil {
		if err.Error() != "" {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func actionHandler(a any, args ...string) error {
	if err := flags.Parse(args...); err != nil {
		return errors.New("")
	}
	t := reflect.TypeOf(a)
	if *list {
		for i := range t.NumMethod() {
			method := t.Method(i)
			fmt.Println(snakecase(method.Name))
		}
		return nil
	}
	if len(flags.Args) < 1 {
		return fmt.Errorf("bad action: no action provided")
	}
	for i := range t.NumMethod() {
		method := t.Method(i)
		if snakecase(method.Name) == args[0] {
			method.Func.Call([]reflect.Value{reflect.ValueOf(a)})
			os.Exit(0)
		}
	}
	return fmt.Errorf("bad action '%s'", args[0])
}
