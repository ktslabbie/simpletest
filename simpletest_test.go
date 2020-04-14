package simpletest

import (
	"errors"
	"fmt"
	"testing"
)

var (
	inputOrErrFunc = func(c *Case) (output interface{}, err error) {
		if len(c.Error) > 0 {
			err = errors.New(c.Error)
		}

		return c.Input, err
	}

	identityFunc = func(c *Case) (output interface{}, err error) {
		return c.Input, err
	}

	errorFunc = func(c *Case) (output interface{}, err error) {
		return nil, errors.New("unexpected")
	}

	panicFunc = func(c *Case) (output interface{}, err error) {
		panic("panicked")
	}

	panicErrorFunc = func(c *Case) (output interface{}, err error) {
		panic(fmt.Errorf("panicked"))
	}
)

func TestRun(t *testing.T) {
	successCases := Cases{
		"Output matches what we wanted": {Input: "A", Want: "A"},
		"Expected error":                {Input: "A", Error: "foo"},
	}

	Run(t, successCases, inputOrErrFunc)
}

func Test_areBothNil(t *testing.T) {
	var nilVar *interface{}

	successCases := Cases{
		"nil == nil value with different type": {Input: nil, Want: nilVar},
	}

	Run(t, successCases, inputOrErrFunc)
}

func Test_areEqualZeroLengthSlices(t *testing.T) {
	var nilSlice []string

	successCases := Cases{
		"Empty slice == nil slice": {Input: []string{}, Want: nilSlice},
		"Nil slice == empty slice": {Input: nilSlice, Want: []string{}},
	}

	Run(t, successCases, inputOrErrFunc)
}

func Test_execute(t *testing.T) {
	failureCase := Case{Input: "A", Want: "B"}
	expectedError := fmt.Sprintf(FailedTemplate, nil, failureCase.Input, failureCase.Want, failureCase.Input)
	t.Run("Output does not match what we wanted", doTestExecuteError(failureCase, inputOrErrFunc, expectedError))

	unexpectedErrorCase := Case{Input: "A", Want: "B"}
	expectedUnexpectedError := fmt.Sprintf(UnexpectedErrorTemplate, nil, unexpectedErrorCase.Input, unexpectedErrorCase.Want, "unexpected")
	t.Run("Unexpected error", doTestExecuteError(unexpectedErrorCase, errorFunc, expectedUnexpectedError))

	wrongErrorCase := Case{Input: "A", Error: "B"}
	expectedWrongError := fmt.Sprintf(WrongErrorTemplate, nil, wrongErrorCase.Input, wrongErrorCase.Error, "unexpected")
	t.Run("Wrong error", doTestExecuteError(wrongErrorCase, errorFunc, expectedWrongError))

	expectedErrorNotThrownCase := Case{Input: "A", Error: "B"}
	expectedNotThrownError := fmt.Sprintf(ExpectedErrorNotThrownTemplate, nil, expectedErrorNotThrownCase.Input, expectedErrorNotThrownCase.Error, "A")
	t.Run("Expected error not thrown", doTestExecuteError(expectedErrorNotThrownCase, identityFunc, expectedNotThrownError))

	unexpectedPanicCase := Case{Input: "A", Want: "B"}
	expectedUnexpectedPanicError := fmt.Sprintf(UnexpectedErrorTemplate, nil, unexpectedPanicCase.Input, unexpectedPanicCase.Want, "panicked")
	t.Run("Unexpected panic", doTestExecuteError(unexpectedPanicCase, panicFunc, expectedUnexpectedPanicError))
	t.Run("Unexpected panic with error", doTestExecuteError(unexpectedPanicCase, panicErrorFunc, expectedUnexpectedPanicError))
}

func doTestExecuteError(testCase Case, f func(c *Case) (output interface{}, err error), expectedError string) func(t *testing.T) {
	return func(t *testing.T) {
		err := execute(testCase, f)
		if err == nil {
			t.Fatal("Expected an error")
		}

		if err.Error() != expectedError {
			t.Error("Expected error", err, "to match", expectedError)
		}
	}
}
