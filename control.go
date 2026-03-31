// Package hypermedia provides types and helpers for HATEOAS-style hypermedia controls
// embedded in HTMX responses. This package has no imports from this project.
package linkwell

// ControlKind identifies the rendering strategy for a Control.
type ControlKind string

const (
	// ControlKindRetry renders an HTMX button that re-issues a request.
	ControlKindRetry ControlKind = "retry"
	// ControlKindLink renders a plain anchor element.
	ControlKindLink ControlKind = "link"
	// ControlKindHTMX renders an arbitrary HTMX button using attribute spread.
	ControlKindHTMX ControlKind = "htmx"
	// ControlKindDismiss renders a HyperScript close button.
	ControlKindDismiss ControlKind = "dismiss"
	// ControlKindBack renders a browser history.back() button.
	ControlKindBack ControlKind = "back"
	// ControlKindHome renders a "Go Home" navigation button.
	ControlKindHome ControlKind = "home"
	// ControlKindReport renders an HTMX button that posts a report and triggers an alert.
	ControlKindReport ControlKind = "report"
)

// ControlVariant determines the visual emphasis of a Control.
// The zero value ("") renders as secondary (safe default).
type ControlVariant string

const (
	// VariantPrimary is the filled call-to-action style.
	VariantPrimary ControlVariant = "primary"
	// VariantSecondary is the supporting action style (default/zero value).
	VariantSecondary ControlVariant = "secondary"
	// VariantDanger is used for destructive actions.
	VariantDanger ControlVariant = "danger"
	// VariantGhost is a low-emphasis style.
	VariantGhost ControlVariant = "ghost"
	// VariantLink renders with a text-only link appearance.
	VariantLink ControlVariant = "link"
)

// SwapMode is a typed HTMX swap strategy for Control.Swap.
// Defined separately from htmx.SwapStrategy to keep this package zero-dependency.
type SwapMode string

// HTMX swap strategy constants.
const (
	SwapInnerHTML   SwapMode = "innerHTML"
	SwapOuterHTML   SwapMode = "outerHTML"
	SwapNone        SwapMode = "none"
	SwapBeforeBegin SwapMode = "beforebegin"
	SwapAfterBegin  SwapMode = "afterbegin"
	SwapBeforeEnd   SwapMode = "beforeend"
	SwapAfterEnd    SwapMode = "afterend"
	SwapDelete      SwapMode = "delete"
)

// Icon is a typed icon name for Control.Icon.
// Custom values are still valid via Icon("my-icon").
type Icon string

// Built-in icon name constants.
const (
	IconPencilSquare Icon = "pencil-square"
	IconXMark        Icon = "x-mark"
	IconCheck        Icon = "check"
	IconHome         Icon = "home"
)

// Include selector constants for HxRequestConfig.Include.
const (
	IncludeClosestTR   = "closest tr"
	IncludeClosestForm = "closest form"
)

// Default labels used by pattern factories. Override by building Controls directly.
const (
	LabelEdit           = "Edit"
	LabelDelete         = "Delete"
	LabelDeleteSelected = "Delete Selected"
	LabelSave           = "Save"
	LabelCancel         = "Cancel"
	LabelDetails        = "Details"
	LabelActivate       = "Activate"
	LabelDeactivate     = "Deactivate"
	LabelGoBack         = "Go Back"
	LabelGoHome         = "Go Home"
	LabelRetry          = "Retry"
	LabelDismiss        = "Close"
	LabelLogIn          = "Log In"
	LabelReportIssue    = "Report Issue"
)

// Default confirm messages used by pattern factories.
const (
	ConfirmDeleteSelected = "Delete selected items?"
)

// Default HTMX target selectors used by pattern factories.
const (
	TargetBody = "body"
)

// HxMethod is the HTTP verb used in an HTMX request attribute (hx-get, hx-post, etc.).
type HxMethod string

// HTTP verb constants for HTMX request attributes.
const (
	HxMethodGet    HxMethod = "get"
	HxMethodPost   HxMethod = "post"
	HxMethodPut    HxMethod = "put"
	HxMethodPatch  HxMethod = "patch"
	HxMethodDelete HxMethod = "delete"
)

// HxRequestConfig describes the HTMX request attributes for a control.
type HxRequestConfig struct {
	Method  HxMethod
	URL     string
	Target  string
	Include string
	Vals    string // JSON-encoded hx-vals for form data (e.g. `{"status":"active"}`)
}

// Attrs converts the config to a map[string]string for interop with NavItem,
// FilterField, or other consumers that use generic attribute maps.
func (r HxRequestConfig) Attrs() map[string]string {
	m := make(map[string]string, 4)
	if r.URL != "" {
		m[string(r.Method)] = r.URL
	}
	if r.Target != "" {
		m["target"] = r.Target
	}
	if r.Include != "" {
		m["include"] = r.Include
	}
	if r.Vals != "" {
		m["vals"] = r.Vals
	}
	return m
}

