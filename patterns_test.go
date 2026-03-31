package linkwell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NotFoundControls
// ---------------------------------------------------------------------------

func TestNotFoundControls_WithHomeURL(t *testing.T) {
	controls := NotFoundControls("/")
	require.Len(t, controls, 2)
	require.Equal(t, ControlKindBack, controls[0].Kind, "first control should be BackButton")
	require.Equal(t, LabelGoBack, controls[0].Label)
	require.Equal(t, ControlKindHome, controls[1].Kind, "second control should be GoHomeButton")
	require.Equal(t, LabelGoHome, controls[1].Label)
}

func TestNotFoundControls_EmptyHomeURL(t *testing.T) {
	controls := NotFoundControls("")
	require.Len(t, controls, 1)
	require.Equal(t, ControlKindBack, controls[0].Kind, "only control should be BackButton")
}

// ---------------------------------------------------------------------------
// ServiceErrorControls
// ---------------------------------------------------------------------------

func TestServiceErrorControls_WithRetryURL(t *testing.T) {
	opts := ErrorControlOpts{RetryURL: "/api/data", RetryTarget: "#content"}
	controls := ServiceErrorControls(opts)
	require.Len(t, controls, 2)
	require.Equal(t, ControlKindRetry, controls[0].Kind, "first control should be RetryButton")
	require.Equal(t, LabelRetry, controls[0].Label)
	require.Equal(t, ControlKindDismiss, controls[1].Kind, "second control should be DismissButton")
	require.Equal(t, LabelDismiss, controls[1].Label)
}

func TestServiceErrorControls_WithoutRetryURL(t *testing.T) {
	controls := ServiceErrorControls(ErrorControlOpts{})
	require.Len(t, controls, 1)
	require.Equal(t, ControlKindDismiss, controls[0].Kind, "only control should be DismissButton")
}

func TestServiceErrorControls_DefaultRetryMethodIsGet(t *testing.T) {
	opts := ErrorControlOpts{RetryURL: "/url", RetryTarget: "#t"}
	controls := ServiceErrorControls(opts)
	retry := controls[0]
	require.Equal(t, HxMethodGet, retry.HxRequest.Method, "default retry method should be 'get'")
}

// ---------------------------------------------------------------------------
// UnauthorizedControls
// ---------------------------------------------------------------------------

func TestUnauthorizedControls_WithLoginURL(t *testing.T) {
	controls := UnauthorizedControls("/login")
	require.Len(t, controls, 2)
	require.Equal(t, ControlKindLink, controls[0].Kind, "first control should be RedirectLink")
	require.Equal(t, LabelLogIn, controls[0].Label)
	require.Equal(t, "/login", controls[0].Href)
	require.Equal(t, ControlKindDismiss, controls[1].Kind, "second control should be DismissButton")
	require.Equal(t, LabelDismiss, controls[1].Label)
}

func TestUnauthorizedControls_EmptyLoginURL(t *testing.T) {
	controls := UnauthorizedControls("")
	require.Len(t, controls, 1)
	require.Equal(t, ControlKindDismiss, controls[0].Kind, "only control should be DismissButton")
}

// ---------------------------------------------------------------------------
// ForbiddenControls
// ---------------------------------------------------------------------------

func TestForbiddenControls_ReturnsTwoControls(t *testing.T) {
	controls := ForbiddenControls()
	require.Len(t, controls, 2)
	require.Equal(t, ControlKindBack, controls[0].Kind, "first control should be BackButton")
	require.Equal(t, LabelGoBack, controls[0].Label)
	require.Equal(t, ControlKindDismiss, controls[1].Kind, "second control should be DismissButton")
	require.Equal(t, LabelDismiss, controls[1].Label)
}

// ---------------------------------------------------------------------------
// InternalErrorControls
// ---------------------------------------------------------------------------

