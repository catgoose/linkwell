// Package linkwell provides types and helpers for HATEOAS-style hypermedia
// controls, link relations (RFC 8288), and navigation primitives. All types are
// pure data descriptors — they carry no rendering logic and can be consumed by
// any template engine (templ, html/template, etc.) or serialized to JSON.
package linkwell

// ControlKind identifies the rendering strategy a template should use for a
// Control. Each kind implies a specific HTML element and behavior pattern.
type ControlKind string

const (
	// ControlKindRetry renders an HTMX button that re-issues a failed request.
	ControlKindRetry ControlKind = "retry"
	// ControlKindLink renders a plain HTML anchor (<a>) element.
	ControlKindLink ControlKind = "link"
	// ControlKindHTMX renders a button with HTMX attributes spread from HxRequest.
	ControlKindHTMX ControlKind = "htmx"
	// ControlKindDismiss renders a close button powered by HyperScript.
	ControlKindDismiss ControlKind = "dismiss"
	// ControlKindBack renders a button that calls browser history.back().
	ControlKindBack ControlKind = "back"
	// ControlKindHome renders a navigation button that redirects to the home page.
	ControlKindHome ControlKind = "home"
	// ControlKindReport renders an HTMX button that fetches a report modal.
	ControlKindReport ControlKind = "report"
)

// ControlVariant determines the visual emphasis of a Control. Templates map
// variants to CSS classes (e.g., DaisyUI btn-primary, btn-ghost). The zero
// value ("") renders as secondary, which is a safe default.
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

// SwapMode is a typed HTMX swap strategy for Control.Swap. These mirror the
// hx-swap attribute values. Defined as a separate type (rather than importing
// an htmx package) to keep linkwell dependency-free.
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

// Icon is a typed icon name for Control.Icon. The built-in constants cover
// common actions, but any string is valid — templates interpret the value to
// select the appropriate icon component (e.g., Heroicons, Lucide).
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

// HxMethod is the HTTP verb used in an HTMX request attribute. The value maps
// directly to the hx-get, hx-post, hx-put, hx-patch, or hx-delete attribute name.
type HxMethod string

// HTTP verb constants for HTMX request attributes.
const (
	HxMethodGet    HxMethod = "get"
	HxMethodPost   HxMethod = "post"
	HxMethodPut    HxMethod = "put"
	HxMethodPatch  HxMethod = "patch"
	HxMethodDelete HxMethod = "delete"
)

// HxRequestConfig describes the HTMX request attributes for a control. It maps
// to the hx-{method}, hx-target, hx-include, and hx-vals attributes. Use the
// HxGet, HxPost, HxPut, HxPatch, and HxDelete helpers for common cases, or
// build directly for full control.
type HxRequestConfig struct {
	// Method is the HTTP verb (maps to hx-get, hx-post, etc.).
	Method HxMethod
	// URL is the request endpoint.
	URL string
	// Target is the CSS selector for hx-target.
	Target string
	// Include is the CSS selector for hx-include (e.g., "closest form").
	Include string
	// Vals is a JSON-encoded string for hx-vals (e.g., `{"status":"active"}`).
	Vals string
}

// Attrs converts the config to a map[string]string suitable for attribute
// spreading in templates. Keys are the HTMX attribute suffixes (e.g., "get",
// "target", "include", "vals"). Used for interop with NavItem.HTMXAttrs and
// other consumers that accept generic attribute maps.
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

// Control is a pure-data descriptor for a single hypermedia affordance — a
// button, link, or action that a user can take. Templates consume controls to
// render the appropriate HTML elements with the correct HTMX attributes,
// confirmation dialogs, icons, and visual styles. Controls are value types;
// use the With* methods to derive modified copies.
type Control struct {
	// HxRequest carries the HTMX request attributes (method, URL, target, etc.).
	HxRequest HxRequestConfig
	// Kind determines how the template renders this control (button, link, etc.).
	Kind ControlKind
	// Label is the user-visible text for the control.
	Label string
	// Href is the URL for link-type controls (ControlKindLink, ControlKindHome).
	Href string
	// Variant sets the visual emphasis (primary, danger, ghost, etc.).
	Variant ControlVariant
	// Confirm is an optional hx-confirm message shown before the action executes.
	Confirm string
	// Icon is an optional icon name rendered alongside the label.
	Icon Icon
	// PushURL is set on navigation controls to update the browser URL via hx-push-url.
	PushURL string
	// Swap overrides the default hx-swap strategy for this control.
	Swap SwapMode
	// Disabled renders the control in a non-interactive state.
	Disabled bool
	// ErrorTarget overrides the parent hx-target-error so error responses render
	// inline near this control rather than in a global error container.
	ErrorTarget string
	// ModalID ties this control to a specific modal dialog.
	ModalID string
}

