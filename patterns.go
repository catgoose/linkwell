package linkwell

// ErrorControlOpts carries context for building status-code-specific control
// sets. Zero values are safe: missing fields cause the corresponding control to
// be omitted from the result. For example, omitting RetryURL skips the Retry
// button; omitting HomeURL skips the GoHome button.
type ErrorControlOpts struct {
	// RetryMethod is the HTTP method for the retry button. Defaults to GET if empty.
	RetryMethod HxMethod
	// RetryURL is the request URL to retry (typically the current request URL).
	RetryURL string
	// RetryTarget is the CSS selector for hx-target on the retry button.
	RetryTarget string
	// HomeURL is the home page URL for the GoHome button (typically "/").
	HomeURL string
	// LoginURL is the login page URL for the Log In button (typically "/login").
	LoginURL string
}

func resolveRetryMethod(opts ErrorControlOpts) HxMethod {
	if opts.RetryMethod == "" {
		return HxMethodGet
	}
	return opts.RetryMethod
}

// NotFoundControls returns [Back] and optionally [GoHome] controls for a 404
// response. Pass an empty homeURL to omit the GoHome button.
func NotFoundControls(homeURL string) []Control {
	controls := []Control{BackButton(LabelGoBack)}
	if homeURL != "" {
		controls = append(controls, GoHomeButton(LabelGoHome, homeURL, TargetBody))
	}
	return controls
}

// ServiceErrorControls returns [Retry, Dismiss] controls for a 503 response.
// The Retry button is omitted if opts.RetryURL is empty.
func ServiceErrorControls(opts ErrorControlOpts) []Control {
	controls := []Control{}
	if opts.RetryURL != "" {
		controls = append(controls, RetryButton(LabelRetry, resolveRetryMethod(opts), opts.RetryURL, opts.RetryTarget))
	}
	controls = append(controls, DismissButton(LabelDismiss))
	return controls
}

// UnauthorizedControls returns [Log In, Dismiss] controls for a 401 response.
// The Log In button is omitted if loginURL is empty.
func UnauthorizedControls(loginURL string) []Control {
	controls := []Control{}
	if loginURL != "" {
		controls = append(controls, RedirectLink(LabelLogIn, loginURL))
	}
	controls = append(controls, DismissButton(LabelDismiss))
	return controls
}

// ForbiddenControls returns [Back, Dismiss] controls for a 403 response.
func ForbiddenControls() []Control {
	return []Control{
		BackButton(LabelGoBack),
		DismissButton(LabelDismiss),
	}
}

// InternalErrorControls returns [Retry, Dismiss] controls for a 500 response.
// The Retry button is omitted if opts.RetryURL is empty.
func InternalErrorControls(opts ErrorControlOpts) []Control {
	controls := []Control{}
	if opts.RetryURL != "" {
		controls = append(controls, RetryButton(LabelRetry, resolveRetryMethod(opts), opts.RetryURL, opts.RetryTarget))
	}
	controls = append(controls, DismissButton(LabelDismiss))
	return controls
}

// ErrorControlsForStatus dispatches to the appropriate control builder based on
// HTTP status code. Returns a [Dismiss] control for unrecognized status codes.
// Use in generic error-handling middleware.
func ErrorControlsForStatus(statusCode int, opts ErrorControlOpts) []Control {
	switch statusCode {
	case 404:
		return NotFoundControls(opts.HomeURL)
	case 401:
		return UnauthorizedControls(opts.LoginURL)
	case 403:
		return ForbiddenControls()
	case 503:
		return ServiceErrorControls(opts)
	case 500:
		return InternalErrorControls(opts)
	default:
		return []Control{DismissButton(LabelDismiss)}
	}
}

// RowActionCfg configures Edit + Delete controls for a table row. Both the Edit
// and Delete actions target the same row element (RowTarget) with outerHTML swap.
type RowActionCfg struct {
	EditURL     string // GET URL to fetch the edit form for this row.
	DeleteURL   string // DELETE URL to remove this row.
	RowTarget   string // CSS selector for the row element (e.g., "#row-42").
	ConfirmMsg  string // hx-confirm message for the delete action.
	ErrorTarget string // CSS selector for inline error display.
}

// TableRowActionCfg configures Edit + Delete controls where Edit swaps the
// individual row (RowTarget) and Delete replaces the entire table
// (TableTarget). Use when deleting a row requires re-rendering the full table
// (e.g., to update row numbers or totals).
type TableRowActionCfg struct {
	EditURL     string // GET URL to fetch the edit form for this row.
	DeleteURL   string // DELETE URL to remove this row.
	RowTarget   string // CSS selector for the row element.
	TableTarget string // CSS selector for the table container.
	ConfirmMsg  string // hx-confirm message for the delete action.
	ErrorTarget string // CSS selector for inline error display.
}

