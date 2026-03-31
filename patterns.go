package linkwell

// ErrorControlOpts carries context for building status-code-specific molecules.
// Zero values are safe: missing fields omit the corresponding control.
type ErrorControlOpts struct {
	RetryMethod HxMethod // defaults to HxMethodGet if zero
	RetryURL    string   // typically c.Request().URL.String()
	RetryTarget string   // CSS selector for hx-target on retry
	HomeURL     string   // for GoHome control; typically "/"
	LoginURL    string   // for Unauthorized; typically "/login"
}

func resolveRetryMethod(opts ErrorControlOpts) HxMethod {
	if opts.RetryMethod == "" {
		return HxMethodGet
	}
	return opts.RetryMethod
}

// NotFoundControls returns [Back] and optionally [GoHome] controls for a 404 error.
func NotFoundControls(homeURL string) []Control {
	controls := []Control{BackButton(LabelGoBack)}
	if homeURL != "" {
		controls = append(controls, GoHomeButton(LabelGoHome, homeURL, TargetBody))
	}
	return controls
}

// ServiceErrorControls returns [Retry?] + [Dismiss] controls for a 503 error.
func ServiceErrorControls(opts ErrorControlOpts) []Control {
	controls := []Control{}
	if opts.RetryURL != "" {
		controls = append(controls, RetryButton(LabelRetry, resolveRetryMethod(opts), opts.RetryURL, opts.RetryTarget))
	}
	controls = append(controls, DismissButton(LabelDismiss))
	return controls
}

// UnauthorizedControls returns [Log In?] + [Dismiss] controls for a 401 error.
func UnauthorizedControls(loginURL string) []Control {
	controls := []Control{}
	if loginURL != "" {
		controls = append(controls, RedirectLink(LabelLogIn, loginURL))
	}
	controls = append(controls, DismissButton(LabelDismiss))
	return controls
}

// ForbiddenControls returns [Back] + [Dismiss] controls for a 403 error.
func ForbiddenControls() []Control {
	return []Control{
		BackButton(LabelGoBack),
		DismissButton(LabelDismiss),
	}
}

// InternalErrorControls returns [Retry?] + [Dismiss] controls for a 500 error.
func InternalErrorControls(opts ErrorControlOpts) []Control {
	controls := []Control{}
	if opts.RetryURL != "" {
		controls = append(controls, RetryButton(LabelRetry, resolveRetryMethod(opts), opts.RetryURL, opts.RetryTarget))
	}
	controls = append(controls, DismissButton(LabelDismiss))
	return controls
}

// ErrorControlsForStatus dispatches to the appropriate molecule by HTTP status code.
// Use in generic middleware or catch-all error handlers.
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

// RowActionCfg configures Edit + Delete controls for a table row.
type RowActionCfg struct {
	EditURL     string
	DeleteURL   string
	RowTarget   string
	ConfirmMsg  string
	ErrorTarget string
}

// TableRowActionCfg configures Edit (swap row) + Delete (replace table) controls.
type TableRowActionCfg struct {
	EditURL     string
	DeleteURL   string
	RowTarget   string
	TableTarget string
	ConfirmMsg  string
	ErrorTarget string
}

// RowFormActionCfg configures Save + Cancel controls for inline edit rows.
type RowFormActionCfg struct {
	SaveURL      string
	CancelURL    string
	SaveTarget   string
	CancelTarget string
	ErrorTarget  string
}

// ResourceActionCfg configures Edit + Delete controls for resource detail pages.
type ResourceActionCfg struct {
	EditURL     string
	DeleteURL   string
	ConfirmMsg  string
	Target      string
	ErrorTarget string
}

// BulkActionCfg configures batch operation controls for checkbox-selected rows.
type BulkActionCfg struct {
	DeleteURL        string
	ActivateURL      string
	DeactivateURL    string
	TableTarget      string
	CheckboxSelector string
	ErrorTarget      string
}

// ResourceActions returns Edit + Delete controls for resource detail pages.
// Controls are conditionally included: omit EditURL to hide Edit, omit DeleteURL
// to hide Delete. Returns an empty slice if both are empty.
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

// FormActions returns [Save (primary)] + [Cancel (ghost)] controls for form footers.
// The Save button uses hx-include="closest form"; the parent form must carry the
// hx-post or hx-put attribute that drives the submission.
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

// RowActions returns Edit + Delete controls for a table row.
// Controls are conditionally included: omit EditURL to hide Edit, omit DeleteURL
// to hide Delete. Returns an empty slice if both are empty.
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

// TableRowActions returns Edit (swap row outerHTML) + Delete (replace tableTarget outerHTML).
// Controls are conditionally included: omit EditURL to hide Edit, omit DeleteURL
// to hide Delete. Returns an empty slice if both are empty.
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

// RowFormActions returns Save (hx-put) + Cancel controls for an inline edit row.
// Controls are conditionally included: omit SaveURL to hide Save, omit CancelURL
// to hide Cancel. Returns an empty slice if both are empty.
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

// NewRowFormActions returns Save (hx-post) + Cancel controls for a new-item inline form row.
// Controls are conditionally included: omit SaveURL to hide Save, omit CancelURL
// to hide Cancel. Returns an empty slice if both are empty.
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

// EmptyStateAction returns a single primary CTA for empty list states.
func EmptyStateAction(label, createURL, target string) Control {
	return Control{
		Kind:      ControlKindHTMX,
		Label:     label,
		Variant:   VariantPrimary,
		HxRequest: HxGet(createURL, target),
	}
}

// CatalogRowAction returns a Details button that fills the adjacent placeholder row via innerHTML.
// detailRowTarget is the CSS selector for the placeholder <tr>, e.g. "#detail-row-42".
func CatalogRowAction(detailURL, detailRowTarget string) Control {
	return Control{
		Kind:      ControlKindHTMX,
		Label:     LabelDetails,
		Variant:   VariantGhost,
		Swap:      SwapInnerHTML,
		HxRequest: HxGet(detailURL, detailRowTarget),
	}
}

// BulkActions returns toolbar controls for batch operations on checkbox-selected rows.
// Controls are conditionally included: omit DeleteURL, ActivateURL, or DeactivateURL
// to hide the corresponding control. Returns an empty slice if all are empty.
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
