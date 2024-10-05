package ops

import (
	"bytes"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type FuncHandler func() error

func (h FuncHandler) Panic()       { _ = h() }
func (h FuncHandler) Error() error { return h() }

func TestHandleSuccess(t *testing.T) {
	var code *int
	n := new(int)
	swap(t, &exit, func(exitcode int) { code = &exitcode })
	swap(t, &os.Args, []string{"", "error"})
	errbuf := new(bytes.Buffer)
	swap[io.Writer](t, &stderr, errbuf)
	handler := FuncHandler(func() error { *n++; return nil })

	Handle(handler)

	if got, want := *n, 1; got != want {
		t.Errorf("got %d FuncHandler.Error() calls, want %d", got, want)
	}
	if code == nil {
		t.Errorf("did not call exit()")
	} else if got, want := *code, 0; got != want {
		t.Errorf("called exit(%d), want exit(%d)", got, want)
	}
	if got, want := errbuf.String(), ""; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestHandlePanic(t *testing.T) {
	var code *int
	swap(t, &exit, func(exitcode int) { code = &exitcode })
	swap(t, &os.Args, []string{"", "panic"})
	errbuf := new(bytes.Buffer)
	swap[io.Writer](t, &stderr, errbuf)
	handler := FuncHandler(func() error { panic("fail") })

	Handle(handler)

	if code == nil {
		t.Errorf("exit() not called")
	} else if got, want := *code, 1; got != want {
		t.Errorf("called exit(%d), want exit(%d)", got, want)
	}
	if got, want := errbuf.String(), "fail\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestHandleError(t *testing.T) {
	var code *int
	swap(t, &exit, func(exitcode int) { code = &exitcode })
	swap(t, &os.Args, []string{"", "error"})
	errbuf := new(bytes.Buffer)
	swap[io.Writer](t, &stderr, errbuf)
	handler := FuncHandler(func() error { return errors.New("fail") })

	Handle(handler)

	if code == nil {
		t.Errorf("exit() not called")
	} else if got, want := *code, 1; got != want {
		t.Errorf("called exit(%d), want exit(%d)", got, want)
	}
	if got, want := errbuf.String(), "fail\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestPassPanic(t *testing.T) {
	var code *int
	swap(t, &exit, func(exitcode int) { code = &exitcode })
	swap(t, &os.Args, []string{"", "error"})
	errbuf := new(bytes.Buffer)
	swap[io.Writer](t, &stderr, errbuf)
	handler := FuncHandler(func() error { panic("fail") })
	defer func() {
		r := recover()
		if r == nil {
			t.Error("recover() = <nil>, want string")
		} else if s, ok := r.(string); !ok {
			t.Errorf("recover() = %T, want string", r)
		} else if got, want := s, "fail"; got != want {
			t.Errorf("recover() = %q, want %q", got, want)
		}
		if code == nil {
			t.Errorf("exit() not called")
		} else if got, want := *code, 0; got != want {
			t.Errorf("called exit(%d), want exit(%d)", got, want)
		}
		if got, want := errbuf.String(), ""; got != want {
			t.Errorf("stderr = %q, want %q", got, want)
		}
	}()

	Handle(handler)
}

type BadHandler struct{}

func (BadHandler) TooManyReturns() (error, int) {
	return nil, 0
}

func TestInvalidMethod(t *testing.T) {
	var code *int
	swap(t, &exit, func(exitcode int) { code = &exitcode })
	// Validation of all op funcs occurs before running an op.
	// It should not matter that this test targets a nonexistent op.
	swap(t, &os.Args, []string{"", "anyop"})
	errbuf := new(bytes.Buffer)
	swap[io.Writer](t, &stderr, errbuf)

	Handle(new(BadHandler))

	if code == nil {
		t.Errorf("exit() not called")
	} else if got, want := *code, 1; got != want {
		t.Errorf("called exit(%d), want exit(%d)", got, want)
	}
	wanterr := "bad op: bad signature: func (*ops.BadHandler) (error, int)\n"
	if got, want := errbuf.String(), wanterr; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

type ParamHandler func(int, string)

func (h ParamHandler) TakesParameters(i int, s string) { h(i, s) }

func TestHandlerWithParameters(t *testing.T) {
	var code *int
	swap(t, &exit, func(exitcode int) { code = &exitcode })
	swap(t, &os.Args, []string{"", "takes_parameters"})
	errbuf := new(bytes.Buffer)
	swap[io.Writer](t, &stderr, errbuf)
	handler := ParamHandler(func(i int, s string) {
		if got, want := i, 0; got != want {
			t.Errorf("ParamHandler i = %d, want %d", got, want)
		}
		if got, want := s, ""; got != want {
			t.Errorf("ParamHandler s = %q, want %q", got, want)
		}
	})

	Handle(handler)

	if code == nil {
		t.Errorf("did not call exit()")
	} else if got, want := *code, 0; got != want {
		t.Errorf("called exit(%d), want exit(%d)", got, want)
	}
	if got, want := errbuf.String(), ""; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

var NonPtrFnCalled bool

type S1 struct{}
type S2 struct {
	S1
}
type S3 struct{}
type S4 struct {
	S2
	S3
}
type S5 struct {
	S1
}
type S6 struct {
	S4
}
type S7 struct {
	S4
}
type S8 struct {
	S1
	ptrFnCalled bool
}
type S9 struct{}
type B1 bool

func (S1) F1() {}
func (S2) F2() {}
func (S3) F2() {}
func (S5) F1() {}
func (S7) F2() {}
func (s *S8) PtrFn() {
	s.ptrFnCalled = true
}
func (S9) NonPtrFn() {
	NonPtrFnCalled = true
}
func (B1) F1() {}

func TestMethodsByName(t *testing.T) {
	tests := []struct {
		rtype reflect.Type
		name  string
		want  []string
	}{{
		reflect.TypeOf(S1{}),
		"f1",
		[]string{"S1.F1"},
	}, {
		reflect.TypeOf(S2{}),
		"f2",
		[]string{"S2.F2"},
	}, {
		reflect.TypeOf(S2{}),
		"f1",
		[]string{"S2.F1"},
	}, {
		reflect.TypeOf(S4{}),
		"f2",
		[]string{"S2.F2", "S3.F2"},
	}, {
		reflect.TypeOf(S5{}),
		"f1",
		[]string{"S5.F1"},
	}, {
		reflect.TypeOf(S6{}),
		"f2",
		[]string{"S2.F2", "S3.F2"},
	}, {
		reflect.TypeOf(S7{}),
		"f2",
		[]string{"S7.F2"},
	}, {
		reflect.TypeOf(S8{}),
		"ptr_fn",
		[]string{"S8.PtrFn"},
	}, {
		reflect.TypeOf(B1(true)),
		"f1",
		[]string{"B1.F1"},
	}, {
		reflect.TypeOf(new(B1)),
		"f1",
		[]string{"B1.F1"},
	}}
	for _, tt := range tests {
		t.Run(strings.Join(tt.want, "/"), func(t *testing.T) {
			got := []string{}
			for _, m := range methodsByName(tt.rtype, tt.name) {
				got = append(got, methodname(m))
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("methodsByName(%s, %q) -want +got:\n%s",
					tt.rtype, tt.name,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func TestMethodNames(t *testing.T) {
	tests := []struct {
		rtype reflect.Type
		want  []string
	}{{
		reflect.TypeOf(S1{}),
		[]string{"F1"},
	}, {
		reflect.TypeOf(S2{}),
		[]string{"F1", "F2"},
	}, {
		reflect.TypeOf(S3{}),
		[]string{"F2"},
	}, {
		reflect.TypeOf(S4{}),
		[]string{"F1", "F2"},
	}, {
		reflect.TypeOf(S5{}),
		[]string{"F1"},
	}, {
		reflect.TypeOf(S8{}),
		[]string{"F1", "PtrFn"},
	}, {
		reflect.TypeOf(S9{}),
		[]string{"NonPtrFn"},
	}, {
		reflect.TypeOf(B1(true)),
		[]string{"F1"},
	}}
	for _, tt := range tests {
		t.Run(tt.rtype.Name(), func(t *testing.T) {
			if got := methodNames(tt.rtype); !cmp.Equal(got, tt.want) {
				t.Errorf("methodNames(%s) -want +got:\n%s",
					tt.rtype,
					cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestCallPointerReceiver(t *testing.T) {
	s := S8{}

	err := opHandler(&s, "ptr_fn")

	if err != nil {
		t.Errorf(`opHandler(s, "ptr_fn") = %q, want <nil>`, err)
	}
	if !s.ptrFnCalled {
		t.Errorf(`opHandler(s, "ptr_fn"): PtrFn() not called`)
	}
}

func TestCallNonPointerReceiver(t *testing.T) {
	t.Cleanup(func() { NonPtrFnCalled = false })

	err := opHandler(S9{}, "non_ptr_fn")

	if err != nil {
		t.Errorf(`opHandler(s, "non_ptr_fn") = %q, want <nil>`, err)
	}
	if !NonPtrFnCalled {
		t.Errorf(`opHandler(s, "non_ptr_fn"): NonPtrFn() not called`)
	}
}

func TestCallNonPointerReceiverWithPointer(t *testing.T) {
	t.Cleanup(func() { NonPtrFnCalled = false })

	err := opHandler(new(S9), "non_ptr_fn")

	if err != nil {
		t.Errorf(`opHandler(s, "non_ptr_fn") = %q, want <nil>`, err)
	}
	if !NonPtrFnCalled {
		t.Errorf(`opHandler(s, "non_ptr_fn"): NonPtrFn() not called`)
	}
}

func TestCallPointerReceiverWithNonPointer(t *testing.T) {
	// This test exists to ensure we do not error or panic.
	// Since we are not operating on a pointer, S8.ptrFnCalled will not update.
	if err := opHandler(S8{}, "ptr_fn"); err != nil {
		t.Errorf(`opHandler(s, "ptr_fn") = %q, want <nil>`, err)
	}
}

func methodname(m reflect.Method) string {
	return m.Func.Type().In(0).Elem().Name() + "." + m.Name
}

func swap[T any](t *testing.T, orig *T, with T) {
	t.Helper()
	o := *orig
	t.Cleanup(func() { *orig = o })
	*orig = with
}
