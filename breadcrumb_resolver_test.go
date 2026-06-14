package linkwell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBreadcrumbResolver_FirstNonEmptySourceWins(t *testing.T) {
	resolver := NewBreadcrumbResolver(
		func(BreadcrumbResolveContext) []Breadcrumb { return nil },
		func(BreadcrumbResolveContext) []Breadcrumb {
			return []Breadcrumb{{Label: "Explicit", Href: ""}}
		},
		PathBreadcrumbs(),
	)
	crumbs := resolver.Resolve(BreadcrumbResolveContext{Path: "/fallback"})
	require.Equal(t, []Breadcrumb{{Label: "Explicit", Href: ""}}, crumbs)
}

func TestBreadcrumbResolver_ExplicitBeatsPolicyAndPathFallback(t *testing.T) {
	policy := Breadcrumbs().Prefix("/app").Root("App").Crumb("/users/:id", "Policy User")
	resolver := NewBreadcrumbResolver(
		ExplicitBreadcrumbs(),
		PolicyBreadcrumbs(policy),
		PathBreadcrumbs(),
	)
	crumbs := resolver.Resolve(BreadcrumbResolveContext{
		Path:     "/app/users/42",
		Explicit: []Breadcrumb{{Label: "Route Label", Href: ""}},
		RuntimeLabels: []CrumbOption{
			CrumbLabel("/users/:id", "Runtime User"),
		},
	})
	require.Equal(t, []Breadcrumb{{Label: "Route Label", Href: ""}}, crumbs)
}

func TestBreadcrumbResolver_PolicySourceUsesRuntimeLabels(t *testing.T) {
	policy := Breadcrumbs().Prefix("/app/sales-goals").Root("Sales Goals").
		Crumb("/agents", "Agents")
	resolver := NewBreadcrumbResolver(ExplicitBreadcrumbs(), PolicyBreadcrumbs(policy), PathBreadcrumbs())
	crumbs := resolver.Resolve(BreadcrumbResolveContext{
		Path: "/app/sales-goals/agents/346",
		RuntimeLabels: []CrumbOption{
			CrumbLabel("/agents/:id", "Ashley Pope"),
		},
	})
	require.Equal(t, []string{"Sales Goals", "Agents", "Ashley Pope"}, labels(crumbs))
	require.Empty(t, crumbs[2].Href)
}

func TestBreadcrumbResolver_ParentSourceCanPrecedePathFallback(t *testing.T) {
	resolver := NewBreadcrumbResolver(ExplicitBreadcrumbs(), ParentBreadcrumbs(), PathBreadcrumbs())
	crumbs := resolver.Resolve(BreadcrumbResolveContext{
		Path: "/app/reports/42",
		Parent: []Breadcrumb{
			{Label: "Reports", Href: "/app/reports"},
			{Label: "Report 42", Href: ""},
		},
	})
	require.Equal(t, []Breadcrumb{
		{Label: "Reports", Href: "/app/reports"},
		{Label: "Report 42", Href: ""},
	}, crumbs)
}

func TestBreadcrumbResolver_PathFallback(t *testing.T) {
	resolver := NewBreadcrumbResolver(ExplicitBreadcrumbs(), ParentBreadcrumbs(), PathBreadcrumbs())
	crumbs := resolver.Resolve(BreadcrumbResolveContext{
		Path:       "/users/42/edit?tab=summary",
		PathLabels: map[int]string{1: "User 42"},
	})
	require.Equal(t, []string{"Home", "Users", "User 42", "Edit"}, labels(crumbs))
	require.Empty(t, crumbs[len(crumbs)-1].Href)
}

func TestBreadcrumbResolver_ClonesCallerOwnedSlices(t *testing.T) {
	explicit := []Breadcrumb{{Label: "Original", Href: ""}}
	resolver := NewBreadcrumbResolver(ExplicitBreadcrumbs())
	crumbs := resolver.Resolve(BreadcrumbResolveContext{Explicit: explicit})
	crumbs[0].Label = "Changed"
	require.Equal(t, "Original", explicit[0].Label)

	resolver = NewBreadcrumbResolver(func(BreadcrumbResolveContext) []Breadcrumb {
		return explicit
	})
	crumbs = resolver.Resolve(BreadcrumbResolveContext{})
	crumbs[0].Label = "Changed Again"
	require.Equal(t, "Original", explicit[0].Label)
}