func TestInternalErrorControls_WithRetryURL(t *testing.T) {
	opts := ErrorControlOpts{RetryURL: "/api", RetryTarget: "#main"}
	controls := InternalErrorControls(opts)
	require.Len(t, controls, 2)
	require.Equal(t, ControlKindRetry, controls[0].Kind, "first control should be RetryButton")
	require.Equal(t, ControlKindDismiss, controls[1].Kind, "second control should be DismissButton")
}

func TestInternalErrorControls_WithoutRetryURL(t *testing.T) {
	controls := InternalErrorControls(ErrorControlOpts{})
	require.Len(t, controls, 1)
	require.Equal(t, ControlKindDismiss, controls[0].Kind, "only control should be DismissButton")
}

// ---------------------------------------------------------------------------
// ErrorControlsForStatus
// ---------------------------------------------------------------------------

func TestErrorControlsForStatus_404(t *testing.T) {
	controls := ErrorControlsForStatus(404, ErrorControlOpts{HomeURL: "/"})
	require.Equal(t, ControlKindBack, controls[0].Kind, "404: first control should be Back")
}

func TestErrorControlsForStatus_401(t *testing.T) {
	controls := ErrorControlsForStatus(401, ErrorControlOpts{LoginURL: "/login"})
	require.Equal(t, ControlKindLink, controls[0].Kind, "401: first control should be Link")
}

func TestErrorControlsForStatus_403(t *testing.T) {
	controls := ErrorControlsForStatus(403, ErrorControlOpts{})
	require.Len(t, controls, 2, "403 should return 2 controls")
	require.Equal(t, ControlKindBack, controls[0].Kind)
	require.Equal(t, ControlKindDismiss, controls[1].Kind)
}

func TestErrorControlsForStatus_500(t *testing.T) {
	opts := ErrorControlOpts{RetryURL: "/api", RetryTarget: "#main"}
	controls := ErrorControlsForStatus(500, opts)
	require.Equal(t, ControlKindRetry, controls[0].Kind, "500: first control should be Retry")
}

func TestErrorControlsForStatus_503(t *testing.T) {
	opts := ErrorControlOpts{RetryURL: "/api", RetryTarget: "#main"}
	controls := ErrorControlsForStatus(503, opts)
	require.Equal(t, ControlKindRetry, controls[0].Kind, "503: first control should be Retry")
}

func TestErrorControlsForStatus_Unknown418(t *testing.T) {
	controls := ErrorControlsForStatus(418, ErrorControlOpts{})
	require.Len(t, controls, 1, "unknown status should return 1 control")
	require.Equal(t, ControlKindDismiss, controls[0].Kind, "fallback should be DismissButton")
}

// ---------------------------------------------------------------------------
// ResourceActions
// ---------------------------------------------------------------------------

func TestResourceActions(t *testing.T) {
	controls := ResourceActions(ResourceActionCfg{
		EditURL:    "/edit/1",
		DeleteURL:  "/delete/1",
		ConfirmMsg: "Sure?",
		Target:     "#row-1",
	})
	require.Len(t, controls, 2)

	edit := controls[0]
	require.Equal(t, LabelEdit, edit.Label)
	require.Equal(t, VariantSecondary, edit.Variant, "Edit should be secondary variant")
	require.Equal(t, IconPencilSquare, edit.Icon, "Edit should have pencil-square icon")
	require.Equal(t, ControlKindHTMX, edit.Kind)
	require.Equal(t, HxMethodGet, edit.HxRequest.Method)
	require.Equal(t, "/edit/1", edit.HxRequest.URL)
	require.Equal(t, "#row-1", edit.HxRequest.Target)
	require.Empty(t, edit.Swap, "ResourceActions Edit should not set Swap")

	del := controls[1]
	require.Equal(t, LabelDelete, del.Label)
	require.Equal(t, VariantDanger, del.Variant, "Delete should be danger variant")
	require.Equal(t, "Sure?", del.Confirm, "Delete should have confirm message")
	require.Equal(t, HxMethodDelete, del.HxRequest.Method)
	require.Equal(t, "/delete/1", del.HxRequest.URL)
}

