// assert provides testing assertions based in generics. This provides
// type safety at compile time, and avoids the use of reflection.
package assert

import (
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, actual, expected T) {
	if expected != actual {
		t.Errorf("got %v, expected %v", actual, expected)
	}
}

func Contains[T comparable](t *testing.T, s []T, elem T) {
	for _, e := range s {
		if e == elem {
			return
		}
	}
	t.Errorf("expected %v to contain %v", s, elem)
}

func StrContains(t *testing.T, s, substr string) {
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

func StrNotContains(t *testing.T, s, substr string) {
	if strings.Contains(s, substr) {
		t.Errorf("expected %q to not contain %q", s, substr)
	}
}
