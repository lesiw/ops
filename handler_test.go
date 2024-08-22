package ops

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type TestHandler int

func (o *TestHandler) Incr() { *o++ }
func (o TestHandler) Fail()  { panic(errors.New("fail")) }

type PrintableError struct{ error }

func (p PrintableError) Print(w io.Writer) {
	fmt.Fprintln(w, "---")
	fmt.Fprintln(w, p.Error())
	fmt.Fprintln(w, "---")
}

func (o TestHandler) PrintableFail() {
	panic(PrintableError{errors.New("fail")})
}

func TestHandleSuccess(t *testing.T) {
	var code *int
	swap(t, &exit, func(exitcode int) { code = &exitcode })
	swap(t, &os.Args, []string{"", "incr"})
	errbuf := new(bytes.Buffer)
	swap[io.Writer](t, &stderr, errbuf)
	counter := new(TestHandler)

	Handle(counter)

	if got, want := int(*counter), 1; got != want {
		t.Errorf("got %d TestHandler.Incr() calls, want %d", got, want)
	}
	if code == nil {
		t.Errorf("TestHandler fail op: did not call exit()")
	} else if got, want := *code, 0; got != want {
		t.Errorf("TestHandler fail op: got exit(%d), want exit(%d)", got, want)
	}
	if got, want := errbuf.String(), ""; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestHandleFailure(t *testing.T) {
	var code *int
	swap(t, &exit, func(exitcode int) { code = &exitcode })
	swap(t, &os.Args, []string{"", "fail"})
	errbuf := new(bytes.Buffer)
	swap[io.Writer](t, &stderr, errbuf)
	counter := new(TestHandler)

	Handle(counter)

	if code == nil {
		t.Errorf("TestHandler fail op: did not call exit()")
	} else if got, want := *code, 1; got != want {
		t.Errorf("TestHandler fail op: got exit(%d), want exit(%d)", got, want)
	}
	if got, want := errbuf.String(), "fail\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestHandlePrintable(t *testing.T) {
	var code *int
	swap(t, &exit, func(exitcode int) { code = &exitcode })
	swap(t, &os.Args, []string{"", "printable_fail"})
	errbuf := new(bytes.Buffer)
	swap[io.Writer](t, &stderr, errbuf)
	counter := new(TestHandler)

	Handle(counter)

	if code == nil {
		t.Errorf("TestHandler fail op: did not call exit()")
	} else if got, want := *code, 1; got != want {
		t.Errorf("TestHandler fail op: got exit(%d), want exit(%d)", got, want)
	}
	if got, want := errbuf.String(), "---\nfail\n---\n"; got != want {
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
