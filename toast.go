package linkwell

// ToastVariant identifies the severity or intent of a Toast notification.
// Templates map variants to visual styles (e.g., DaisyUI alert-success).
type ToastVariant string

const (
	// ToastSuccess indicates a successful operation.
	ToastSuccess ToastVariant = "success"
	// ToastInfo provides neutral informational feedback.
	ToastInfo ToastVariant = "info"
	// ToastWarning signals a non-blocking issue that deserves attention.
	ToastWarning ToastVariant = "warning"
	// ToastError indicates a failed operation (complement to ErrorContext).
	ToastError ToastVariant = "error"
)

// Toast is a pure-data descriptor for a transient notification. It is the
// success/info/warning complement to ErrorContext: the server decides what
// feedback to show and the template renders the appropriate alert with optional
// action controls. Toasts are value types; use the With* methods to derive
// modified copies.
type Toast struct {
	// Message is the user-visible notification text.
	Message string
	// Variant determines the visual style (success, info, warning, error).
	Variant ToastVariant
	// Controls is an optional set of action affordances rendered with the toast
	// (e.g., an "Undo" button).
	Controls []Control
	// AutoDismiss is the number of seconds before the toast auto-closes.
	// Zero means the toast is sticky and must be dismissed manually.
	AutoDismiss int
	// OOBTarget is the CSS selector for an HTMX OOB swap target. When set,
	// the toast renders as an out-of-band fragment.
	OOBTarget string
	// OOBSwap is the hx-swap-oob strategy (e.g., "afterbegin").
	OOBSwap string
}

// SuccessToast creates a Toast with the success variant.
func SuccessToast(message string) Toast {
	return Toast{Message: message, Variant: ToastSuccess}
}

// InfoToast creates a Toast with the info variant.
func InfoToast(message string) Toast {
	return Toast{Message: message, Variant: ToastInfo}
}

// WarningToast creates a Toast with the warning variant.
func WarningToast(message string) Toast {
	return Toast{Message: message, Variant: ToastWarning}
}

// ErrorToast creates a Toast with the error variant.
func ErrorToast(message string) Toast {
	return Toast{Message: message, Variant: ToastError}
}

// WithControls returns a copy of the Toast with the given controls appended to
// the existing set. The original is not modified.
func (t Toast) WithControls(controls ...Control) Toast {
	t.Controls = append(append([]Control(nil), t.Controls...), controls...)
	return t
}

// WithAutoDismiss returns a copy of the Toast with the given auto-dismiss
// duration in seconds. Pass 0 for a sticky toast.
func (t Toast) WithAutoDismiss(seconds int) Toast {
	t.AutoDismiss = seconds
	return t
}

// WithOOB returns a copy of the Toast configured for HTMX out-of-band swap
// rendering. Set target to the CSS selector and swap to the hx-swap-oob
// strategy (e.g., "afterbegin").
func (t Toast) WithOOB(target, swap string) Toast {
	t.OOBTarget = target
	t.OOBSwap = swap
	return t
}