// RowFormActionCfg configures Save + Cancel controls for inline table row
// editing. RowFormActions uses PUT for existing rows; NewRowFormActions uses
// POST for new rows.
type RowFormActionCfg struct {
	SaveURL      string // PUT or POST URL for saving the row.
	CancelURL    string // GET URL to restore the display row.
	SaveTarget   string // CSS selector for the save response target.
	CancelTarget string // CSS selector for the cancel response target.
	ErrorTarget  string // CSS selector for inline error display.
}

// ResourceActionCfg configures Edit + Delete controls for resource detail
// pages (as opposed to table rows). Both actions target the same content area.
type ResourceActionCfg struct {
	EditURL     string // GET URL to fetch the edit form.
	DeleteURL   string // DELETE URL to remove the resource.
	ConfirmMsg  string // hx-confirm message for the delete action.
	Target      string // CSS selector for the content area.
	ErrorTarget string // CSS selector for inline error display.
}

// BulkActionCfg configures toolbar controls for batch operations on
// checkbox-selected table rows. Each URL is optional — omit to hide that
// action. The CheckboxSelector is the CSS selector for row checkboxes included
// via hx-include.
type BulkActionCfg struct {
	DeleteURL        string // DELETE URL for bulk deletion.
	ActivateURL      string // PUT URL for bulk activation.
	DeactivateURL    string // PUT URL for bulk deactivation.
	TableTarget      string // CSS selector for the table to replace after the operation.
	CheckboxSelector string // CSS selector for row checkboxes (e.g., ".user-checkbox").
	ErrorTarget      string // CSS selector for inline error display.
}

// ResourceActions returns Edit + Delete controls for resource detail pages.
// Controls are conditionally included: omit EditURL to hide Edit, omit
// DeleteURL to hide Delete. Returns an empty slice if both URLs are empty.
func ResourceActions(cfg ResourceActionCfg) []Control {
	controls := []Control{}
	if cfg.EditURL != "" {
		ctrl := Control{
			Kind:        ControlKindHTMX,
			Label:       LabelEdit,
			Variant:     VariantSecondary,
			Icon:        IconPencilSquare,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxGet(cfg.EditURL, cfg.Target),
		}
		controls = append(controls, ctrl)
	}
	if cfg.DeleteURL != "" {
		ctrl := ConfirmAction(LabelDelete, HxMethodDelete, cfg.DeleteURL, cfg.Target, cfg.ConfirmMsg)
		ctrl.ErrorTarget = cfg.ErrorTarget
		controls = append(controls, ctrl)
	}
	return controls
}

// FormActions returns [Save, Cancel] controls for form footers. The Save button
// uses hx-include="closest form" to submit the form data; the parent form
// element must carry the hx-post or hx-put attribute that drives submission.
// The Cancel button is a plain link to cancelHref.
func FormActions(cancelHref string) []Control {
	return []Control{
		{
			Kind:      ControlKindHTMX,
			Label:     LabelSave,
			Variant:   VariantPrimary,
			Icon:      IconCheck,
			HxRequest: HxRequestConfig{Include: IncludeClosestForm},
		},
		{
			Kind:    ControlKindLink,
			Label:   LabelCancel,
			Href:    cancelHref,
			Icon:    IconXMark,
			Variant: VariantGhost,
		},
	}
}

// RowActions returns Edit + Delete controls for a table row where both actions
// swap the same row target with outerHTML. Controls are conditionally included:
// omit EditURL to hide Edit, omit DeleteURL to hide Delete.
func RowActions(cfg RowActionCfg) []Control {
	controls := []Control{}
	if cfg.EditURL != "" {
		controls = append(controls, Control{
			Kind:        ControlKindHTMX,
			Label:       LabelEdit,
			Variant:     VariantSecondary,
			Icon:        IconPencilSquare,
			Swap:        SwapOuterHTML,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxGet(cfg.EditURL, cfg.RowTarget),
		})
	}
	if cfg.DeleteURL != "" {
		controls = append(controls, Control{
			Kind:        ControlKindHTMX,
			Label:       LabelDelete,
			Variant:     VariantDanger,
			Confirm:     cfg.ConfirmMsg,
			Swap:        SwapOuterHTML,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxDelete(cfg.DeleteURL, cfg.RowTarget),
		})
	}
	return controls
}

// TableRowActions returns Edit + Delete controls where Edit swaps the row and
// Delete replaces the entire table container. Controls are conditionally
// included: omit EditURL to hide Edit, omit DeleteURL to hide Delete.
func TableRowActions(cfg TableRowActionCfg) []Control {
	controls := []Control{}
	if cfg.EditURL != "" {
		controls = append(controls, Control{
			Kind:        ControlKindHTMX,
			Label:       LabelEdit,
			Variant:     VariantSecondary,
			Icon:        IconPencilSquare,
			Swap:        SwapOuterHTML,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxGet(cfg.EditURL, cfg.RowTarget),
		})
	}
	if cfg.DeleteURL != "" {
		controls = append(controls, Control{
			Kind:        ControlKindHTMX,
			Label:       LabelDelete,
			Variant:     VariantDanger,
			Confirm:     cfg.ConfirmMsg,
			Swap:        SwapOuterHTML,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxDelete(cfg.DeleteURL, cfg.TableTarget),
		})
	}
	return controls
}

