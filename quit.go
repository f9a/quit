package quit

import (
	"fmt"
	"os"
)

// Code represents the programm exit code
type Code int

func (c Code) apply(opts *catchOptions) {
	opts.code = c
	opts.isCode = true
}

// Error error
type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %v - exit with %d", e.Message, e.Err, e.Code)
}

// ErrorHandler handels err. Error is never nil. The ErrorHandler can assume
// that at least one of the following fields is set Message, Err.
type ErrorHandler func(Error)

// NopHandler does absolutely nothing with err.
func NopHandler(err Error) {}

// PrintHandler prints err to stdout.
func PrintHandler(err Error) {
	switch {
	case err.Err == nil && err.Message != "":
		fmt.Println(err.Message)
	case err.Err != nil && err.Message == "":
		fmt.Println(err.Err)
	case err.Err != nil && err.Message != "":
		fmt.Printf("%s: %v\n", err.Message, err.Err)
	}
}

type catchOptions struct {
	code   Code
	isCode bool
}

// Option is a Catch option
type Option interface {
	apply(*catchOptions)
}

// Catch handles the graceful exit of the program.
// If os.Exit is called directly in any other function the deferred functions are NOT executed.
// Therefore Catch must be called within the first defer statement in the main function.
// If then a function like to exit the program panic(quit.Code(int)) must be called.
// If no exit code is given default is 1. If handler is nil PrintHandler is used.
func Catch(handler ErrorHandler, options ...Option) {
	errCode := Code(1)
	opts := catchOptions{}

	for _, opt := range options {
		opt.apply(&opts)
	}

	if opts.isCode {
		errCode = opts.code
	}

	if handler == nil {
		handler = PrintHandler
	}

	if e := recover(); e != nil {
		switch x := e.(type) {
		case Code:
			os.Exit(int(x))
		case Error:
			x.Code = errCode

			if x.Err == nil && x.Message == "" {
				handler(Error{
					Message: "invalid quit.Error was send to quit.Catch, either Err or Message must be given",
				})
			} else {
				handler(x)
			}

			os.Exit(int(x.Code))
		default:
			panic(e) // unknown error type, bubble up
		}
	}
}

// With panic with given code
func With(code int) {
	panic(Code(code))
}

// OnErrf quit when err is not nil.
func OnErrf(err error, format string, args ...interface{}) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		panic(Error{
			Message: msg,
			Err:     err,
		})
	}
}

// OnErr quit when err is not nil.
func OnErr(err error) {
	if err != nil {
		panic(Error{
			Err: err,
		})
	}
}

// WithErr quit with err.
func WithErr(err error) {
	panic(Error{
		Err: err,
	})
}

// WithErrf quit with message.
func WithErrf(err error, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	panic(Error{
		Message: msg,
		Err:     err,
	})
}

// WithMsgf quit with message.
func WithMsgf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	panic(Error{
		Message: msg,
	})
}
