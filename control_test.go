package linkwell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRetryButton(t *testing.T) {
	ctrl := RetryButton("Try again", HxMethodPost, "/items", "#content")

	require.Equal(t, ControlKindRetry, ctrl.Kind)
	require.Equal(t, VariantPrimary, ctrl.Variant)
	require.Equal(t, "Try again", ctrl.Label)
	require.Equal(t, HxMethodPost, ctrl.HxRequest.Method)
	require.Equal(t, "/items", ctrl.HxRequest.URL)
	require.Equal(t, "#content", ctrl.HxRequest.Target)
}

func TestConfirmAction(t *testing.T) {
	ctrl := ConfirmAction("Delete", HxMethodDelete, "/items/1", "#row-1", "Are you sure?")

	require.Equal(t, ControlKindHTMX, ctrl.Kind)
	require.Equal(t, VariantDanger, ctrl.Variant)
	require.Equal(t, "Are you sure?", ctrl.Confirm)
	require.Equal(t, HxMethodDelete, ctrl.HxRequest.Method)
	require.Equal(t, "/items/1", ctrl.HxRequest.URL)
	require.Equal(t, "#row-1", ctrl.HxRequest.Target)
}

func TestBackButton(t *testing.T) {
	ctrl := BackButton("Go back")

	require.Equal(t, ControlKindBack, ctrl.Kind)
	require.Equal(t, "Go back", ctrl.Label)
	require.Empty(t, ctrl.HxRequest.URL)
	require.Empty(t, ctrl.Href)
}

func TestGoHomeButton(t *testing.T) {
	ctrl := GoHomeButton("Home", "/", "#main")

	require.Equal(t, ControlKindHome, ctrl.Kind)
	require.Equal(t, "/", ctrl.Href)
	require.Equal(t, "/", ctrl.PushURL)
	require.Equal(t, HxMethodGet, ctrl.HxRequest.Method)
	require.Equal(t, "/", ctrl.HxRequest.URL)
	require.Equal(t, "#main", ctrl.HxRequest.Target)
}

func TestRedirectLink(t *testing.T) {
	ctrl := RedirectLink("Dashboard", "/dashboard")

	require.Equal(t, ControlKindLink, ctrl.Kind)
	require.Equal(t, "/dashboard", ctrl.Href)
	require.Empty(t, ctrl.HxRequest.URL)
}

func TestHTMXAction(t *testing.T) {
	req := HxGet("/refresh", "#panel")
	ctrl := HTMXAction("Refresh", req)

	require.Equal(t, ControlKindHTMX, ctrl.Kind)
	require.Equal(t, "Refresh", ctrl.Label)
	require.Equal(t, HxMethodGet, ctrl.HxRequest.Method)
	require.Equal(t, "/refresh", ctrl.HxRequest.URL)
	require.Equal(t, "#panel", ctrl.HxRequest.Target)
}

func TestDismissButton(t *testing.T) {
	ctrl := DismissButton("Close")

	require.Equal(t, ControlKindDismiss, ctrl.Kind)
	require.Equal(t, "Close", ctrl.Label)
}

// ---------------------------------------------------------------------------
// Hx* request config factories
// ---------------------------------------------------------------------------

func TestHxGet(t *testing.T) {
	req := HxGet("/items", "#list")
	require.Equal(t, HxMethodGet, req.Method)
	require.Equal(t, "/items", req.URL)
	require.Equal(t, "#list", req.Target)
	require.Empty(t, req.Include)
}

func TestHxPost(t *testing.T) {
	req := HxPost("/items", "#list")
	require.Equal(t, HxMethodPost, req.Method)
	require.Equal(t, "/items", req.URL)
	require.Equal(t, "#list", req.Target)
	require.Empty(t, req.Include)
}

func TestHxPut(t *testing.T) {
	req := HxPut("/items/1", "#row")
	require.Equal(t, HxMethodPut, req.Method)
	require.Equal(t, "/items/1", req.URL)
	require.Equal(t, "#row", req.Target)
	require.Empty(t, req.Include)
}

func TestHxPatch(t *testing.T) {
	req := HxPatch("/items/1", "#row")
	require.Equal(t, HxMethodPatch, req.Method)
	require.Equal(t, "/items/1", req.URL)
	require.Equal(t, "#row", req.Target)
	require.Empty(t, req.Include)
}

func TestHxDelete(t *testing.T) {
	req := HxDelete("/items/1", "#row")
	require.Equal(t, HxMethodDelete, req.Method)
	require.Equal(t, "/items/1", req.URL)
	require.Equal(t, "#row", req.Target)
	require.Empty(t, req.Include)
}

func TestHxRequestConfig_Attrs(t *testing.T) {
	req := HxRequestConfig{Method: HxMethodPut, URL: "/save", Target: "#tc", Include: "closest tr"}
	attrs := req.Attrs()
	require.Equal(t, "/save", attrs["put"])
	require.Equal(t, "#tc", attrs["target"])
	require.Equal(t, "closest tr", attrs["include"])
}

// ---------------------------------------------------------------------------
// With* fluent methods
// ---------------------------------------------------------------------------

