package ops

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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

func methodname(m reflect.Method) string {
	return m.Func.Type().In(0).Elem().Name() + "." + m.Name
}
