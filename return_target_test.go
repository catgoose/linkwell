package linkwell

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func vendorsCfg() ReturnTargetConfig {
	return ReturnTargetConfig{
		Param:    "back_to",
		Fallback: CanonicalTarget{Label: "Back to vendors", Href: "/vendors"},
	}
}

func TestReturnTarget_AcceptsExactReturn(t *testing.T) {
	got := ReturnTargetFromValue("/vendors/42", vendorsCfg())
	require.True(t, got.Exact)
	require.Equal(t, "/vendors/42", got.Href)
	require.Equal(t, "Back", got.Label)
}

func TestReturnTarget_PreservesQueryAndFragment(t *testing.T) {
	got := ReturnTargetFromValue("/vendors?q=foo&page=2#row-3", vendorsCfg())
	require.True(t, got.Exact)
	require.Equal(t, "/vendors?q=foo&page=2#row-3", got.Href)
}

func TestReturnTarget_MissingFallsBack(t *testing.T) {
	got := ReturnTargetFromValue("", vendorsCfg())
	require.False(t, got.Exact)
	require.Equal(t, "/vendors", got.Href)
	require.Equal(t, "Back to vendors", got.Label)
}

func TestReturnTarget_RejectsAbsoluteURL(t *testing.T) {
	got := ReturnTargetFromValue("https://evil.example.com/x", vendorsCfg())
	require.False(t, got.Exact)
	require.Equal(t, "/vendors", got.Href)
}

func TestReturnTarget_RejectsProtocolRelative(t *testing.T) {
	require.False(t, ReturnTargetFromValue("//evil.example.com/x", vendorsCfg()).Exact)
}

func TestReturnTarget_RejectsBackslashBypass(t *testing.T) {
	require.False(t, ReturnTargetFromValue(`/\evil.example.com`, vendorsCfg()).Exact)
	require.False(t, ReturnTargetFromValue(`/vendors\evil`, vendorsCfg()).Exact)
}

func TestReturnTarget_RejectsEscapedBackslashBypass(t *testing.T) {
	require.False(t, ReturnTargetFromValue(`/%5Cevil.example.com`, vendorsCfg()).Exact)
	require.False(t, ReturnTargetFromValue(`/vendors%5cevil`, vendorsCfg()).Exact)
}

func TestReturnTarget_RejectsControlChars(t *testing.T) {
	require.False(t, ReturnTargetFromValue("/vendors\n/42", vendorsCfg()).Exact)
	require.False(t, ReturnTargetFromValue("/vendors\x00", vendorsCfg()).Exact)
	require.False(t, ReturnTargetFromValue("/vendors\x7f", vendorsCfg()).Exact)
}

func TestReturnTarget_RejectsEscapedControlChars(t *testing.T) {
	require.False(t, ReturnTargetFromValue("/vendors%0a/42", vendorsCfg()).Exact)
	require.False(t, ReturnTargetFromValue("/vendors%00", vendorsCfg()).Exact)
	require.False(t, ReturnTargetFromValue("/vendors%7F", vendorsCfg()).Exact)
}

func TestReturnTarget_RejectsRelativePath(t *testing.T) {
	require.False(t, ReturnTargetFromValue("vendors/42", vendorsCfg()).Exact)
}

func TestReturnTarget_DefaultParamName(t *testing.T) {
	req := httptest.NewRequest("GET", "/page?back_to=/vendors/7", nil)
	got := ReturnTargetFromRequest(req, ReturnTargetConfig{
		Fallback: CanonicalTarget{Label: "Back to vendors", Href: "/vendors"},
	})
	require.True(t, got.Exact)
	require.Equal(t, "/vendors/7", got.Href)
}

func TestReturnTarget_CustomParamName(t *testing.T) {
	cfg := ReturnTargetConfig{Param: "return", Fallback: CanonicalTarget{Href: "/vendors"}}

	hit := httptest.NewRequest("GET", "/page?return=/vendors/7", nil)
	require.True(t, ReturnTargetFromRequest(hit, cfg).Exact)

	// The default name is not consulted once Param is set.
	miss := httptest.NewRequest("GET", "/page?back_to=/vendors/7", nil)
	require.False(t, ReturnTargetFromRequest(miss, cfg).Exact)
}

func TestReturnTarget_PreservesQueryFromRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/page?back_to=%2Fvendors%3Fq%3Dfoo%26page%3D2", nil)
	got := ReturnTargetFromRequest(req, vendorsCfg())
	require.True(t, got.Exact)
	require.Equal(t, "/vendors?q=foo&page=2", got.Href)
}

func TestReturnTarget_ExactLabelOverride(t *testing.T) {
	cfg := vendorsCfg()
	cfg.ExactLabel = "Back to results"
	got := ReturnTargetFromValue("/vendors?page=2", cfg)
	require.Equal(t, "Back to results", got.Label)
}

func TestReturnTarget_NilRequestFallsBack(t *testing.T) {
	got := ReturnTargetFromRequest(nil, vendorsCfg())
	require.False(t, got.Exact)
	require.Equal(t, "/vendors", got.Href)
	require.Equal(t, "Back to vendors", got.Label)
}
