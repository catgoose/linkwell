package linkwell

import "fmt"

// DefaultErrorStatusTarget is the CSS selector for the error status container
// used by OOB placements and response builders to target the error panel.
const DefaultErrorStatusTarget = "#error-status"

// ErrorContext carries all information needed to render a hypermedia error
// response with action controls. It flows through the error handling pipeline
// from handler to middleware to template. Use WithControls and WithOOB to build
// up the context fluently.
type ErrorContext struct {
	// Err is the underlying Go error (not shown to users, used for logging).
	Err error
	// Message is the user-facing error message.
	Message string
	// Route is the request path that produced the error (for logging/reporting).
	Route string
	// RequestID is the correlation ID for the request (passed to ReportIssueButton).
	RequestID string
	// OOBTarget is the CSS selector for an HTMX OOB swap target. When set, the
	// error renders as an out-of-band fragment rather than the main response.
	OOBTarget string
	// OOBSwap is the hx-swap-oob strategy (e.g., "innerHTML", "outerHTML").
	OOBSwap string
	// Controls is the set of hypermedia affordances rendered with the error
	// (e.g., Retry, Back, Dismiss buttons).
	Controls []Control
	// StatusCode is the HTTP status code for the error response.
	StatusCode int
	// Closable indicates whether the error panel should show a close/dismiss button.
	Closable bool
	// Theme is the DaisyUI theme for full-page error renders. Empty defaults to "dark".
	Theme string
}

// WithControls returns a copy of the ErrorContext with the given controls
// appended to the existing set. The original is not modified.
func (ec ErrorContext) WithControls(controls ...Control) ErrorContext {
	ec.Controls = append(append([]Control(nil), ec.Controls...), controls...)
	return ec
}

// WithOOB returns a copy of the ErrorContext configured for HTMX out-of-band
// swap rendering. Set target to the CSS selector and swap to the hx-swap-oob
// strategy (e.g., "innerHTML").
func (ec ErrorContext) WithOOB(target, swap string) ErrorContext {
	ec.OOBTarget = target
	ec.OOBSwap = swap
	return ec
}

// HTTPError is a Go error that wraps an ErrorContext. Return it from handlers
// and let error-handling middleware detect it via errors.As to render the full
// hypermedia error response with action controls.
type HTTPError struct {
	// EC is the error context carrying the status code, message, and controls.
	EC ErrorContext
}

func (e *HTTPError) Error() string {
	if e.EC.Err != nil {
		return fmt.Sprintf("HTTP %d: %s: %v", e.EC.StatusCode, e.EC.Message, e.EC.Err)
	}
	return fmt.Sprintf("HTTP %d: %s", e.EC.StatusCode, e.EC.Message)
}

// NewHTTPError wraps an ErrorContext in an HTTPError suitable for returning
// from a handler. The error handler middleware intercepts it and renders the
// appropriate hypermedia error response.
func NewHTTPError(ec ErrorContext) *HTTPError {
	return &HTTPError{EC: ec}
}
