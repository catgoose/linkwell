package linkwell

import "fmt"

// DefaultErrorStatusTarget is the CSS selector for the error status container.
// Used by OOB placements and the response builder to target the error panel.
const DefaultErrorStatusTarget = "#error-status"

// ErrorContext is the single struct that flows through the hypermedia error pipeline.
// It carries all information needed to render a richly-linked error response.
type ErrorContext struct {
	Err        error
	Message    string
	Route      string
	RequestID  string
	OOBTarget  string
	OOBSwap    string
	Controls   []Control
	StatusCode int
	Closable   bool
	Theme      string // DaisyUI theme for full-page error renders; empty = "dark"
}

// WithControls returns a copy of the ErrorContext with the given controls appended.
func (ec ErrorContext) WithControls(controls ...Control) ErrorContext {
	ec.Controls = append(append([]Control(nil), ec.Controls...), controls...)
	return ec
}

// WithOOB returns a copy of the ErrorContext configured for OOB rendering.
func (ec ErrorContext) WithOOB(target, swap string) ErrorContext {
	ec.OOBTarget = target
	ec.OOBSwap = swap
	return ec
}

// HTTPError is a returnable error that wraps an ErrorContext.
// middleware.ErrorHandlerMiddleware detects it via errors.As and renders
// the full hypermedia error response including action controls.
type HTTPError struct {
	EC ErrorContext
}

func (e *HTTPError) Error() string {
	if e.EC.Err != nil {
		return fmt.Sprintf("HTTP %d: %s: %v", e.EC.StatusCode, e.EC.Message, e.EC.Err)
	}
	return fmt.Sprintf("HTTP %d: %s", e.EC.StatusCode, e.EC.Message)
}

// NewHTTPError wraps an ErrorContext in an HTTPError ready to be returned
// from a handler. The error handler middleware will intercept and render it.
func NewHTTPError(ec ErrorContext) *HTTPError {
	return &HTTPError{EC: ec}
}
