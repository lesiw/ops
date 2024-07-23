package ci

import (
	"reflect"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

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

func (_ S1) F1() {}
func (_ S2) F2() {}
func (_ S3) F2() {}
func (_ S5) F1() {}
func (_ S7) F2() {}

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
	}}
	for _, tt := range tests {
		t.Run(strings.Join(tt.want, "/"), func(t *testing.T) {
			got := []string{}
			for _, m := range methodsByName(tt.rtype, tt.name) {
				got = append(
					got,
					m.Func.Type().In(0).Name()+"."+m.Name,
				)
			}
			assert.DeepEqual(t, tt.want, got)
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
	}}
	for _, tt := range tests {
		t.Run(tt.rtype.Name(), func(t *testing.T) {
			assert.DeepEqual(t, tt.want, methodNames(tt.rtype))
		})
	}
}
