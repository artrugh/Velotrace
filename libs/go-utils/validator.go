package utils

import (
	"github.com/go-playground/validator/v10"
)

// CustomValidator is a shared Echo-compatible validator
type CustomValidator struct {
	Validator *validator.Validate
}

// NewValidator returns a new instance of the shared validator
func NewValidator() *CustomValidator {
	return &CustomValidator{
		Validator: validator.New(),
	}
}

// Validate implements the echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.Validator.Struct(i); err != nil {
		return err
	}
	return nil
}
