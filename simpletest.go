package simpletest

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

// FailedTemplate is the template for the error message in case a test failed.
const FailedTemplate = `Test failed!
receiver
  %v
with input
  %v
expected value
  %v
but got value
  %v
`

// UnexpectedErrorTemplate is the template for the error message in case a test threw an unexpected error.
const UnexpectedErrorTemplate = `Unexpected error!
receiver
  %v
with input
  %v
expected value
  %v
but got error
  %+v
`

// WrongErrorTemplate is the template for the error message in case a test threw an error different from the one expected.
const WrongErrorTemplate = `Wrong error!
receiver
  %v
with input
  %v
expected error to contain text
  %s
but got error
  %+v
`

// ExpectedErrorNotThrownTemplate is the template for the error message in case a test didn't throw an expected error.
const ExpectedErrorNotThrownTemplate = `Expected error not thrown!
receiver
  %v
with input
  %v
expected error to contain text
  %s
but got nil error, and value
  %v
`

type (
	// Case is a generic test case intended to simplify testing of simple value objects.
	Case struct {
		Receiver interface{} // The receiver (can be undefined if not testing a receiver method).
		Input    interface{} // The input to the function under test.
		Want     interface{} // The desired output of the function under test.
		Error    string      // If an error is expected, then the error message should contain this substring.
	}

	// Cases is a mapping from test names to test cases.
	Cases map[string]Case
)

// Run executes input Cases in a random order (see RunSingle).
// It returns false if any one test case failed, true otherwise.
func Run(t *testing.T, testCases Cases, f func(tc *Case) (interface{}, error)) bool {
	success := true

	for name, testCase := range testCases {
		success = success && RunSingle(t, name, testCase, f)
	}

	return success
}

// RunSingle takes a Case as input, as well as a function that should define how a result is to be obtained
// from the test case. It returns whether the test succeeded or not.
func RunSingle(t *testing.T, name string, testCase Case, f func(tc *Case) (interface{}, error)) bool {
	return t.Run(name, func(t *testing.T) {
		if err := execute(testCase, f); err != nil {
			t.Errorf("%v", err)
		}
	})
}

// execute actually executes each individual test and handles the result.
func execute(testCase Case, f func(*Case) (interface{}, error)) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = handleError(testCase, e)
			} else {
				err = handleError(testCase, toError(r))
			}
		}
	}()

	got, err := f(&testCase)
	if err != nil {
		return handleError(testCase, err)
	}

	if len(testCase.Error) > 0 {
		return errors.Errorf(ExpectedErrorNotThrownTemplate, testCase.Receiver, testCase.Input, testCase.Error, got)
	}

	return compare(testCase, got)
}

// compare compares the wanted output defined in the Case with the actual output,
// returning an error if there is no match.
func compare(testCase Case, got interface{}) error {
	if areBothNil(testCase.Want, got) {
		return nil
	}

	if areEqualZeroLengthSlices(testCase.Want, got) {
		return nil
	}

	if reflect.DeepEqual(testCase.Want, got) {
		return nil
	}

	return errors.Errorf(FailedTemplate, testCase.Receiver, testCase.Input, testCase.Want, got)
}

// areBothNil checks whether two values are nil, regardless of their interface type.
func areBothNil(a interface{}, b interface{}) bool {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	aNil := a == nil || aVal.Kind() == reflect.Ptr && aVal.IsNil()
	bNil := b == nil || bVal.Kind() == reflect.Ptr && bVal.IsNil()

	if aNil && bNil {
		return true
	}

	return false
}

// areEqualZeroLengthSlices checks whether two values are slices, and if so, whether they both satisfy len(slice) == 0.
// This is a workaround for reflect.DeepEqual treating empty slices and nil slices as different.
func areEqualZeroLengthSlices(a interface{}, b interface{}) bool {
	aValue, bValue := reflect.ValueOf(a), reflect.ValueOf(b)

	if aValue.Kind() == reflect.Slice && bValue.Kind() == reflect.Slice {
		if aValue.Len() == 0 && bValue.Len() == 0 {
			return true
		}
	}

	return false
}

// handleError will check if the error thrown by a test was expected,
// and whether the error message contains the expected Error phrase defined in the Case.
func handleError(testCase Case, err error) error {
	if len(testCase.Error) == 0 {
		return errors.Errorf(UnexpectedErrorTemplate, testCase.Receiver, testCase.Input, testCase.Want, err)
	}

	if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.Error)) {
		return nil
	}

	return errors.Errorf(WrongErrorTemplate, testCase.Receiver, testCase.Input, testCase.Error, err)
}

func toError(recovered interface{}) error {
	return fmt.Errorf("%v", recovered)
}
