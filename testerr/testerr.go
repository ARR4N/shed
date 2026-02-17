// Package testerr provides mechanisms for specifying expected properties of
// errors during testing.
package testerr

import (
	"errors"
	"fmt"
	"strings"
)

// Want defines a type that can compare an error to an expected value or
// property. The return of an empty string is idiomatically considered to
// represent an expected `got` while any unexpected `error` is represented by a
// non-empty string representing the diff. See [DiffMessage] for constructing
// canonical diffs.
type Want interface {
	ErrDiff(got error) string
}

// Diff compares the error with what is wanted. A nil [Want] corresponds to a
// nil error.
func Diff(got error, want Want) string {
	if want == nil {
		if got == nil {
			return ""
		}
		return DiffMessage(got, "nil")
	}
	return want.ErrDiff(got)
}

// DiffMessage constructs a canonical diff message for use in test failures.
func DiffMessage(got error, wantFormat string, a ...any) string {
	format := fmt.Sprintf("got error %%v; want %s", wantFormat)
	return fmt.Sprintf(format, append([]any{got}, a...)...)
}

// A Func is an adaptor to convert an ordinary function into a [Want] by calling
// the function in lieu of `ErrDiff()`.
type Func func(error) string

// ErrDiff implements [Want] by calling `fn` itself.
func (fn Func) ErrDiff(got error) string {
	return fn(got)
}

// Is checks that the `got` error [errors.Is] `target`.
func Is(target error) Want {
	return Func(func(got error) string {
		if errors.Is(got, target) {
			return ""
		}
		return DiffMessage(got, "error that Is() %v", target)
	})
}

// As creates a new `T` and checks that the `got` error can be unwrapped via
// [errors.As] to said type. The unwrapped error is passed to `match()` for
// checking.
//
// The return of an empty string from `match()` results in `ErrDiff()`
// also returning an empty string. On mismatch there is no need to prepend the
// `expected` description with the `got` message. See the [Diff] example.
func As[T error](match func(got T) (expected string)) Want {
	return Func(func(got error) string {
		var target T
		if !errors.As(got, &target) {
			return DiffMessage(got, "error tree containing type %T", target)
		}
		if d := match(target); d != "" {
			return DiffMessage(got, "%s", d)
		}
		return ""
	})
}

// Equals checks that `got == want`. [Is] SHOULD be used instead.
func Equals(want error) Want {
	return Func(func(got error) string {
		if got == want {
			return ""
		}
		return DiffMessage(got, "== %v", want)
	})
}

// Contains checks that the `got` error's string contains the substring. Note
// that the empty string is *not* the same as a nil error, for which a nil
// [Want] MUST be used.
func Contains(substr string) Want {
	return Func(func(got error) string {
		if got != nil && strings.Contains(got.Error(), substr) {
			return ""
		}
		return DiffMessage(got, "containing substring %q", substr)
	})
}
