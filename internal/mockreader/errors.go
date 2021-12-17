package mockreader

import (
	"fmt"
)

type ExpectationFailedError struct {
	Index  uint
	Expect Expectation
	Actual Expectation
}

func (err ExpectationFailedError) Error() string {
	return fmt.Sprintf("MockReader expectation #%d failed: expected «%v»; got «%v»", err.Index, err.Expect, err.Actual)
}

var _ error = ExpectationFailedError{}
