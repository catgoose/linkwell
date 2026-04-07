package linkwell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Sitemap
// ---------------------------------------------------------------------------

func TestSitemap_Empty(t *testing.T) {
	resetLinks(t)

	entries := Sitemap()
	assert.Empty(t, entries)
}

func TestSitemap_HubWithSpokes(t *testing.T) {
	resetLinks(t)

	Hub("/admin", "Admin",
		Rel("/admin/users", "Users"),
		Rel("/admin/groups", "Groups"),
	)

	entries := Sitemap()
	require.NotEmpty(t, entries)

	// Build a lookup by path for easier assertions.
	byPath := make(map[string]SitemapEntry, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e
	}

	// Hub center should have children and no parent.
	admin := byPath["/admin"]
	assert.Equal(t, "Admin", admin.Title)
	assert.Empty(t, admin.Parent)
	assert.ElementsMatch(t, []string{"/admin/groups", "/admin/users"}, admin.Children)

	// Spokes should have parent pointing to hub center.
	users := byPath["/admin/users"]
	assert.Equal(t, "Users", users.Title)
	assert.Equal(t, "/admin", users.Parent)
	assert.Nil(t, users.Children)

	groups := byPath["/admin/groups"]
	assert.Equal(t, "Groups", groups.Title)
	assert.Equal(t, "/admin", groups.Parent)
	assert.Nil(t, groups.Children)
}

func TestSitemap_RingGroup(t *testing.T) {
	resetLinks(t)

	Ring("reports",
		Rel("/reports/daily", "Daily"),
		Rel("/reports/weekly", "Weekly"),
		Rel("/reports/monthly", "Monthly"),
	)

	entries := Sitemap()
	require.NotEmpty(t, entries)

	for _, e := range entries {
		assert.Equal(t, "reports", e.Group, "all ring members should have group set")
		assert.Empty(t, e.Parent, "ring members have no parent")
		assert.Nil(t, e.Children, "ring members are not hub centers")
	}
}

func TestSitemap_SortedByPath(t *testing.T) {
	resetLinks(t)

	Hub("/z-section", "Z Section", Rel("/z-section/page", "Page"))
	Hub("/a-section", "A Section", Rel("/a-section/page", "Page"))

	entries := Sitemap()
	require.True(t, len(entries) >= 4)

	for i := 1; i < len(entries); i++ {
		assert.True(t, entries[i-1].Path <= entries[i].Path,
			"entries must be sorted by path: %s > %s", entries[i-1].Path, entries[i].Path)
	}
}

func TestSitemap_MixedTopology(t *testing.T) {
	resetLinks(t)

	Hub("/admin", "Admin",
		Rel("/admin/users", "Users"),
	)
	Ring("tools",
		Rel("/tools/import", "Import"),
		Rel("/tools/export", "Export"),
	)
	Link("/about", RelRelated, "/contact", "Contact")

	entries := Sitemap()
	require.NotEmpty(t, entries)

	byPath := make(map[string]SitemapEntry, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e
	}

	// Hub center
	assert.NotEmpty(t, byPath["/admin"].Children)
	// Hub spoke
	assert.Equal(t, "/admin", byPath["/admin/users"].Parent)
	// Ring members
	assert.Equal(t, "tools", byPath["/tools/import"].Group)
	assert.Equal(t, "tools", byPath["/tools/export"].Group)
	// Plain related links
	assert.Contains(t, byPath, "/about")
	assert.Contains(t, byPath, "/contact")
}

func TestSitemap_TitleDerivedFromPath(t *testing.T) {
	resetLinks(t)

	Link("/demo/my-feature", RelUp, "/demo", "Demo")

	entries := Sitemap()
	byPath := make(map[string]SitemapEntry, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e
	}

	assert.Equal(t, "My Feature", byPath["/demo/my-feature"].Title)
}

func TestSitemap_HubTitlePreferred(t *testing.T) {
	resetLinks(t)

	Hub("/dashboard", "Main Dashboard",
		Rel("/dashboard/stats", "Statistics"),
	)

	entries := Sitemap()
	byPath := make(map[string]SitemapEntry, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e
	}

	assert.Equal(t, "Main Dashboard", byPath["/dashboard"].Title,
		"hub title should be preferred over TitleFromPath")
}

