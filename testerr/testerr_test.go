package testerr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/arr4n/shed/testerr"
)

// numError is an internal error type that carries semantics for very precise
// error checking. In practice this could be, for example, a gRPC `Status` error
// with its `Code` checked with [testerr.As].
type numError struct {
	val int
}

func (e numError) Error() string {
	return fmt.Sprintf("val %d is not good", e.val)
}

func ExampleDiff() {
	errUhOh := errors.New("uh oh")
	errWrapped := fmt.Errorf("wrapped(%w)", errUhOh)
	errOther := errors.New("something else")

	err42 := numError{42}
	err43 := numError{43}
	want42 := testerr.As(func(got numError) string {
		if got.val != 42 {
			return "42 (of course)"
		}
		return ""
	})

	wantNilOrOtherOr42 := testerr.AnyOf(
		nil,
		testerr.Equals(errOther),
		want42,
	)

	tests := []struct {
		name string
		err  error // typically declared in the test loop
		want testerr.Want
	}{
		{
			name: "got nil; want nil",
		},
		{
			name: "got non-nil; want nil",
			err:  errUhOh,
		},
		{
			name: "Equals() when equal",
			err:  errUhOh,
			want: testerr.Equals(errUhOh),
		},
		{
			name: "Equals() when not equal",
			err:  errWrapped,
			want: testerr.Equals(errUhOh),
		},
		{
			name: "Is() when wrapped",
			err:  errWrapped,
			want: testerr.Is(errUhOh),
		},
		{
			name: "Is() when equal",
			err:  errUhOh,
			want: testerr.Is(errUhOh),
		},
		{
			name: "Is() when different",
			err:  errOther,
			want: testerr.Is(errUhOh),
		},
		{
			name: "Contains() when superstring",
			err:  errUhOh,
			want: testerr.Contains("uh"),
		},
		{
			name: "Contains() when different",
			err:  errUhOh,
			want: testerr.Contains("foobar"),
		},
		{
			name: "Contains() when got nil, even with empty substring",
			want: testerr.Contains(""),
		},
		{
			name: "As() without diff from matcher function",
			err:  err42,
			want: want42,
		},
		{
			name: "As() with correct type and no matcher function",
			err:  err42,
			want: testerr.As[numError](nil),
		},
		{
			name: "As() with diff from matcher function",
			err:  err43,
			want: want42,
		},
		{
			name: "As() with incorrect type",
			err:  errUhOh,
			want: want42,
		},
		{
			name: "AnyOf() with matching nil",
			err:  nil,
			want: wantNilOrOtherOr42,
		},
		{
			name: "AnyOf() with matching non-nil",
			err:  errOther,
			want: wantNilOrOtherOr42,
		},
		{
			name: "AnyOf() without matches",
			err:  errUhOh,
			want: wantNilOrOtherOr42,
		},
		{
			name: "AllOf() when all met",
			err:  err42,
			want: testerr.AllOf(
				testerr.Is(err42),
				testerr.As[numError](nil),
			),
		},
		{
			name: "AllOf() when none met",
			err:  errUhOh,
			want: testerr.AllOf(
				testerr.Is(err42),
				testerr.As[numError](nil),
			),
		},
		{
			name: "AllOf() when some met",
			err:  err42,
			want: testerr.AllOf(
				testerr.Is(err42),
				testerr.As[numError](nil),
				nil, // impossible
			),
		},
	}

	for _, tt := range tests {
		if false {
			// Typical usage in tests, assuming these variables exist:
			var (
				t   testing.T
				err error
			)
			if diff := testerr.Diff(err, tt.want); diff != "" {
				t.Errorf("Something(arg) %s", diff)
				// or
				t.Fatalf("Something(arg) %s", diff)
			}
		}

		fmt.Println("---", tt.name, "---")
		if diff := testerr.Diff(tt.err, tt.want); diff != "" {
			fmt.Println(diff)
		} else {
			fmt.Println("<empty>")
		}
	}

	// Output:
	// --- got nil; want nil ---
	// <empty>
	// --- got non-nil; want nil ---
	// got error uh oh; want nil
	// --- Equals() when equal ---
	// <empty>
	// --- Equals() when not equal ---
	// got error wrapped(uh oh); want == uh oh
	// --- Is() when wrapped ---
	// <empty>
	// --- Is() when equal ---
	// <empty>
	// --- Is() when different ---
	// got error something else; want error that Is() uh oh
	// --- Contains() when superstring ---
	// <empty>
	// --- Contains() when different ---
	// got error uh oh; want containing substring "foobar"
	// --- Contains() when got nil, even with empty substring ---
	// got error <nil>; want containing substring ""
	// --- As() without diff from matcher function ---
	// <empty>
	// --- As() with correct type and no matcher function ---
	// <empty>
	// --- As() with diff from matcher function ---
	// got error val 43 is not good; want 42 (of course)
	// --- As() with incorrect type ---
	// got error uh oh; want error tree containing type testerr_test.numError
	// --- AnyOf() with matching nil ---
	// <empty>
	// --- AnyOf() with matching non-nil ---
	// <empty>
	// --- AnyOf() without matches ---
	// got error uh oh; want nil OR == something else OR error tree containing type testerr_test.numError
	// --- AllOf() when all met ---
	// <empty>
	// --- AllOf() when none met ---
	// got error uh oh; want error that Is() val 42 is not good AND error tree containing type testerr_test.numError
	// --- AllOf() when some met ---
	// got error val 42 is not good; want nil AND 2 other condition(s) already met
}