// ---------------------------------------------------------------------------
// FormActions
// ---------------------------------------------------------------------------

func TestFormActions(t *testing.T) {
	controls := FormActions("/cancel")
	require.Len(t, controls, 2)

	save := controls[0]
	require.Equal(t, LabelSave, save.Label)
	require.Equal(t, VariantPrimary, save.Variant, "Save should be primary variant")
	require.Equal(t, IconCheck, save.Icon, "Save should have check icon")
	require.Equal(t, ControlKindHTMX, save.Kind)
	require.Equal(t, IncludeClosestForm, save.HxRequest.Include, "Save should include closest form")
	require.Empty(t, save.Swap, "FormActions Save should not set Swap")

	cancel := controls[1]
	require.Equal(t, LabelCancel, cancel.Label)
	require.Equal(t, VariantGhost, cancel.Variant, "Cancel should be ghost variant")
	require.Equal(t, IconXMark, cancel.Icon, "Cancel should have x-mark icon")
	require.Equal(t, ControlKindLink, cancel.Kind, "Cancel should be a link")
	require.Equal(t, "/cancel", cancel.Href)
}

// ---------------------------------------------------------------------------
// RowActions
// ---------------------------------------------------------------------------

func TestRowActions(t *testing.T) {
	controls := RowActions(RowActionCfg{
		EditURL:    "/edit/1",
		DeleteURL:  "/delete/1",
		RowTarget:  "#row-1",
		ConfirmMsg: "Delete this?",
	})
	require.Len(t, controls, 2)

	edit := controls[0]
	require.Equal(t, LabelEdit, edit.Label)
	require.Equal(t, VariantSecondary, edit.Variant, "Edit should be secondary variant")
	require.Equal(t, IconPencilSquare, edit.Icon, "Edit should have pencil-square icon")
	require.Equal(t, SwapOuterHTML, edit.Swap, "Edit should use outerHTML swap")
	require.Equal(t, HxMethodGet, edit.HxRequest.Method)
	require.Equal(t, "/edit/1", edit.HxRequest.URL)
	require.Equal(t, "#row-1", edit.HxRequest.Target)

	del := controls[1]
	require.Equal(t, LabelDelete, del.Label)
	require.Equal(t, VariantDanger, del.Variant, "Delete should be danger variant")
	require.Equal(t, SwapOuterHTML, del.Swap, "Delete should use outerHTML swap")
	require.Equal(t, "Delete this?", del.Confirm)
	require.Equal(t, HxMethodDelete, del.HxRequest.Method)
	require.Equal(t, "/delete/1", del.HxRequest.URL)
	require.Equal(t, "#row-1", del.HxRequest.Target)
}

func TestRowActions_EditOnly(t *testing.T) {
	controls := RowActions(RowActionCfg{EditURL: "/edit/1", RowTarget: "#row-1"})
	require.Len(t, controls, 1)
	require.Equal(t, "Edit", controls[0].Label)
}

func TestRowActions_DeleteOnly(t *testing.T) {
	controls := RowActions(RowActionCfg{DeleteURL: "/delete/1", RowTarget: "#row-1", ConfirmMsg: "Sure?"})
	require.Len(t, controls, 1)
	require.Equal(t, "Delete", controls[0].Label)
}

func TestRowActions_Empty(t *testing.T) {
	controls := RowActions(RowActionCfg{RowTarget: "#row-1"})
	require.Empty(t, controls)
}

// ---------------------------------------------------------------------------
// TableRowActions — partial configs
// ---------------------------------------------------------------------------

