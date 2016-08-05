package roll

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestBuildStack(t *testing.T) {
	frame := buildStack(1)[0]
	if !strings.HasSuffix(frame.Filename, "/roll/stack_test.go") {
		t.Errorf("got: %s", frame.Filename)
	}
	if frame.Method != "roll.TestBuildStack" {
		t.Errorf("got: %s", frame.Method)
	}
	if frame.Line != 11 {
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

func TestUnwrapStack(t *testing.T) {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	err, ok := errors.Wrap(errors.New("foo bar"), "fooing bar").(stackTracer)
	if !ok {
		t.Fatalf("errors.Wrap didn't return a stackTracer!")
	}

	s := unwrapStack(err.StackTrace())
	if len(s) != 3 {
		t.Errorf("Expected 3 frames, got %d", len(s))
	} else if !strings.HasSuffix(s[0].Filename, "stack_test.go") {
		t.Errorf("Expected first frame to be contained in stack_test.go, found %q instead", s[0].Filename)
	} else if !strings.HasSuffix(s[0].Method, "TestUnwrapStack") {
		t.Errorf("Expected first frame to contain method TestUnwrapStack, found %q instead", s[0].Method)
	}
}
