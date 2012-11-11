package app

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
)

type AppError struct {
	CallStack   string
	OriginalErr error
	Code        int
}

func (err *AppError) Error() string {
	return fmt.Sprintf("[status code %d] %s\n\n%s", err.Code, err.OriginalErr, err.CallStack)
}

func (err *AppError) Log() {
	if !strings.Contains(err.OriginalErr.Error(), "exec error:") {
		log.Printf("ERROR: %s\n", err.Error())
	}
}

func Errorf(format string, args ...interface{}) error {
	return Error(fmt.Errorf(format, args...))
}

func Error(original error) error {
	return &AppError{
		OriginalErr: original,
		Code:        500,
		CallStack:   fmt.Sprintf("%s", debug.Stack()),
	}
}

func NotFound() error {
	return &AppError{
		Code:      404,
		CallStack: fmt.Sprintf("%s", debug.Stack()),
	}
}

func Forbidden() error {
	return &AppError{
		Code:      403,
		CallStack: fmt.Sprintf("%s", debug.Stack()),
	}
}

func NotAllowed() error {
	return &AppError{
		Code:      405,
		CallStack: fmt.Sprintf("%s", debug.Stack()),
	}
}
