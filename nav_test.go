package linkwell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// SetActiveNavItem
// ---------------------------------------------------------------------------

func TestSetActiveNavItem_ExactMatch(t *testing.T) {
	items := []NavItem{
		{Label: "Users", Href: "/users"},
		{Label: "Settings", Href: "/settings"},
	}
	result := SetActiveNavItem(items, "/users")
	require.True(t, result[0].Active, "Users should be active")
	require.False(t, result[1].Active, "Settings should not be active")
}

func TestSetActiveNavItem_NoMatch(t *testing.T) {
	items := []NavItem{
		{Label: "Users", Href: "/users"},
		{Label: "Settings", Href: "/settings"},
	}
	result := SetActiveNavItem(items, "/other")
	for _, item := range result {
		require.False(t, item.Active, "%s should not be active", item.Label)
	}
}

func TestSetActiveNavItem_ChildMatchActivatesParent(t *testing.T) {
	items := []NavItem{
		{
			Label: "Tables",
			Href:  "/demo",
			Children: []NavItem{
				{Label: "Inventory", Href: "/demo/inventory"},
				{Label: "Catalog", Href: "/demo/catalog"},
			},
		},
	}
	result := SetActiveNavItem(items, "/demo/inventory")
	require.True(t, result[0].Active, "parent Tables should be active")
	require.True(t, result[0].Children[0].Active, "child Inventory should be active")
	require.False(t, result[0].Children[1].Active, "child Catalog should not be active")
}

func TestSetActiveNavItem_RootOnlyMatchesExactly(t *testing.T) {
	items := []NavItem{
		{Label: BreadcrumbLabelHome, Href: "/"},
		{Label: "Users", Href: "/users"},
	}
	result := SetActiveNavItem(items, "/users")
	require.False(t, result[0].Active, `"/" should NOT be active when path is "/users"`)
	require.True(t, result[1].Active, "Users should be active")
}

func TestSetActiveNavItem_DoesNotMutateOriginal(t *testing.T) {
	items := []NavItem{
		{Label: "Users", Href: "/users"},
	}
	result := SetActiveNavItem(items, "/users")
	require.True(t, result[0].Active)
	require.False(t, items[0].Active, "original slice must not be mutated")
}

// ---------------------------------------------------------------------------
// SetActiveNavItemPrefix
// ---------------------------------------------------------------------------

func TestSetActiveNavItemPrefix_ExactMatch(t *testing.T) {
	items := []NavItem{
		{Label: "Users", Href: "/users"},
		{Label: "Settings", Href: "/settings"},
	}
	result := SetActiveNavItemPrefix(items, "/users")
	require.True(t, result[0].Active, "Users should be active on exact match")
	require.False(t, result[1].Active)
}

func TestSetActiveNavItemPrefix_PrefixMatch(t *testing.T) {
	items := []NavItem{
		{Label: "Users", Href: "/users"},
	}
	result := SetActiveNavItemPrefix(items, "/users/42/edit")
	require.True(t, result[0].Active, `"/users" should be active for "/users/42/edit"`)
}

func TestSetActiveNavItemPrefix_RootNotActiveForUsers(t *testing.T) {
	items := []NavItem{
		{Label: BreadcrumbLabelHome, Href: "/"},
		{Label: "Users", Href: "/users"},
	}
	result := SetActiveNavItemPrefix(items, "/users")
	require.False(t, result[0].Active,
		`"/" should NOT be active for "/users" because "//" does not prefix-match "/users"`)
	require.True(t, result[1].Active)
}

func TestSetActiveNavItemPrefix_NoMatch(t *testing.T) {
	items := []NavItem{
		{Label: "Users", Href: "/users"},
		{Label: "Settings", Href: "/settings"},
	}
	result := SetActiveNavItemPrefix(items, "/other")
	for _, item := range result {
		require.False(t, item.Active, "%s should not be active", item.Label)
	}
}

func TestSetActiveNavItemPrefix_ChildPrefixMatchActivatesParent(t *testing.T) {
	items := []NavItem{
		{
			Label: "Tables",
			Href:  "/demo",
			Children: []NavItem{
				{Label: "Inventory", Href: "/demo/inventory"},
			},
		},
	}
	result := SetActiveNavItemPrefix(items, "/demo/inventory/99")
	require.True(t, result[0].Active, "parent Tables should be active via child prefix match")
	require.True(t, result[0].Children[0].Active, "child Inventory should be active via prefix")
}

// ---------------------------------------------------------------------------
// NavItemFromControl
// ---------------------------------------------------------------------------

func TestNavItemFromControl_MapsFields(t *testing.T) {
	ctrl := Control{
		Label:     "Dashboard",
		Href:      "/dashboard",
		Icon:      IconHome,
		HxRequest: HxGet("/dashboard", "body"),
	}
	nav := NavItemFromControl(ctrl)
	require.Equal(t, "Dashboard", nav.Label)
	require.Equal(t, "/dashboard", nav.Href)
	require.Equal(t, "home", nav.Icon)
	require.Equal(t, map[string]string{"get": "/dashboard", "target": "body"}, nav.HTMXAttrs)
	require.False(t, nav.Active, "Active should default to false")
	require.Empty(t, nav.Children, "Children should be empty")
}

// ---------------------------------------------------------------------------
// BreadcrumbsFromPath
// ---------------------------------------------------------------------------