// RowFormActions returns Save + Cancel controls for an inline edit row. Save
// uses hx-put with hx-include="closest tr" to submit the row's form fields.
// Controls are conditionally included: omit SaveURL to hide Save, omit
// CancelURL to hide Cancel.
func RowFormActions(cfg RowFormActionCfg) []Control {
	controls := []Control{}
	if cfg.SaveURL != "" {
		controls = append(controls, Control{
			Kind:        ControlKindHTMX,
			Label:       LabelSave,
			Variant:     VariantPrimary,
			Icon:        IconCheck,
			Swap:        SwapOuterHTML,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxRequestConfig{Method: HxMethodPut, URL: cfg.SaveURL, Target: cfg.SaveTarget, Include: IncludeClosestTR},
		})
	}
	if cfg.CancelURL != "" {
		controls = append(controls, Control{
			Kind:        ControlKindHTMX,
			Label:       LabelCancel,
			Variant:     VariantGhost,
			Icon:        IconXMark,
			Swap:        SwapOuterHTML,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxGet(cfg.CancelURL, cfg.CancelTarget),
		})
	}
	return controls
}

// NewRowFormActions returns Save + Cancel controls for a new-item inline form
// row. Save uses hx-post (instead of hx-put) with hx-include="closest tr".
// Controls are conditionally included: omit SaveURL to hide Save, omit
// CancelURL to hide Cancel.
func NewRowFormActions(cfg RowFormActionCfg) []Control {
	controls := []Control{}
	if cfg.SaveURL != "" {
		controls = append(controls, Control{
			Kind:        ControlKindHTMX,
			Label:       LabelSave,
			Variant:     VariantPrimary,
			Icon:        IconCheck,
			Swap:        SwapOuterHTML,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxRequestConfig{Method: HxMethodPost, URL: cfg.SaveURL, Target: cfg.SaveTarget, Include: IncludeClosestTR},
		})
	}
	if cfg.CancelURL != "" {
		controls = append(controls, Control{
			Kind:        ControlKindHTMX,
			Label:       LabelCancel,
			Variant:     VariantGhost,
			Icon:        IconXMark,
			Swap:        SwapOuterHTML,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxGet(cfg.CancelURL, cfg.CancelTarget),
		})
	}
	return controls
}

// EmptyStateAction returns a single primary call-to-action control for empty
// list states (e.g., "Create First User" when a table has no rows).
func EmptyStateAction(label, createURL, target string) Control {
	return Control{
		Kind:      ControlKindHTMX,
		Label:     label,
		Variant:   VariantPrimary,
		HxRequest: HxGet(createURL, target),
	}
}

// CatalogRowAction returns a ghost-variant Details button that fills an
// adjacent placeholder row via innerHTML swap. Use for catalog/listing views
// where clicking a row loads detail content into an expandable area below it.
func CatalogRowAction(detailURL, detailRowTarget string) Control {
	return Control{
		Kind:      ControlKindHTMX,
		Label:     LabelDetails,
		Variant:   VariantGhost,
		Swap:      SwapInnerHTML,
		HxRequest: HxGet(detailURL, detailRowTarget),
	}
}

// BulkActions returns toolbar controls for batch operations on
// checkbox-selected rows. Controls are conditionally included: omit DeleteURL,
// ActivateURL, or DeactivateURL to hide the corresponding action.
func BulkActions(cfg BulkActionCfg) []Control {
	controls := []Control{}
	if cfg.DeleteURL != "" {
		controls = append(controls, Control{
			Kind:        ControlKindHTMX,
			Label:       LabelDeleteSelected,
			Variant:     VariantDanger,
			Confirm:     ConfirmDeleteSelected,
			Swap:        SwapOuterHTML,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxRequestConfig{Method: HxMethodDelete, URL: cfg.DeleteURL, Target: cfg.TableTarget, Include: cfg.CheckboxSelector},
		})
	}
	if cfg.ActivateURL != "" {
		controls = append(controls, Control{
			Kind:        ControlKindHTMX,
			Label:       LabelActivate,
			Variant:     VariantSecondary,
			Swap:        SwapOuterHTML,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxRequestConfig{Method: HxMethodPut, URL: cfg.ActivateURL, Target: cfg.TableTarget, Include: cfg.CheckboxSelector},
		})
	}
	if cfg.DeactivateURL != "" {
		controls = append(controls, Control{
			Kind:        ControlKindHTMX,
			Label:       LabelDeactivate,
			Variant:     VariantGhost,
			Swap:        SwapOuterHTML,
			ErrorTarget: cfg.ErrorTarget,
			HxRequest:   HxRequestConfig{Method: HxMethodPut, URL: cfg.DeactivateURL, Target: cfg.TableTarget, Include: cfg.CheckboxSelector},
		})
	}
	return controls
}
