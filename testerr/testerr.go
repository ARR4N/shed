// Package testerr provides mechanisms for specifying expected properties of
// errors during testing.
package testerr

import (
	"errors"
	"fmt"
	"strings"
)

// Want defines a type that can compare an error to an expected value or
// property. If `got` meets the expectation, UnmetBy returns the empty string.
// Otherwise it returns a description of the expectation alone, phrased to read
// after the word "want" (e.g. "nil", or "error that Is() foo"); it MUST NOT
// describe `got`, which [Diff] renders.
type Want interface {
	UnmetBy(got error) (expectation string)
}

// Diff compares the error with what is wanted, returning the empty string if
// the expectation is met, otherwise a message suitable for test failures. A nil
// [Want] corresponds to a nil error.
func Diff(got error, want Want) string {
	if u := unmetBy(got, want); u != "" {
		return DiffMessage(got, u)
	}
	return ""
}

func unmetBy(got error, want Want) string {
	if want == nil {
		if got == nil {
			return ""
		}
		return "nil"
	}
	return want.UnmetBy(got)
}

// DiffMessage constructs a canonical diff message for use in test failures.
//
// Deprecated: use [Diff].
func DiffMessage(got error, wantFormat string, a ...any) string {
	format := fmt.Sprintf("got error %%v; want %s", wantFormat)
	return fmt.Sprintf(format, append([]any{got}, a...)...)
}

// A Func is an adaptor to convert an ordinary function into a [Want] by calling
// the function as the implementation of [Want.UnmetBy].
type Func func(got error) (expectation string)

// UnmetBy implements [Want] by calling `fn` itself.
func (fn Func) UnmetBy(got error) string {
	return fn(got)
}

// Is checks that the `got` error [errors.Is] `target`.
func Is(target error) Want {
	return Func(func(got error) string {
		if errors.Is(got, target) {
			return ""
		}
		return fmt.Sprintf("error that Is() %v", target)
	})
}

// As checks that the `got` error can be unwrapped via [errors.As] into a new
// `T`. A non-nil `unmet` function is called to check the resulting `T`, and its
// result is returned by [Want.UnmetBy]; a nil function allows all errors of
// type `T`.
func As[T error](unmet func(got T) (expectation string)) Want {
	return Func(func(got error) string {
		var target T
		if !errors.As(got, &target) {
			return fmt.Sprintf("error tree containing type %T", target)
		}
		if unmet == nil {
			return ""
		}
		return unmet(target)
	})
}

// Equals checks that `got == want`. [Is] SHOULD be used instead.
func Equals(want error) Want {
	return Func(func(got error) string {
		if got == want {
			return ""
		}
		return fmt.Sprintf("== %v", want)
	})
}

// Contains checks that the `got` error's string contains the substring. Note
// that Contains("") matches any non-nil error, but never a nil one, which is
// represented by a nil [Want].
func Contains(substr string) Want {
	return Func(func(got error) string {
		if got != nil && strings.Contains(got.Error(), substr) {
			return ""
		}
		return fmt.Sprintf("containing substring %q", substr)
	})
}

// AnyOf checks that `got` matches at least one of the provided [Want] values,
// any of which MAY be nil.
func AnyOf(a, b Want, rest ...Want) Want {
	all := append([]Want{a, b}, rest...)

	return Func(func(got error) string {
		unmet := make([]string, len(all))
		for i, w := range all {
			unmet[i] = unmetBy(got, w)
			if unmet[i] == "" {
				return ""
			}
		}
		return strings.Join(unmet, " OR ")
	})
}
