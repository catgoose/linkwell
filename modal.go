package linkwell

// ModalButtonRole identifies the semantic role of a button within a modal
// dialog. Templates use this to determine behavior: primary buttons submit,
// cancel buttons dismiss, and secondary buttons perform auxiliary actions.
type ModalButtonRole string

const (
	// ModalRolePrimary is the main action button (Submit, Save, Yes, OK, Delete).
	ModalRolePrimary ModalButtonRole = "primary"
	// ModalRoleCancel dismisses the modal without action.
	ModalRoleCancel ModalButtonRole = "cancel"
	// ModalRoleSecondary is an additional action (Reset, etc.).
	ModalRoleSecondary ModalButtonRole = "secondary"
)

// ModalButton describes a single button inside a modal dialog footer.
type ModalButton struct {
	// Label is the button text.
	Label string
	// Role determines the button's behavior (submit, cancel, auxiliary).
	Role ModalButtonRole
	// Variant sets the visual style of the button.
	Variant ControlVariant
}

// ModalButtonSet is an ordered list of buttons for a modal footer. Buttons
// render left-to-right in the order they appear in the slice. Use the preset
// variables (ModalOK, ModalYesNo, etc.) for common patterns.
type ModalButtonSet []ModalButton

// Preset modal button sets.
var (
	ModalOK = ModalButtonSet{
		{Label: "OK", Role: ModalRolePrimary, Variant: VariantPrimary},
	}
	ModalYesNo = ModalButtonSet{
		{Label: "No", Role: ModalRoleCancel, Variant: VariantGhost},
		{Label: "Yes", Role: ModalRolePrimary, Variant: VariantPrimary},
	}
	ModalSaveCancel = ModalButtonSet{
		{Label: LabelCancel, Role: ModalRoleCancel, Variant: VariantGhost},
		{Label: LabelSave, Role: ModalRolePrimary, Variant: VariantPrimary},
	}
	ModalSaveCancelReset = ModalButtonSet{
		{Label: "Reset", Role: ModalRoleSecondary, Variant: VariantSecondary},
		{Label: LabelCancel, Role: ModalRoleCancel, Variant: VariantGhost},
		{Label: LabelSave, Role: ModalRolePrimary, Variant: VariantPrimary},
	}
	ModalSubmitCancel = ModalButtonSet{
		{Label: LabelCancel, Role: ModalRoleCancel, Variant: VariantGhost},
		{Label: "Submit", Role: ModalRolePrimary, Variant: VariantPrimary},
	}
	ModalConfirmCancel = ModalButtonSet{
		{Label: LabelCancel, Role: ModalRoleCancel, Variant: VariantGhost},
		{Label: "Confirm", Role: ModalRolePrimary, Variant: VariantDanger},
	}
	ModalDeleteCancel = ModalButtonSet{
		{Label: LabelCancel, Role: ModalRoleCancel, Variant: VariantGhost},
		{Label: LabelDelete, Role: ModalRolePrimary, Variant: VariantDanger},
	}
)

// ModalConfig describes everything needed to render a modal dialog. The ID
// must be unique on the page. When HxPost is set, the primary button submits
// the modal's form content via HTMX POST.
type ModalConfig struct {
	// ID is the unique HTML id for the modal element.
	ID string
	// Title is displayed in the modal header.
	Title string
	// Buttons defines the footer button set.
	Buttons ModalButtonSet
	// HxPost is the URL for the primary button's hx-post attribute. When empty,
	// the primary button uses client-side behavior only (e.g., close on confirm).
	HxPost string
	// HxTarget is the hx-target CSS selector for the primary button's HTMX
	// response.
	HxTarget string
	// HxSwap is the hx-swap strategy for the primary button's HTMX response.
	HxSwap SwapMode
}

// ReportIssueModal returns a ModalConfig preconfigured for the Report Issue
// flow. The modal posts to /report-issue (or /report-issue/{requestID}) with
// SwapNone so the response triggers a toast/alert without replacing DOM content.
func ReportIssueModal(requestID string) ModalConfig {
	url := "/report-issue"
	if requestID != "" {
		url += "/" + requestID
	}
	return ModalConfig{
		ID:      "report-issue-modal",
		Title:   "Report Issue",
		Buttons: ModalSubmitCancel,
		HxPost:  url,
		HxSwap:  SwapNone,
	}
}