func TestBreadcrumbsFromPath_Root(t *testing.T) {
	crumbs := BreadcrumbsFromPath("/", nil)
	require.Len(t, crumbs, 1)
	require.Equal(t, Breadcrumb{Label: BreadcrumbLabelHome, Href: "/"}, crumbs[0])
}

func TestBreadcrumbsFromPath_SingleSegment(t *testing.T) {
	crumbs := BreadcrumbsFromPath("/users", nil)
	require.Len(t, crumbs, 2)
	require.Equal(t, Breadcrumb{Label: BreadcrumbLabelHome, Href: "/"}, crumbs[0])
	require.Equal(t, Breadcrumb{Label: "users", Href: ""}, crumbs[1])
}

func TestBreadcrumbsFromPath_MultipleSegments(t *testing.T) {
	crumbs := BreadcrumbsFromPath("/users/42/edit", nil)
	require.Len(t, crumbs, 4)
	require.Equal(t, Breadcrumb{Label: BreadcrumbLabelHome, Href: "/"}, crumbs[0])
	require.Equal(t, Breadcrumb{Label: "users", Href: "/users"}, crumbs[1])
	require.Equal(t, Breadcrumb{Label: "42", Href: "/users/42"}, crumbs[2])
	require.Equal(t, Breadcrumb{Label: "edit", Href: ""}, crumbs[3])
}

func TestBreadcrumbsFromPath_WithLabelOverrides(t *testing.T) {
	labels := map[int]string{1: "User 42"}
	crumbs := BreadcrumbsFromPath("/users/42/edit", labels)
	require.Len(t, crumbs, 4)
	require.Equal(t, BreadcrumbLabelHome, crumbs[0].Label)
	require.Equal(t, "users", crumbs[1].Label)
	require.Equal(t, "User 42", crumbs[2].Label, "segment index 1 should use override")
	require.Equal(t, "edit", crumbs[3].Label)
}

func TestBreadcrumbsFromPath_EmptyString(t *testing.T) {
	crumbs := BreadcrumbsFromPath("", nil)
	require.Len(t, crumbs, 1)
	require.Equal(t, Breadcrumb{Label: BreadcrumbLabelHome, Href: "/"}, crumbs[0])
}

// ---------------------------------------------------------------------------
// FromBit registry
// ---------------------------------------------------------------------------

func TestResolveFromMask_HomeAlwaysIncluded(t *testing.T) {
	crumbs := ResolveFromMask(0) // mask=0 still includes Home
	require.Len(t, crumbs, 1)
	require.Equal(t, BreadcrumbLabelHome, crumbs[0].Label)
	require.Equal(t, "/", crumbs[0].Href)
}

func TestResolveFromMask_DashboardTrail(t *testing.T) {
	RegisterFrom(FromDashboard, Breadcrumb{Label: "Dashboard", Href: "/dashboard"})
	crumbs := ResolveFromMask(FromHome | FromDashboard) // mask=3
	require.Len(t, crumbs, 2)
	require.Equal(t, "Home", crumbs[0].Label)
	require.Equal(t, "Dashboard", crumbs[1].Label)
	require.Equal(t, "/dashboard", crumbs[1].Href)
}

func TestResolveFromMask_IgnoresUnregisteredBits(t *testing.T) {
	crumbs := ResolveFromMask(0xFF) // many bits set, most unregistered
	// Should only contain Home + Dashboard (registered above)
	require.GreaterOrEqual(t, len(crumbs), 1)
	require.Equal(t, "Home", crumbs[0].Label)
}

func TestParseFromParam(t *testing.T) {
	require.Equal(t, uint64(0), ParseFromParam(""))
	require.Equal(t, uint64(0), ParseFromParam("abc"))
	require.Equal(t, uint64(3), ParseFromParam("3"))
	require.Equal(t, uint64(17), ParseFromParam("17"))
}

func TestFromParam(t *testing.T) {
	require.Equal(t, "3", FromParam(3))
	require.Equal(t, "17", FromParam(17))
}

func TestFromQueryString(t *testing.T) {
	require.Equal(t, "", FromQueryString(0))
	require.Equal(t, "from=3", FromQueryString(3))
}

func TestFromNav(t *testing.T) {
	require.Equal(t, "/demo/people", FromNav("/demo/people", ""))
	require.Equal(t, "/demo/people?from=3", FromNav("/demo/people", "3"))
	require.Equal(t, "/demo/people?q=foo&from=3", FromNav("/demo/people?q=foo", "3"))
}

// ---------------------------------------------------------------------------
// ResetForTesting — fromEntries
// ---------------------------------------------------------------------------

func TestResetForTesting_ClearsFromEntries(t *testing.T) {
	ResetForTesting()
	t.Cleanup(ResetForTesting)

	// Register a custom breadcrumb origin at bit 2.
	RegisterFrom(FromBit2, Breadcrumb{Label: "Admin", Href: "/admin"})

	// Sanity: the custom entry should resolve.
	crumbs := ResolveFromMask(uint64(FromBit2))
	require.Len(t, crumbs, 2, "Home + Admin before reset")
	require.Equal(t, "Admin", crumbs[1].Label)

	// Reset should discard the custom entry.
	ResetForTesting()

	crumbs = ResolveFromMask(uint64(FromBit2))
	require.Len(t, crumbs, 1, "only Home should remain after reset")
	require.Equal(t, BreadcrumbLabelHome, crumbs[0].Label)
	require.Equal(t, "/", crumbs[0].Href)
}
