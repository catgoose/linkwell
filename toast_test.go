package linkwell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuccessToast(t *testing.T) {
	toast := SuccessToast("saved")

	require.Equal(t, "saved", toast.Message)
	require.Equal(t, ToastSuccess, toast.Variant)
	require.Empty(t, toast.Controls)
	require.Zero(t, toast.AutoDismiss)
	require.Empty(t, toast.OOBTarget)
	require.Empty(t, toast.OOBSwap)
}

func TestInfoToast(t *testing.T) {
	toast := InfoToast("processing")

	require.Equal(t, "processing", toast.Message)
	require.Equal(t, ToastInfo, toast.Variant)
}

func TestWarningToast(t *testing.T) {
	toast := WarningToast("rate limit approaching")

	require.Equal(t, "rate limit approaching", toast.Message)
	require.Equal(t, ToastWarning, toast.Variant)
}

func TestErrorToast(t *testing.T) {
	toast := ErrorToast("upload failed")

	require.Equal(t, "upload failed", toast.Message)
	require.Equal(t, ToastError, toast.Variant)
}

func TestToast_WithControls(t *testing.T) {
	original := SuccessToast("user deleted").
		WithControls(BackButton("Back"))

	updated := original.WithControls(GoHomeButton("Home", "/", "#main"))

	// Original is not mutated.
	require.Len(t, original.Controls, 1)
	// Updated has both controls.
	require.Len(t, updated.Controls, 2)
	require.Equal(t, ControlKindBack, updated.Controls[0].Kind)
	require.Equal(t, ControlKindHome, updated.Controls[1].Kind)
}

func TestToast_WithAutoDismiss(t *testing.T) {
	original := SuccessToast("saved")

	updated := original.WithAutoDismiss(5)

	// Original is not mutated.
	require.Zero(t, original.AutoDismiss)
	// Updated carries the duration.
	require.Equal(t, 5, updated.AutoDismiss)
}

func TestToast_WithOOB(t *testing.T) {
	original := InfoToast("synced")

	updated := original.WithOOB("#toast-container", "afterbegin")

	// Original is not mutated.
	require.Empty(t, original.OOBTarget)
	require.Empty(t, original.OOBSwap)
	// Updated carries OOB fields.
	require.Equal(t, "#toast-container", updated.OOBTarget)
	require.Equal(t, "afterbegin", updated.OOBSwap)
}

func TestToast_FullBuilder(t *testing.T) {
	toast := SuccessToast("User deleted").
		WithControls(HTMXAction("Undo", HxPost("/users/42/restore", "#user-table"))).
		WithAutoDismiss(5).
		WithOOB("#toast-container", "afterbegin")

	require.Equal(t, "User deleted", toast.Message)
	require.Equal(t, ToastSuccess, toast.Variant)
	require.Len(t, toast.Controls, 1)
	require.Equal(t, "Undo", toast.Controls[0].Label)
	require.Equal(t, HxMethodPost, toast.Controls[0].HxRequest.Method)
	require.Equal(t, "/users/42/restore", toast.Controls[0].HxRequest.URL)
	require.Equal(t, 5, toast.AutoDismiss)
	require.Equal(t, "#toast-container", toast.OOBTarget)
	require.Equal(t, "afterbegin", toast.OOBSwap)
}