func TestTableRowActions_Full(t *testing.T) {
	controls := TableRowActions(TableRowActionCfg{
		EditURL: "/edit/1", DeleteURL: "/delete/1",
		RowTarget: "#row-1", TableTarget: "#tc", ConfirmMsg: "Sure?",
	})
	require.Len(t, controls, 2)

	edit := controls[0]
	require.Equal(t, LabelEdit, edit.Label)
	require.Equal(t, VariantSecondary, edit.Variant)
	require.Equal(t, IconPencilSquare, edit.Icon)
	require.Equal(t, SwapOuterHTML, edit.Swap)
	require.Equal(t, HxMethodGet, edit.HxRequest.Method)
	require.Equal(t, "#row-1", edit.HxRequest.Target, "Edit targets RowTarget")

	del := controls[1]
	require.Equal(t, LabelDelete, del.Label)
	require.Equal(t, VariantDanger, del.Variant)
	require.Equal(t, SwapOuterHTML, del.Swap)
	require.Equal(t, "Sure?", del.Confirm)
	require.Equal(t, HxMethodDelete, del.HxRequest.Method)
	require.Equal(t, "#tc", del.HxRequest.Target, "Delete targets TableTarget")
}

func TestTableRowActions_EditOnly(t *testing.T) {
	controls := TableRowActions(TableRowActionCfg{EditURL: "/edit/1", RowTarget: "#row-1"})
	require.Len(t, controls, 1)
	require.Equal(t, "Edit", controls[0].Label)
}

func TestTableRowActions_DeleteOnly(t *testing.T) {
	controls := TableRowActions(TableRowActionCfg{DeleteURL: "/del/1", TableTarget: "#tc"})
	require.Len(t, controls, 1)
	require.Equal(t, "Delete", controls[0].Label)
}

func TestTableRowActions_Empty(t *testing.T) {
	controls := TableRowActions(TableRowActionCfg{})
	require.Empty(t, controls)
}

// ---------------------------------------------------------------------------
// ResourceActions — partial configs
// ---------------------------------------------------------------------------

func TestResourceActions_EditOnly(t *testing.T) {
	controls := ResourceActions(ResourceActionCfg{EditURL: "/edit/1", Target: "#row-1"})
	require.Len(t, controls, 1)
	require.Equal(t, "Edit", controls[0].Label)
}

func TestResourceActions_DeleteOnly(t *testing.T) {
	controls := ResourceActions(ResourceActionCfg{DeleteURL: "/del/1", Target: "#row-1", ConfirmMsg: "Sure?"})
	require.Len(t, controls, 1)
	require.Equal(t, "Delete", controls[0].Label)
}

func TestResourceActions_Empty(t *testing.T) {
	controls := ResourceActions(ResourceActionCfg{})
	require.Empty(t, controls)
}

// ---------------------------------------------------------------------------
// RowFormActions — partial configs
// ---------------------------------------------------------------------------

func TestRowFormActions_Full(t *testing.T) {
	controls := RowFormActions(RowFormActionCfg{
		SaveURL: "/save/1", CancelURL: "/cancel/1",
		SaveTarget: "#tc", CancelTarget: "#row-1",
	})
	require.Len(t, controls, 2)

	save := controls[0]
	require.Equal(t, LabelSave, save.Label)
	require.Equal(t, VariantPrimary, save.Variant)
	require.Equal(t, IconCheck, save.Icon)
	require.Equal(t, SwapOuterHTML, save.Swap)
	require.Equal(t, HxMethodPut, save.HxRequest.Method, "RowFormActions Save should use hx-put")
	require.Equal(t, "/save/1", save.HxRequest.URL)
	require.Equal(t, "#tc", save.HxRequest.Target)
	require.Equal(t, IncludeClosestTR, save.HxRequest.Include, "Save should include closest tr")

	cancel := controls[1]
	require.Equal(t, LabelCancel, cancel.Label)
	require.Equal(t, VariantGhost, cancel.Variant)
	require.Equal(t, IconXMark, cancel.Icon)
	require.Equal(t, SwapOuterHTML, cancel.Swap)
	require.Equal(t, HxMethodGet, cancel.HxRequest.Method)
	require.Equal(t, "/cancel/1", cancel.HxRequest.URL)
	require.Equal(t, "#row-1", cancel.HxRequest.Target)
}

