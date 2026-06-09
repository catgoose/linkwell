package linkwell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func salesPolicy() BreadcrumbPolicy {
	return Breadcrumbs().
		Prefix("/app/sales-goals").
		Root("Sales Goals").
		Crumb("/subdivisions", "Subdivisions").
		Crumb("/agents", "Agents")
}

func labels(crumbs []Breadcrumb) []string {
	out := make([]string, len(crumbs))
	for i, c := range crumbs {
		out[i] = c.Label
	}
	return out
}

func TestBreadcrumbPolicy_StaticSection(t *testing.T) {
	crumbs := salesPolicy().Resolve("/app/sales-goals/subdivisions")
	require.Equal(t, []string{"Sales Goals", "Subdivisions"}, labels(crumbs))
	require.Equal(t, "/app/sales-goals", crumbs[0].Href)
	require.Empty(t, crumbs[1].Href, "terminal crumb is unlinked")
}

func TestBreadcrumbPolicy_RuntimeEntityLabel(t *testing.T) {
	crumbs := salesPolicy().Resolve("/app/sales-goals/subdivisions/81",
		CrumbLabel("/subdivisions/:id", "Atwater Villas"),
	)
	require.Equal(t, []string{"Sales Goals", "Subdivisions", "Atwater Villas"}, labels(crumbs))
	require.Equal(t, "/app/sales-goals/subdivisions", crumbs[1].Href)
	require.Empty(t, crumbs[2].Href)
}

func TestBreadcrumbPolicy_RuntimeAgentLabel(t *testing.T) {
	crumbs := salesPolicy().Resolve("/app/sales-goals/agents/346",
		CrumbLabel("/agents/:id", "Ashley Pope"),
	)
	require.Equal(t, []string{"Sales Goals", "Agents", "Ashley Pope"}, labels(crumbs))
	require.NotEqual(t, "346", crumbs[2].Label, "no raw numeric terminal label")
}

func TestBreadcrumbPolicy_DeepDynamicRoute(t *testing.T) {
	crumbs := salesPolicy().Resolve("/app/sales-goals/subdivisions/81/edit",
		CrumbLabel("/subdivisions/:id", "Atwater Villas"),
	)
	require.Equal(t,
		[]string{"Sales Goals", "Subdivisions", "Atwater Villas", "Edit"},
		labels(crumbs))
	require.Equal(t, "/app/sales-goals/subdivisions/81", crumbs[2].Href)
	require.Empty(t, crumbs[3].Href)
}

func TestBreadcrumbPolicy_PrefixItselfIsTerminalRoot(t *testing.T) {
	crumbs := salesPolicy().Resolve("/app/sales-goals")
	require.Equal(t, []string{"Sales Goals"}, labels(crumbs))
	require.Empty(t, crumbs[0].Href, "root is the current page")
}

func TestBreadcrumbPolicy_RuntimeBeatsPolicyLabel(t *testing.T) {
	p := salesPolicy().Crumb("/subdivisions/:id", "Subdivision")
	crumbs := p.Resolve("/app/sales-goals/subdivisions/81",
		CrumbLabel("/subdivisions/:id", "Atwater Villas"),
	)
	require.Equal(t, "Atwater Villas", crumbs[2].Label)
}

func TestBreadcrumbPolicy_ExactBeatsParamPattern(t *testing.T) {
	p := Breadcrumbs().Prefix("/app").Root("App").
		Crumb("/users/:id", "Some User").
		Crumb("/users/new", "New User")
	crumbs := p.Resolve("/app/users/new")
	require.Equal(t, "New User", crumbs[len(crumbs)-1].Label)
}

func TestBreadcrumbPolicy_FallbackTitleFromPath(t *testing.T) {
	crumbs := salesPolicy().Resolve("/app/sales-goals/error-traces")
	require.Equal(t, "Error Traces", crumbs[1].Label)
}

func TestBreadcrumbPolicy_NormalizesTrailingSlashAndQuery(t *testing.T) {
	crumbs := salesPolicy().Resolve("/app/sales-goals/subdivisions/?page=2#top")
	require.Equal(t, []string{"Sales Goals", "Subdivisions"}, labels(crumbs))
	require.Empty(t, crumbs[1].Href)
}

func TestBreadcrumbPolicy_ParamMatchesSingleSegmentOnly(t *testing.T) {
	// /subdivisions/:id must not match /subdivisions/81/edit at the :id level.
	p := Breadcrumbs().Prefix("/app").Crumb("/a/:id", "One")
	crumbs := p.Resolve("/app/a/1/2")
	require.NotEqual(t, "One", crumbs[len(crumbs)-1].Label)
}

func TestBreadcrumbPolicy_RootHrefOverride(t *testing.T) {
	p := salesPolicy().RootHref("/app/sales-goals/home")
	crumbs := p.Resolve("/app/sales-goals/agents")
	require.Equal(t, "/app/sales-goals/home", crumbs[0].Href)
}

func TestBreadcrumbPolicy_GlobalRootWithoutPrefix(t *testing.T) {
	p := Breadcrumbs().Root("Home").Crumb("/admin/users", "Users")
	crumbs := p.Resolve("/admin/users")
	require.Equal(t, []string{"Home", "Admin", "Users"}, labels(crumbs))
	require.Equal(t, "/", crumbs[0].Href)
	require.Equal(t, "/admin", crumbs[1].Href)
	require.Empty(t, crumbs[2].Href)
}

func TestBreadcrumbPolicy_ReuseDoesNotMutate(t *testing.T) {
	base := Breadcrumbs().Prefix("/app").Root("App")
	withUsers := base.Crumb("/users", "People")
	// base must remain unaffected by the derived policy's Crumb.
	bc := base.Resolve("/app/users")
	wc := withUsers.Resolve("/app/users")
	require.Equal(t, "People", wc[1].Label)
	require.Equal(t, "Users", bc[1].Label, "fallback title, not the derived rule")
	require.Len(t, base.rules, 0)
	require.Len(t, withUsers.rules, 1)
}
