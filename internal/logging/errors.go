// autodoc/internal/logging/errors.go

package logging

// Error represents a custom error type for logging
type Error struct {
	Message string
	Code    int
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}