func TestRowFormActions_SaveOnly(t *testing.T) {
	controls := RowFormActions(RowFormActionCfg{SaveURL: "/save/1", SaveTarget: "#tc"})
	require.Len(t, controls, 1)
	require.Equal(t, "Save", controls[0].Label)
}

func TestRowFormActions_CancelOnly(t *testing.T) {
	controls := RowFormActions(RowFormActionCfg{CancelURL: "/cancel/1", CancelTarget: "#row-1"})
	require.Len(t, controls, 1)
	require.Equal(t, "Cancel", controls[0].Label)
}

func TestRowFormActions_Empty(t *testing.T) {
	controls := RowFormActions(RowFormActionCfg{})
	require.Empty(t, controls)
}

// ---------------------------------------------------------------------------
// NewRowFormActions — partial configs
// ---------------------------------------------------------------------------

func TestNewRowFormActions_Full(t *testing.T) {
	controls := NewRowFormActions(RowFormActionCfg{
		SaveURL: "/new", CancelURL: "/cancel",
		SaveTarget: "#tc", CancelTarget: "#row-new",
	})
	require.Len(t, controls, 2)

	save := controls[0]
	require.Equal(t, LabelSave, save.Label)
	require.Equal(t, VariantPrimary, save.Variant)
	require.Equal(t, IconCheck, save.Icon)
	require.Equal(t, SwapOuterHTML, save.Swap)
	require.Equal(t, HxMethodPost, save.HxRequest.Method, "NewRowFormActions Save should use hx-post")
	require.Equal(t, "/new", save.HxRequest.URL)
	require.Equal(t, "#tc", save.HxRequest.Target)
	require.Equal(t, IncludeClosestTR, save.HxRequest.Include, "Save should include closest tr")

	cancel := controls[1]
	require.Equal(t, LabelCancel, cancel.Label)
	require.Equal(t, VariantGhost, cancel.Variant)
	require.Equal(t, IconXMark, cancel.Icon)
	require.Equal(t, SwapOuterHTML, cancel.Swap)
	require.Equal(t, HxMethodGet, cancel.HxRequest.Method)
	require.Equal(t, "/cancel", cancel.HxRequest.URL)
	require.Equal(t, "#row-new", cancel.HxRequest.Target)
}

func TestNewRowFormActions_SaveOnly(t *testing.T) {
	controls := NewRowFormActions(RowFormActionCfg{SaveURL: "/new", SaveTarget: "#tc"})
	require.Len(t, controls, 1)
	require.Equal(t, "Save", controls[0].Label)
}

func TestNewRowFormActions_CancelOnly(t *testing.T) {
	controls := NewRowFormActions(RowFormActionCfg{CancelURL: "/cancel", CancelTarget: "#row-new"})
	require.Len(t, controls, 1)
	require.Equal(t, "Cancel", controls[0].Label)
}

func TestNewRowFormActions_Empty(t *testing.T) {
	controls := NewRowFormActions(RowFormActionCfg{})
	require.Empty(t, controls)
}

// ---------------------------------------------------------------------------
// BulkActions — partial configs
// ---------------------------------------------------------------------------

