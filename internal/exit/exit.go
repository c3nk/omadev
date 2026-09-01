// Package exit defines omadev's process exit codes and a typed error that carries
// the code it should produce. cmd translates a returned error into an exit code in
// one place (CodeOf), so command logic can stay focused on behavior.
package exit

import "errors"

// Code is a process exit code with a documented meaning.
type Code int

const (
	// OK indicates success.
	OK Code = 0
	// Generic covers untagged failures and Cobra usage/argument errors.
	Generic Code = 1
	// Unsupported means the project is unknown or not supported (e.g. confidence
	// below HIGH for execution).
	Unsupported Code = 2
	// MissingPrereq means a required host tool is unavailable (e.g. Docker).
	MissingPrereq Code = 3
	// InvalidConfig means project configuration could not be parsed (e.g. Compose).
	InvalidConfig Code = 4
	// Canceled means the user declined at a confirmation prompt.
	Canceled Code = 5
	// ExecFailure means an underlying command returned a non-zero status.
	ExecFailure Code = 6
)

// Error couples an error with the exit code it should produce.
type Error struct {
	Code Code
	Err  error
}

func (e *Error) Error() string { return e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }

// New wraps err with an exit code.
func New(code Code, err error) *Error { return &Error{Code: code, Err: err} }

// CodeOf returns the exit code for err: OK when nil, the code of the first *Error in
// the chain when present, or Generic for any other non-nil error.
func CodeOf(err error) Code {
	if err == nil {
		return OK
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return Generic
}
