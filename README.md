simpletest
-----

`simpletest` is a small Go library to simplify [table-driven testing](https://github.com/golang/go/wiki/TableDrivenTests).

### How it works
Imagine a simple function:
```go
func ToUpperCase(it string) (string, error) {
	if it == "" {
		return "", errors.New("invalid input")
	}
	return strings.ToUpper(it), nil
}
```

To test this function, including error scenarios, a common pattern is to use table-driven testing:
```go
func TestToUpperCase(t *testing.T) {
	var testCases = map[string]struct {
		input       string
		want        string
		expectedErr string
	}{
		"OK": {
			input: "test",
			want:  "TEST",
		},
		"Expect error": {
			input:       "",
			expectedErr: "invalid input",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := ToUpperCase(testCase.input)

			if err != nil {
				if testCase.expectedErr != "" {
					if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedErr)) {
						return
					}

					t.Fatal("expected error", testCase.expectedErr, "to match", err)
				}

				t.Fatal("unexpected error")
			}

			if !reflect.DeepEqual(testCase.want, got) {
				t.Error("expected", testCase.want, "to match", got)
			}
		})
	}
}
```

There's a lot of boilerplate here, and we usually end up writing very similar code for every function we test.
`simpletest` allows to populate a generic test case struct, forgoing the need to define it for each test:
```go
type (
    Case struct {
        Receiver interface{} // The receiver (can be undefined if not testing a receiver method).
        Input    interface{} // The input to the function under test.
        Want     interface{} // The desired output of the function under test.
        Error    string      // If an error is expected, then the error message should contain this substring.
    }

    // Cases is a mapping from test names to test cases.
    Cases map[string]Case
)
```

It will process all equality and error checks under the hood. The same test becomes:

```go
func TestToUpperCase(t *testing.T) {
	testCases := simpletest.Cases{
		"OK": simpletest.Case{
			Input: "test",
			Want:  "TEST",
		},
		"Expect error": simpletest.Case{
			Input: "",
			Error: "invalid input",
		},
	}

	simpletest.Run(t, testCases, func(tc *simpletest.Case) (interface{}, error) {
		return ToUpperCase(tc.Input.(string))
	})
}
```

The downside is that we need to add type assertions since Go has no generics.

For testing methods, you can specify a `Receiver`:
```go
type MyString struct {
	Value string
}

func (s MyString) Append(it string) string {
	return s.Value + it
}
```

```go
func TestMyString_Append(t *testing.T) {
	testCases := simpletest.Cases{
		"OK": simpletest.Case{
			Receiver: MyString{Value: "te"},
			Input:    "st",
			Want:     "test",
		},
	}

	simpletest.Run(t, testCases, func(tc *simpletest.Case) (interface{}, error) {
		return tc.Receiver.(MyString).Append(tc.Input.(string)), nil
	})
}
```

For functions that take more than one input or return more than one value, the values can be wrapped into a struct.

### Failure handling
Four types of test failures are handled, each outputting appropriate message templates with input, expected and received values:
* Failed equality assertion
* Error was returned even though none was expected
* Error message didn't contain the expected substring
* Expected error was not returned

Panics are recovered from and converted to normal error values, so panics can be tested as well.

### Execution order
Since test cases are defined as a `map[string]Case`, they are executed in random order.
Usually this is fine, but in case order needs to be enforced, use multiple invocations of the `simpletest.RunSingle` function.
But at that point, it might be easier to use a library like [testify](https://github.com/stretchr/testify) instead.

### Nil handling
Nil handling is slightly opinionated:

* Nil slices and empty slices are considered equal
  * This means if we have e.g. `var nilSlice []string`, then `nilSlice == []string{} => true`
* Nil values are considered equal regardless of type
  * This means if we have e.g. `var nilVar *interface{}`, then `nilVar == nil => true`

This goes against how nil equality in Go works, but in practice, we should almost never care about these differences, 
especially in tests where this often leads to unexpected results when comparing expected to actual values.
As a bonus, this allows to inline expected values in test cases, without requiring separate `var` declarations.
