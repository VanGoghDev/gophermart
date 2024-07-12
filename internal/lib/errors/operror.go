package errors

import "fmt"

type OpError struct {
	Inner error
	Op    string
}

func (e *OpError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Inner)
}

func (e *OpError) Unwrap() error {
	return e.Inner
}

func WrapOp(op string, err error) error {
	return &OpError{
		Op:    op,
		Inner: err,
	}
}