// HxGet returns a GET request config.
func HxGet(url, target string) HxRequestConfig {
	return HxRequestConfig{Method: HxMethodGet, URL: url, Target: target}
}

// HxPost returns a POST request config.
func HxPost(url, target string) HxRequestConfig {
	return HxRequestConfig{Method: HxMethodPost, URL: url, Target: target}
}

// HxPut returns a PUT request config.
func HxPut(url, target string) HxRequestConfig {
	return HxRequestConfig{Method: HxMethodPut, URL: url, Target: target}
}

// HxPatch returns a PATCH request config.
func HxPatch(url, target string) HxRequestConfig {
	return HxRequestConfig{Method: HxMethodPatch, URL: url, Target: target}
}

// HxDelete returns a DELETE request config.
func HxDelete(url, target string) HxRequestConfig {
	return HxRequestConfig{Method: HxMethodDelete, URL: url, Target: target}
}

// Control is a pure-data descriptor for a single hypermedia affordance.
// Templ components consume these to render the appropriate HTML element.
type Control struct {
	HxRequest   HxRequestConfig
	Kind        ControlKind
	Label       string
	Href        string
	Variant     ControlVariant
	Confirm     string
	Icon        Icon
	PushURL     string
	Swap        SwapMode
	Disabled    bool
	ErrorTarget string
	ModalID     string
}

// RetryButton creates an HTMX button that re-issues a request using the given method.
// method must be one of: "get", "post", "put", "delete", "patch".
// target is the CSS selector for hx-target.
func RetryButton(label string, method HxMethod, url, target string) Control {
	return Control{
		Kind:      ControlKindRetry,
		Label:     label,
		Variant:   VariantPrimary,
		HxRequest: HxRequestConfig{Method: method, URL: url, Target: target},
	}
}

// ConfirmAction creates a danger-variant HTMX button with an hx-confirm gate.
// Use for destructive operations that require user confirmation before proceeding.
func ConfirmAction(label string, method HxMethod, url, target, confirmMsg string) Control {
	return Control{
		Kind:      ControlKindHTMX,
		Label:     label,
		Variant:   VariantDanger,
		Confirm:   confirmMsg,
		HxRequest: HxRequestConfig{Method: method, URL: url, Target: target},
	}
}

// BackButton creates a browser history.back() control rendered via HyperScript.
// No server round-trip occurs.
func BackButton(label string) Control {
	return Control{
		Kind:  ControlKindBack,
		Label: label,
	}
}

// GoHomeButton creates a "Go Home" navigation control that pushes homeURL to browser history.
func GoHomeButton(label, homeURL, target string) Control {
	return Control{
		Kind:      ControlKindHome,
		Label:     label,
		Href:      homeURL,
		PushURL:   homeURL,
		HxRequest: HxGet(homeURL, target),
	}
}

// RedirectLink creates a same-tab anchor control.
func RedirectLink(label, href string) Control {
	return Control{
		Kind:  ControlKindLink,
		Label: label,
		Href:  href,
	}
}

// HTMXAction creates an arbitrary HTMX button using the supplied request config.
func HTMXAction(label string, req HxRequestConfig) Control {
	return Control{
		Kind:      ControlKindHTMX,
		Label:     label,
		HxRequest: req,
	}
}

// ReportIssueButton creates a button that fetches the Report Issue modal via
// HTMX GET. The server returns the modal fragment with collected log data; the
// modal auto-opens on load.
func ReportIssueButton(label, requestID string) Control {
	url := "/report-issue"
	if requestID != "" {
		url += "/" + requestID
	}
	return Control{
		Kind:      ControlKindReport,
		Label:     label,
		HxRequest: HxGet(url, "#report-modal-container"),
	}
}

// DismissButton creates a HyperScript-powered dismiss control.
func DismissButton(label string) Control {
	return Control{
		Kind:  ControlKindDismiss,
		Label: label,
	}
}

// WithSwap returns a copy with the given swap mode.
func (c Control) WithSwap(s SwapMode) Control {
	c.Swap = s
	return c
}

// WithVariant returns a copy with the given variant.
func (c Control) WithVariant(v ControlVariant) Control {
	c.Variant = v
	return c
}

// WithConfirm returns a copy with the given confirmation message.
func (c Control) WithConfirm(msg string) Control {
	c.Confirm = msg
	return c
}

// WithIcon returns a copy with the given icon.
func (c Control) WithIcon(i Icon) Control {
	c.Icon = i
	return c
}

// WithDisabled returns a copy with the disabled flag set.
func (c Control) WithDisabled(d bool) Control {
	c.Disabled = d
	return c
}

// WithErrorTarget returns a copy with the given hx-target-error selector.
// This overrides the parent hx-target-error so error responses render inline.
func (c Control) WithErrorTarget(target string) Control {
	c.ErrorTarget = target
	return c
}

// WithInclude returns a copy with the given include selector.
func (r HxRequestConfig) WithInclude(selector string) HxRequestConfig {
	r.Include = selector
	return r
}