func TestSitemap_TargetOnlyPages(t *testing.T) {
	resetLinks(t)

	Link("/source", RelUp, "/target", "Custom Target")

	entries := Sitemap()
	byPath := make(map[string]SitemapEntry, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e
	}

	// Target-only page must appear in the sitemap.
	require.Contains(t, byPath, "/target")
	assert.Equal(t, "Custom Target", byPath["/target"].Title,
		"target-only page should use registered title")

	// Source page should also be present.
	require.Contains(t, byPath, "/source")
	assert.Equal(t, "/target", byPath["/source"].Parent,
		"source should have parent set from rel=up")
}

func TestSitemap_RegisteredTitlePreferred(t *testing.T) {
	resetLinks(t)

	// Register a link with a custom title for the target.
	Link("/docs", RelRelated, "/docs/api", "API Reference")

	entries := Sitemap()
	byPath := make(map[string]SitemapEntry, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e
	}

	// The target should use the registered title, not TitleFromPath.
	require.Contains(t, byPath, "/docs/api")
	assert.Equal(t, "API Reference", byPath["/docs/api"].Title,
		"registered title should be preferred over TitleFromPath")
}

func TestSitemap_HubSpokeTitlePreserved(t *testing.T) {
	resetLinks(t)

	Hub("/settings", "Settings",
		Rel("/settings/profile", "My Profile"),
		Rel("/settings/security", "Security Options"),
	)

	entries := Sitemap()
	byPath := make(map[string]SitemapEntry, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e
	}

	assert.Equal(t, "My Profile", byPath["/settings/profile"].Title,
		"hub spoke should preserve its registered title")
	assert.Equal(t, "Security Options", byPath["/settings/security"].Title,
		"hub spoke should preserve its registered title")
}

func TestSitemap_MixedSourceAndTargetPaths(t *testing.T) {
	resetLinks(t)

	// /a links to /b (target-only) and /c links to /a (so /a is both source and target).
	Link("/a", RelRelated, "/b", "Page B")
	Link("/c", RelUp, "/a", "Page A")

	entries := Sitemap()
	byPath := make(map[string]SitemapEntry, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e
	}

	require.Contains(t, byPath, "/a")
	require.Contains(t, byPath, "/b")
	require.Contains(t, byPath, "/c")

	// /a is a target of /c with title "Page A".
	assert.Equal(t, "Page A", byPath["/a"].Title,
		"path that is both source and target should use registered title")
	// /b is target-only with title "Page B".
	assert.Equal(t, "Page B", byPath["/b"].Title)
	// /c is source-only, parent is /a.
	assert.Equal(t, "/a", byPath["/c"].Parent)
}

// ---------------------------------------------------------------------------
// SitemapRoots
// ---------------------------------------------------------------------------

func TestSitemapRoots_Empty(t *testing.T) {
	resetLinks(t)

	roots := SitemapRoots()
	assert.Empty(t, roots)
}

func TestSitemapRoots_OnlyRoots(t *testing.T) {
	resetLinks(t)

	Hub("/admin", "Admin",
		Rel("/admin/users", "Users"),
		Rel("/admin/groups", "Groups"),
	)
	Ring("tools",
		Rel("/tools/import", "Import"),
		Rel("/tools/export", "Export"),
	)

	roots := SitemapRoots()
	require.NotEmpty(t, roots)

	for _, r := range roots {
		assert.Empty(t, r.Parent, "root entries must have no parent")
	}

	// Hub center is a root; spokes are not.
	rootPaths := make(map[string]bool, len(roots))
	for _, r := range roots {
		rootPaths[r.Path] = true
	}
	assert.True(t, rootPaths["/admin"], "hub center should be a root")
	assert.False(t, rootPaths["/admin/users"], "spoke should not be a root")
	assert.False(t, rootPaths["/admin/groups"], "spoke should not be a root")
	// Ring members have no parent, so they are roots.
	assert.True(t, rootPaths["/tools/import"], "ring member should be a root")
	assert.True(t, rootPaths["/tools/export"], "ring member should be a root")
}

func TestSitemapRoots_SortedByPath(t *testing.T) {
	resetLinks(t)

	Ring("nav",
		Rel("/z-page", "Z"),
		Rel("/a-page", "A"),
		Rel("/m-page", "M"),
	)

	roots := SitemapRoots()
	require.True(t, len(roots) >= 3)

	for i := 1; i < len(roots); i++ {
		assert.True(t, roots[i-1].Path <= roots[i].Path,
			"roots must be sorted by path: %s > %s", roots[i-1].Path, roots[i].Path)
	}
}