// fullControl returns a Control with every field populated for preservation tests.
func fullControl() Control {
	return Control{
		Kind:      ControlKindHTMX,
		Label:     "Full",
		Href:      "/full",
		Variant:   VariantSecondary,
		Confirm:   "confirm?",
		Icon:      IconHome,
		PushURL:   "/push",
		Swap:      SwapInnerHTML,
		Disabled:  false,
		HxRequest: HxRequestConfig{Method: HxMethodPut, URL: "/save", Target: "#tc", Include: IncludeClosestTR},
	}
}

// assertFieldsPreserved verifies every field except the one under test is unchanged.
func assertFieldsPreserved(t *testing.T, got, orig Control, skip string) {
	t.Helper()
	if skip != "Kind" {
		require.Equal(t, orig.Kind, got.Kind, "Kind should be preserved")
	}
	if skip != "Label" {
		require.Equal(t, orig.Label, got.Label, "Label should be preserved")
	}
	if skip != "Href" {
		require.Equal(t, orig.Href, got.Href, "Href should be preserved")
	}
	if skip != "Variant" {
		require.Equal(t, orig.Variant, got.Variant, "Variant should be preserved")
	}
	if skip != "Confirm" {
		require.Equal(t, orig.Confirm, got.Confirm, "Confirm should be preserved")
	}
	if skip != "Icon" {
		require.Equal(t, orig.Icon, got.Icon, "Icon should be preserved")
	}
	if skip != "PushURL" {
		require.Equal(t, orig.PushURL, got.PushURL, "PushURL should be preserved")
	}
	if skip != "Swap" {
		require.Equal(t, orig.Swap, got.Swap, "Swap should be preserved")
	}
	if skip != "Disabled" {
		require.Equal(t, orig.Disabled, got.Disabled, "Disabled should be preserved")
	}
	if skip != "HxRequest" {
		require.Equal(t, orig.HxRequest, got.HxRequest, "HxRequest should be preserved")
	}
}

func TestControl_WithSwap(t *testing.T) {
	orig := fullControl()
	got := orig.WithSwap(SwapOuterHTML)
	require.Equal(t, SwapOuterHTML, got.Swap)
	require.Equal(t, SwapInnerHTML, orig.Swap, "original must not be mutated")
	assertFieldsPreserved(t, got, orig, "Swap")
}

func TestControl_WithVariant(t *testing.T) {
	orig := fullControl()
	got := orig.WithVariant(VariantDanger)
	require.Equal(t, VariantDanger, got.Variant)
	require.Equal(t, VariantSecondary, orig.Variant, "original must not be mutated")
	assertFieldsPreserved(t, got, orig, "Variant")
}

func TestControl_WithConfirm(t *testing.T) {
	orig := fullControl()
	got := orig.WithConfirm("Really?")
	require.Equal(t, "Really?", got.Confirm)
	require.Equal(t, "confirm?", orig.Confirm, "original must not be mutated")
	assertFieldsPreserved(t, got, orig, "Confirm")
}

func TestControl_WithIcon(t *testing.T) {
	orig := fullControl()
	got := orig.WithIcon(IconPencilSquare)
	require.Equal(t, IconPencilSquare, got.Icon)
	require.Equal(t, IconHome, orig.Icon, "original must not be mutated")
	assertFieldsPreserved(t, got, orig, "Icon")
}

func TestControl_WithDisabled(t *testing.T) {
	orig := fullControl()
	got := orig.WithDisabled(true)
	require.True(t, got.Disabled)
	require.False(t, orig.Disabled, "original must not be mutated")
	assertFieldsPreserved(t, got, orig, "Disabled")
}

func TestControl_WithChaining(t *testing.T) {
	ctrl := HTMXAction("Test", HxGet("/test", "#t")).
		WithSwap(SwapOuterHTML).
		WithVariant(VariantDanger).
		WithConfirm("Sure?").
		WithIcon(IconXMark).
		WithDisabled(true)

	require.Equal(t, SwapOuterHTML, ctrl.Swap)
	require.Equal(t, VariantDanger, ctrl.Variant)
	require.Equal(t, "Sure?", ctrl.Confirm)
	require.Equal(t, IconXMark, ctrl.Icon)
	require.True(t, ctrl.Disabled)
	require.Equal(t, "Test", ctrl.Label, "Label should survive chaining")
	require.Equal(t, ControlKindHTMX, ctrl.Kind, "Kind should survive chaining")
	require.Equal(t, "/test", ctrl.HxRequest.URL, "HxRequest should survive chaining")
}

func TestHxRequestConfig_WithInclude(t *testing.T) {
	orig := HxRequestConfig{Method: HxMethodPut, URL: "/save", Target: "#tc"}
	got := orig.WithInclude(IncludeClosestTR)
	require.Equal(t, IncludeClosestTR, got.Include)
	require.Empty(t, orig.Include, "original must not be mutated")
	require.Equal(t, orig.Method, got.Method, "Method should be preserved")
	require.Equal(t, orig.URL, got.URL, "URL should be preserved")
	require.Equal(t, orig.Target, got.Target, "Target should be preserved")
}

func TestHxRequestConfig_Attrs_OmitsEmpty(t *testing.T) {
	req := HxGet("/items", "")
	attrs := req.Attrs()
	require.Equal(t, "/items", attrs["get"])
	_, hasTarget := attrs["target"]
	require.False(t, hasTarget, "empty target should be omitted")
	_, hasInclude := attrs["include"]
	require.False(t, hasInclude, "empty include should be omitted")
}
