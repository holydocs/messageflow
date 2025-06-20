package messageflow

import "fmt"

// UnsupportedFormatError represents an error when an unsupported format is provided.
type UnsupportedFormatError struct {
	given    TargetType
	expected TargetType
}

// NewUnsupportedFormatError creates a new UnsupportedFormatError.
func NewUnsupportedFormatError(given, expected TargetType) error {
	return &UnsupportedFormatError{
		given:    given,
		expected: expected,
	}
}

// Error implements the error interface for UnsupportedFormatError.
func (err *UnsupportedFormatError) Error() string {
	return fmt.Sprintf("%s format is not supported, %s expected", err.given, err.expected)
}

// UnsupportedFormatModeError represents an error when an unsupported format mode is provided.
type UnsupportedFormatModeError struct {
	given    FormatMode
	expected []FormatMode
}

// NewUnsupportedFormatModeError creates a new UnsupportedFormatModeError.
func NewUnsupportedFormatModeError(given FormatMode, expected []FormatMode) error {
	return &UnsupportedFormatModeError{
		given:    given,
		expected: expected,
	}
}

// Error implements the error interface for UnsupportedFormatError.
func (err *UnsupportedFormatModeError) Error() string {
	return fmt.Sprintf("%s format mode is not supported, %v expected", err.given, err.expected)
}