func TestBulkActions_Full(t *testing.T) {
	controls := BulkActions(BulkActionCfg{
		DeleteURL: "/del", ActivateURL: "/act", DeactivateURL: "/deact",
		TableTarget: "#tc", CheckboxSelector: ".row-check",
	})
	require.Len(t, controls, 3)

	del := controls[0]
	require.Equal(t, LabelDeleteSelected, del.Label)
	require.Equal(t, VariantDanger, del.Variant)
	require.Equal(t, SwapOuterHTML, del.Swap)
	require.Equal(t, ConfirmDeleteSelected, del.Confirm)
	require.Equal(t, HxMethodDelete, del.HxRequest.Method)
	require.Equal(t, "/del", del.HxRequest.URL)
	require.Equal(t, "#tc", del.HxRequest.Target)
	require.Equal(t, ".row-check", del.HxRequest.Include, "Delete should include checkbox selector")

	act := controls[1]
	require.Equal(t, LabelActivate, act.Label)
	require.Equal(t, VariantSecondary, act.Variant)
	require.Equal(t, SwapOuterHTML, act.Swap)
	require.Equal(t, HxMethodPut, act.HxRequest.Method)
	require.Equal(t, ".row-check", act.HxRequest.Include)

	deact := controls[2]
	require.Equal(t, LabelDeactivate, deact.Label)
	require.Equal(t, VariantGhost, deact.Variant)
	require.Equal(t, SwapOuterHTML, deact.Swap)
	require.Equal(t, HxMethodPut, deact.HxRequest.Method)
	require.Equal(t, ".row-check", deact.HxRequest.Include)
}

func TestBulkActions_DeleteOnly(t *testing.T) {
	controls := BulkActions(BulkActionCfg{DeleteURL: "/del", TableTarget: "#tc", CheckboxSelector: ".c"})
	require.Len(t, controls, 1)
	require.Equal(t, "Delete Selected", controls[0].Label)
}

func TestBulkActions_ActivateDeactivateOnly(t *testing.T) {
	controls := BulkActions(BulkActionCfg{
		ActivateURL: "/act", DeactivateURL: "/deact",
		TableTarget: "#tc", CheckboxSelector: ".c",
	})
	require.Len(t, controls, 2)
	require.Equal(t, "Activate", controls[0].Label)
	require.Equal(t, "Deactivate", controls[1].Label)
}

func TestBulkActions_SingleActivate(t *testing.T) {
	controls := BulkActions(BulkActionCfg{ActivateURL: "/act", TableTarget: "#tc", CheckboxSelector: ".c"})
	require.Len(t, controls, 1)
	require.Equal(t, "Activate", controls[0].Label)
}

func TestBulkActions_Empty(t *testing.T) {
	controls := BulkActions(BulkActionCfg{TableTarget: "#tc", CheckboxSelector: ".c"})
	require.Empty(t, controls)
}

// ---------------------------------------------------------------------------
// EmptyStateAction
// ---------------------------------------------------------------------------

func TestEmptyStateAction(t *testing.T) {
	ctrl := EmptyStateAction("Add Item", "/items/new", "#list")
	require.Equal(t, VariantPrimary, ctrl.Variant, "should be primary variant")
	require.Equal(t, "Add Item", ctrl.Label)
	require.Equal(t, ControlKindHTMX, ctrl.Kind)
	require.Equal(t, HxMethodGet, ctrl.HxRequest.Method)
	require.Equal(t, "/items/new", ctrl.HxRequest.URL)
	require.Equal(t, "#list", ctrl.HxRequest.Target)
	require.Empty(t, ctrl.Swap, "EmptyStateAction should not set Swap")
	require.Empty(t, ctrl.Icon, "EmptyStateAction should not set Icon")
}

// ---------------------------------------------------------------------------
// CatalogRowAction
// ---------------------------------------------------------------------------

func TestCatalogRowAction(t *testing.T) {
	ctrl := CatalogRowAction("/catalog/42/detail", "#detail-row-42")

	require.Equal(t, ControlKindHTMX, ctrl.Kind)
	require.Equal(t, LabelDetails, ctrl.Label)
	require.Equal(t, VariantGhost, ctrl.Variant)
	require.Equal(t, SwapInnerHTML, ctrl.Swap, "CatalogRowAction should use innerHTML swap")
	require.Equal(t, HxMethodGet, ctrl.HxRequest.Method)
	require.Equal(t, "/catalog/42/detail", ctrl.HxRequest.URL)
	require.Equal(t, "#detail-row-42", ctrl.HxRequest.Target)
	require.Empty(t, ctrl.Icon, "CatalogRowAction should not set Icon")
	require.Empty(t, ctrl.Confirm, "CatalogRowAction should not set Confirm")
}

