package linkwell

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorContext_WithControls(t *testing.T) {
	original := ErrorContext{
		StatusCode: 404,
		Message:    "not found",
		Controls:   []Control{BackButton("Back")},
	}

	updated := original.WithControls(GoHomeButton("Home", "/", "#main"))

	// Original is not mutated.
	require.Len(t, original.Controls, 1)
	// Updated has both controls.
	require.Len(t, updated.Controls, 2)
	require.Equal(t, ControlKindBack, updated.Controls[0].Kind)
	require.Equal(t, ControlKindHome, updated.Controls[1].Kind)
}

func TestErrorContext_WithOOB(t *testing.T) {
	original := ErrorContext{
		StatusCode: 500,
		Message:    "internal error",
	}

	updated := original.WithOOB("#error-panel", "innerHTML")

	// Original is not mutated.
	require.Empty(t, original.OOBTarget)
	require.Empty(t, original.OOBSwap)
	// Updated carries OOB fields.
	require.Equal(t, "#error-panel", updated.OOBTarget)
	require.Equal(t, "innerHTML", updated.OOBSwap)
}

func TestHTTPError_Error_WithErr(t *testing.T) {
	he := &HTTPError{
		EC: ErrorContext{
			StatusCode: 502,
			Message:    "bad gateway",
			Err:        errors.New("upstream timeout"),
		},
	}

	require.Equal(t, "HTTP 502: bad gateway: upstream timeout", he.Error())
}

func TestHTTPError_Error_WithoutErr(t *testing.T) {
	he := &HTTPError{
		EC: ErrorContext{
			StatusCode: 404,
			Message:    "page not found",
		},
	}

	require.Equal(t, "HTTP 404: page not found", he.Error())
}

func TestNewHTTPError(t *testing.T) {
	ec := ErrorContext{
		StatusCode: 422,
		Message:    "validation failed",
		Controls:   []Control{BackButton("Back")},
	}

	he := NewHTTPError(ec)

	require.NotNil(t, he)
	require.Equal(t, ec.StatusCode, he.EC.StatusCode)
	require.Equal(t, ec.Message, he.EC.Message)
	require.Equal(t, ec.Controls, he.EC.Controls)
}
