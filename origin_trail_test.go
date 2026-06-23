package linkwell

import (
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOriginTrail_RoundTripPreservesOrderLabelsHrefs(t *testing.T) {
	trail := []OriginCrumb{
		{Label: "Sales Goals", Href: "/app/sales-goals"},
		{Label: "Agents", Href: "/app/sales-goals/agents?quarter=2026Q3"},
		{Label: "Ashley Pope", Href: "/app/sales-goals/agents/346"},
	}
	value := OriginTrailParam(trail)
	require.NotEmpty(t, value)

	decoded, ok := OriginTrailFromValue(value)
	require.True(t, ok)
	require.Equal(t, trail, decoded)
}

func TestOriginTrail_TrimsLabels(t *testing.T) {
	value := OriginTrailParam([]OriginCrumb{{Label: "  Agents  ", Href: "/agents"}})
	decoded, ok := OriginTrailFromValue(value)
	require.True(t, ok)
	require.Equal(t, []OriginCrumb{{Label: "Agents", Href: "/agents"}}, decoded)
}

func TestOriginTrail_RejectsInvalidCrumbs(t *testing.T) {
	cases := map[string]OriginCrumb{
		"absolute URL":      {Label: "Ext", Href: "https://evil.example/x"},
		"protocol-relative": {Label: "Ext", Href: "//evil.example/x"},
		"backslash bypass":  {Label: "Ext", Href: "/\\evil.example"},
		"raw backslash":     {Label: "Ext", Href: "/app\\admin"},
		"raw control char":  {Label: "Ext", Href: "/app\x01admin"},
		"escaped backslash": {Label: "Ext", Href: "/app%5Cadmin"},
		"escaped control":   {Label: "Ext", Href: "/app%01admin"},
		"empty label":       {Label: "   ", Href: "/agents"},
		"empty href":        {Label: "Agents", Href: ""},
		"relative href":     {Label: "Agents", Href: "agents"},
		"oversized label":   {Label: strings.Repeat("a", maxOriginLabelLen+1), Href: "/agents"},
		"oversized href":    {Label: "Agents", Href: "/" + strings.Repeat("a", maxOriginHrefLen)},
	}
	for name, crumb := range cases {
		t.Run(name, func(t *testing.T) {
			require.Empty(t, OriginTrailParam([]OriginCrumb{crumb}), "encode should reject")

			value := base64.RawURLEncoding.EncodeToString([]byte(
				`[{"l":"` + crumb.Label + `","h":"` + crumb.Href + `"}]`))
			_, ok := OriginTrailFromValue(value)
			require.False(t, ok, "decode should reject")
		})
	}
}

func TestOriginTrail_RejectsTooManyCrumbs(t *testing.T) {
	trail := make([]OriginCrumb, maxOriginCrumbCount+1)
	for i := range trail {
		trail[i] = OriginCrumb{Label: "Crumb", Href: "/crumb"}
	}
	require.Empty(t, OriginTrailParam(trail))
}

func TestOriginTrailFromValue_RejectsMalformedAndOversized(t *testing.T) {
	_, ok := OriginTrailFromValue("not*base64*")
	require.False(t, ok)

	_, ok = OriginTrailFromValue(base64.RawURLEncoding.EncodeToString([]byte("{not json")))
	require.False(t, ok)

	_, ok = OriginTrailFromValue(strings.Repeat("A", maxOriginTrailEncodedLen+1))
	require.False(t, ok)

	_, ok = OriginTrailFromValue("")
	require.False(t, ok)
}

func TestOriginTrailFromRequest_ReadsDefaultParam(t *testing.T) {
	value := OriginTrailParam([]OriginCrumb{{Label: "Agents", Href: "/agents"}})
	r := httptest.NewRequest("GET", "/agents/346?"+DefaultOriginTrailParam+"="+value, nil)
	decoded, ok := OriginTrailFromRequest(r)
	require.True(t, ok)
	require.Equal(t, []OriginCrumb{{Label: "Agents", Href: "/agents"}}, decoded)

	_, ok = OriginTrailFromRequest(nil)
	require.False(t, ok)

	_, ok = OriginTrailFromRequest(httptest.NewRequest("GET", "/agents/346", nil))
	require.False(t, ok)
}

func TestOriginNav_PreservesQueryAndFragment(t *testing.T) {
	trail := []OriginCrumb{{Label: "Agents", Href: "/agents"}}
	value := OriginTrailParam(trail)

	require.Equal(t, "/agents/346?"+DefaultOriginTrailParam+"="+value,
		OriginNav("/agents/346", trail))
	require.Equal(t, "/agents/346?tab=stats&"+DefaultOriginTrailParam+"="+value,
		OriginNav("/agents/346?tab=stats", trail))
	require.Equal(t, "/agents/346?"+DefaultOriginTrailParam+"="+value+"#summary",
		OriginNav("/agents/346#summary", trail))

	require.Equal(t, "/agents/346", OriginNav("/agents/346", nil))
}

func TestOriginCrumbsToBreadcrumbs(t *testing.T) {
	require.Nil(t, OriginCrumbsToBreadcrumbs(nil))
	require.Equal(t,
		[]Breadcrumb{{Label: "Agents", Href: "/agents"}},
		OriginCrumbsToBreadcrumbs([]OriginCrumb{{Label: "Agents", Href: "/agents"}}))
}

func TestFromNav_StaticBehaviorUnchanged(t *testing.T) {
	require.Equal(t, "/users/42", FromNav("/users/42", ""))
	require.Equal(t, "/users/42?from=3", FromNav("/users/42", "3"))
	require.Equal(t, "/users/42?tab=x&from=3", FromNav("/users/42?tab=x", "3"))

	ResetForTesting()
	RegisterFrom(FromDashboard, Breadcrumb{Label: "Dashboard", Href: "/dashboard"})
	crumbs := ResolveFromMask(FromDashboard)
	require.Equal(t, []Breadcrumb{
		{Label: BreadcrumbLabelHome, Href: "/"},
		{Label: "Dashboard", Href: "/dashboard"},
	}, crumbs)
}
