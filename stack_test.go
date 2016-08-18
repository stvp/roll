package roll

import (
	"runtime"
	"strings"
	"testing"
)

func TestWalkStack(t *testing.T) {
	frame := walkStack(1)[0]
	if !strings.HasSuffix(frame.Filename, "/roll/stack_test.go") {
		t.Errorf("got: %s", frame.Filename)
	}
	if frame.Method != "roll.TestWalkStack" {
		t.Errorf("got: %s", frame.Method)
	}
	if frame.Line != 10 {
		t.Errorf("got: %d", frame.Line)
	}
}

func TestBuildStack(t *testing.T) {
	frame := buildStack(testStack())[1] // skip the testStack frame
	if !strings.HasSuffix(frame.Filename, "/roll/stack_test.go") {
		t.Errorf("got: %s", frame.Filename)
	}
	if frame.Method != "roll.TestBuildStack" {
		t.Errorf("got: %s", frame.Method)
	}
	if frame.Line != 23 {
		t.Errorf("got: %d", frame.Line)
	}
}

func TestStackFingerprint(t *testing.T) {
	tests := []struct {
		Fingerprint string
		Title       string
		Stack       stack
	}{
		{
			"c9dfdc0e",
			"broken",
			stack{
				frame{"foo.go", "Oops", 1},
			},
		},
		{
			"21037bf5",
			"very broken",
			stack{
				frame{"foo.go", "Oops", 1},
			},
		},
		{
			"50d68db4",
			"broken",
			stack{
				frame{"foo.go", "Oops", 2},
			},
		},
		{
			"b341ee82",
			"broken",
			stack{
				frame{"foo.go", "Oops", 1},
				frame{"foo.go", "Oops", 2},
			},
		},
	}

	for i, test := range tests {
		fp := stackFingerprint(test.Title, test.Stack)
		if fp != test.Fingerprint {
			t.Errorf("tests[%d]: got %s", i, fp)
		}
	}
}

func TestShortenFilePath(t *testing.T) {
	tests := []struct {
		Given    string
		Expected string
	}{
		{"", ""},
		{"foo.go", "foo.go"},
		{"/usr/local/go/src/pkg/runtime/proc.c", "pkg/runtime/proc.c"},
		{"/home/foo/go/src/github.com/stvp/rollbar.go", "github.com/stvp/rollbar.go"},
	}
	for i, test := range tests {
		got := shortenFilePath(test.Given)
		if got != test.Expected {
			t.Errorf("tests[%d]: got %s", i, got)
		}
	}
}

func testStack() []uintptr {
	var s []uintptr
	for i := 0; ; i++ {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		s = append(s, pc)
	}
	return s
}
