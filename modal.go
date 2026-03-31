package linkwell

// ModalButtonRole identifies the semantic role of a modal button.
type ModalButtonRole string

const (
	// ModalRolePrimary is the main action button (Submit, Save, Yes, OK, Delete).
	ModalRolePrimary ModalButtonRole = "primary"
	// ModalRoleCancel dismisses the modal without action.
	ModalRoleCancel ModalButtonRole = "cancel"
	// ModalRoleSecondary is an additional action (Reset, etc.).
	ModalRoleSecondary ModalButtonRole = "secondary"
)

// ModalButton describes a single button inside a modal.
type ModalButton struct {
	Label   string
	Role    ModalButtonRole
	Variant ControlVariant
}

// ModalButtonSet is an ordered list of buttons for a modal footer.
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

// ModalConfig describes everything needed to render a modal.
type ModalConfig struct {
	ID      string
	Title   string
	Buttons ModalButtonSet
	// HxPost is the URL for the primary button's hx-post attribute.
	// When set, the primary button submits the modal's form via HTMX.
	HxPost string
	// HxTarget is the hx-target selector for the primary button's HTMX request.
	HxTarget string
	// HxSwap is the hx-swap strategy for the primary button's HTMX request.
	HxSwap SwapMode
}

// ReportIssueModal returns a ModalConfig for the Report Issue flow.
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
