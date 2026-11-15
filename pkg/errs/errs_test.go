package errs

import (
	"errors"
	"testing"
)

func TestSentinels(t *testing.T) {
	if ErrValidation.Error() != "validation failed" {
		t.Fatalf("ErrValidation: %q", ErrValidation.Error())
	}
	if ErrNotFound.Error() != "not found" {
		t.Fatalf("ErrNotFound: %q", ErrNotFound.Error())
	}
	if ErrNotUnique.Error() != "not unique" {
		t.Fatalf("ErrNotUnique: %q", ErrNotUnique.Error())
	}
}

func TestValidationErrorError(t *testing.T) {
	tests := []struct {
		name string
		err  *ValidationError
		want string
	}{
		{
			name: "empty",
			err:  &ValidationError{},
			want: "validation failed",
		},
		{
			name: "one field",
			err:  &ValidationError{Fields: []FieldError{{Field: "login", Message: "too short"}}},
			want: "validation failed:  login: too short",
		},
		{
			name: "two fields",
			err: &ValidationError{Fields: []FieldError{
				{Field: "login", Message: "too short"},
				{Field: "password", Message: "empty"},
			}},
			want: "validation failed:  login: too short password: empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Fatalf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewValidation(t *testing.T) {
	err := NewValidation("title", "empty field")

	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("NewValidation type = %T", err)
	}
	if len(validation.Fields) != 1 {
		t.Fatalf("fields = %d", len(validation.Fields))
	}
	if validation.Fields[0].Field != "title" || validation.Fields[0].Message != "empty field" {
		t.Fatalf("field = %+v", validation.Fields[0])
	}

	if !errors.Is(err, ErrValidation) {
		t.Fatal("expected errors.Is ErrValidation")
	}
	if errors.Is(err, ErrNotFound) {
		t.Fatal("validation error matched ErrNotFound")
	}
	if errors.Is(err, ErrNotUnique) {
		t.Fatal("validation error matched ErrNotUnique")
	}
}
