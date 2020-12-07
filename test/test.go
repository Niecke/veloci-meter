package test

import (
	"fmt"
	"runtime"
	"testing"
)

func funcName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[1:n])
	frame, _ := frames.Next()

	return fmt.Sprintf("%s():%v", frame.Func.Name(), frame.Line)
}

// CheckResult is a small helper function which compares result and expected and in case they are not equal some information are logged via t.Errorf().
func CheckResult(t *testing.T, result interface{}, expected interface{}) {
	if result != expected {
		t.Errorf("%v test returned an unexpected result: got %v want %v", funcName(), result, expected)
	}
}