// ---------------------------------------------------------------------------
// SelectOptions (from filter.go)
// ---------------------------------------------------------------------------

func TestSelectOptions_BasicPairs(t *testing.T) {
	opts := SelectOptions("b", "a", "Alpha", "b", "Beta")
	require.Len(t, opts, 2)
	require.Equal(t, "a", opts[0].Value)
	require.Equal(t, "Alpha", opts[0].Label)
	require.False(t, opts[0].Selected, "a should not be selected")
	require.Equal(t, "b", opts[1].Value)
	require.Equal(t, "Beta", opts[1].Label)
	require.True(t, opts[1].Selected, "b should be selected")
}

func TestSelectOptions_NoCurrentMatch(t *testing.T) {
	opts := SelectOptions("z", "a", "Alpha", "b", "Beta")
	for _, o := range opts {
		require.False(t, o.Selected, "%s should not be selected", o.Value)
	}
}

func TestSelectOptions_OddNumberOfArgs(t *testing.T) {
	opts := SelectOptions("", "a", "Alpha", "b")
	require.Len(t, opts, 1, "trailing unpaired value should be ignored")
	require.Equal(t, "a", opts[0].Value)
}

func TestSelectOptions_EmptyPairs(t *testing.T) {
	opts := SelectOptions("")
	require.Empty(t, opts)
}

// ---------------------------------------------------------------------------
// NewFilterBar
// ---------------------------------------------------------------------------

func TestNewFilterBar(t *testing.T) {
	f1 := SearchField("q", "Search...", "")
	f2 := CheckboxField("active", "Active", "true")
	bar := NewFilterBar("/users", "#tc", f1, f2)
	require.Equal(t, DefaultFilterFormID, bar.ID)
	require.Equal(t, "/users", bar.Action)
	require.Equal(t, "#tc", bar.Target)
	require.Len(t, bar.Fields, 2)
	require.Equal(t, f1, bar.Fields[0])
	require.Equal(t, f2, bar.Fields[1])
}

// ---------------------------------------------------------------------------
// Field factories
// ---------------------------------------------------------------------------

func TestSearchField(t *testing.T) {
	f := SearchField("q", "Search...", "hello")
	require.Equal(t, FilterKindSearch, f.Kind)
	require.Equal(t, "q", f.Name)
	require.Equal(t, "Search...", f.Placeholder)
	require.Equal(t, "hello", f.Value)
}

func TestSelectField(t *testing.T) {
	opts := []FilterOption{{Value: "a", Label: "A"}}
	f := SelectField("status", "Status", "a", opts)
	require.Equal(t, FilterKindSelect, f.Kind)
	require.Equal(t, "status", f.Name)
	require.Equal(t, "Status", f.Label)
	require.Equal(t, "a", f.Value)
	require.Equal(t, opts, f.Options)
}

func TestRangeField(t *testing.T) {
	f := RangeField("price", "Price", "50", "0", "100", "5")
	require.Equal(t, FilterKindRange, f.Kind)
	require.Equal(t, "price", f.Name)
	require.Equal(t, "Price", f.Label)
	require.Equal(t, "50", f.Value)
	require.Equal(t, "0", f.Min)
	require.Equal(t, "100", f.Max)
	require.Equal(t, "5", f.Step)
}

func TestCheckboxField(t *testing.T) {
	f := CheckboxField("active", "Active Only", "true")
	require.Equal(t, FilterKindCheckbox, f.Kind)
	require.Equal(t, "active", f.Name)
	require.Equal(t, "Active Only", f.Label)
	require.Equal(t, "true", f.Value)
}

func TestDateField(t *testing.T) {
	f := DateField("from", "From Date", "2025-01-01")
	require.Equal(t, FilterKindDate, f.Kind)
	require.Equal(t, "from", f.Name)
	require.Equal(t, "From Date", f.Label)
	require.Equal(t, "2025-01-01", f.Value)
}
