package errs

import (
	"errors"
	"strings"
)

var (
	ErrValidation = errors.New("validation failed")
	ErrNotFound   = errors.New("not found")
	ErrNotUnique  = errors.New("not unique")
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Fields []FieldError `json:"fields"`
}

func (e *ValidationError) Error() string {
	if len(e.Fields) == 0 {
		return "validation failed"
	}

	var r []string = make([]string, 0, len(e.Fields)+1)
	r = append(r, "validation failed: ")

	for _, v := range e.Fields {
		r = append(r, v.Field+": "+v.Message)
	}

	return strings.Join(r, " ")
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrValidation
}

func NewValidation(field, msg string) error {
	return &ValidationError{Fields: []FieldError{{field, msg}}}
}