// RetryButton creates a primary-variant HTMX button that re-issues a failed
// request. Typically used in error states to let the user retry the operation
// that produced the error.
func RetryButton(label string, method HxMethod, url, target string) Control {
	return Control{
		Kind:      ControlKindRetry,
		Label:     label,
		Variant:   VariantPrimary,
		HxRequest: HxRequestConfig{Method: method, URL: url, Target: target},
	}
}

// ConfirmAction creates a danger-variant HTMX button with an hx-confirm
// confirmation dialog. Use for destructive operations (delete, archive) where
// the user should explicitly confirm before the request is sent.
func ConfirmAction(label string, method HxMethod, url, target, confirmMsg string) Control {
	return Control{
		Kind:      ControlKindHTMX,
		Label:     label,
		Variant:   VariantDanger,
		Confirm:   confirmMsg,
		HxRequest: HxRequestConfig{Method: method, URL: url, Target: target},
	}
}

// BackButton creates a client-side browser history.back() control. No server
// round-trip occurs — the template renders this as a HyperScript-powered button.
func BackButton(label string) Control {
	return Control{
		Kind:  ControlKindBack,
		Label: label,
	}
}

// GoHomeButton creates a navigation control that loads the home page via HTMX
// GET and pushes homeURL to browser history via hx-push-url.
func GoHomeButton(label, homeURL, target string) Control {
	return Control{
		Kind:      ControlKindHome,
		Label:     label,
		Href:      homeURL,
		PushURL:   homeURL,
		HxRequest: HxGet(homeURL, target),
	}
}

// RedirectLink creates a plain anchor (<a>) control for same-tab navigation.
func RedirectLink(label, href string) Control {
	return Control{
		Kind:  ControlKindLink,
		Label: label,
		Href:  href,
	}
}

// HTMXAction creates an HTMX button with the supplied request config. Use when
// the preset factory functions do not fit — this gives full control over the
// HTMX attributes while still producing a typed Control.
func HTMXAction(label string, req HxRequestConfig) Control {
	return Control{
		Kind:      ControlKindHTMX,
		Label:     label,
		HxRequest: req,
	}
}

// ReportIssueButton creates an HTMX button that fetches the Report Issue modal
// via GET to /report-issue (or /report-issue/{requestID} if provided). The
// server returns the modal fragment which auto-opens on load.
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

// DismissButton creates a HyperScript-powered close/dismiss control. Typically
// used to close error banners, notifications, or alert panels.
func DismissButton(label string) Control {
	return Control{
		Kind:  ControlKindDismiss,
		Label: label,
	}
}

// WithSwap returns a copy of the control with the given HTMX swap strategy.
func (c Control) WithSwap(s SwapMode) Control {
	c.Swap = s
	return c
}

// WithVariant returns a copy of the control with the given visual variant.
func (c Control) WithVariant(v ControlVariant) Control {
	c.Variant = v
	return c
}

// WithConfirm returns a copy of the control with the given hx-confirm message.
func (c Control) WithConfirm(msg string) Control {
	c.Confirm = msg
	return c
}

// WithIcon returns a copy of the control with the given icon name.
func (c Control) WithIcon(i Icon) Control {
	c.Icon = i
	return c
}

// WithDisabled returns a copy of the control with the given disabled state.
func (c Control) WithDisabled(d bool) Control {
	c.Disabled = d
	return c
}

// WithErrorTarget returns a copy of the control with the given hx-target-error
// CSS selector, overriding any parent error target so error responses render
// inline near this control.
func (c Control) WithErrorTarget(target string) Control {
	c.ErrorTarget = target
	return c
}

// WithInclude returns a copy of the request config with the given hx-include
// CSS selector.
func (r HxRequestConfig) WithInclude(selector string) HxRequestConfig {
	r.Include = selector
	return r
}
