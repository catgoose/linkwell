
package linkwell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetLinks(t *testing.T) {
	t.Helper()
	t.Cleanup(ResetForTesting)
	ResetForTesting()
}

// ---------------------------------------------------------------------------
// TitleFromPath
// ---------------------------------------------------------------------------

func TestTitleFromPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"nested path", "/demo/inventory", "Inventory"},
		{"hyphenated segment", "/admin/error-traces", "Error Traces"},
		{"root", "/", ""},
		{"single segment", "/single", "Single"},
		{"trailing slash trimmed", "/trailing/slash/", "Slash"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, TitleFromPath(tt.path))
		})
	}
}

// ---------------------------------------------------------------------------
// Link (symmetric vs non-symmetric)
// ---------------------------------------------------------------------------

func TestLink_Symmetric(t *testing.T) {
	resetLinks(t)

	Link("a", "related", "b", "B")

	aLinks := LinksFor("a")
	require.Len(t, aLinks, 1)
	assert.Equal(t, "b", aLinks[0].Href)
	assert.Equal(t, "related", aLinks[0].Rel)
	assert.Equal(t, "B", aLinks[0].Title)

	// Inverse should be auto-registered
	bLinks := LinksFor("b")
	require.Len(t, bLinks, 1)
	assert.Equal(t, "a", bLinks[0].Href)
	assert.Equal(t, "related", bLinks[0].Rel)
}

func TestLink_NonSymmetric(t *testing.T) {
	resetLinks(t)

	Link("a", "up", "b", "B")

	aLinks := LinksFor("a")
	require.Len(t, aLinks, 1)
	assert.Equal(t, "b", aLinks[0].Href)

	// rel="up" is not symmetric — "b" should have no links
	bLinks := LinksFor("b")
	assert.Empty(t, bLinks)
}

// ---------------------------------------------------------------------------
// LinksFor
// ---------------------------------------------------------------------------

func TestLinksFor_AllRels(t *testing.T) {
	resetLinks(t)

	Link("/x", "related", "/y", "Y")
	Link("/x", "up", "/z", "Z")

	all := LinksFor("/x")
	assert.Len(t, all, 2)
}

func TestLinksFor_FilterByRel(t *testing.T) {
	resetLinks(t)

	Link("/x", "related", "/y", "Y")
	Link("/x", "up", "/z", "Z")

	upOnly := LinksFor("/x", "up")
	require.Len(t, upOnly, 1)
	assert.Equal(t, "/z", upOnly[0].Href)
}

func TestLinksFor_UnknownPath(t *testing.T) {
	resetLinks(t)

	links := LinksFor("/nonexistent")
	assert.Empty(t, links)
}

func TestLinksFor_ReturnsCopy(t *testing.T) {
	resetLinks(t)

	Link("/x", "related", "/y", "Y")

	result := LinksFor("/x")
	result[0].Title = "MUTATED"

	original := LinksFor("/x")
	assert.Equal(t, "Y", original[0].Title, "modifying returned slice must not affect registry")
}

// ---------------------------------------------------------------------------
// RelatedLinksFor
// ---------------------------------------------------------------------------

func TestRelatedLinksFor_OnlyRelated(t *testing.T) {
	resetLinks(t)

	Link("/a", "related", "/b", "B")
	Link("/a", "up", "/c", "C")

	related := RelatedLinksFor("/a")
	require.Len(t, related, 1)
	assert.Equal(t, "/b", related[0].Href)
}

func TestRelatedLinksFor_ExcludesSelf(t *testing.T) {
	resetLinks(t)

	// Manually insert a self-referencing related link
	linksMu.Lock()
	linksMap["/a"] = append(linksMap["/a"],
		LinkRelation{Rel: "related", Href: "/a", Title: "Self"},
		LinkRelation{Rel: "related", Href: "/b", Title: "B"},
	)
	linksMu.Unlock()

	related := RelatedLinksFor("/a")
	require.Len(t, related, 1)
	assert.Equal(t, "/b", related[0].Href)
}

func TestRelatedLinksFor_DeduplicatesByHref(t *testing.T) {
	resetLinks(t)

	linksMu.Lock()
	linksMap["/a"] = []LinkRelation{
		{Rel: "related", Href: "/b", Title: "B"},
		{Rel: "related", Href: "/b", Title: "B duplicate"},
	}
	linksMu.Unlock()

	related := RelatedLinksFor("/a")
	require.Len(t, related, 1)
	assert.Equal(t, "/b", related[0].Href)
}

// ---------------------------------------------------------------------------
// LinkHeader
// ---------------------------------------------------------------------------

func TestLinkHeader_Empty(t *testing.T) {
	assert.Equal(t, "", LinkHeader(nil))
}

func TestLinkHeader_Single(t *testing.T) {
	links := []LinkRelation{{Rel: "related", Href: "/b", Title: "B"}}
	got := LinkHeader(links)
	assert.Equal(t, `</b>; rel="related"; title="B"`, got)
}

func TestLinkHeader_Multiple(t *testing.T) {
	links := []LinkRelation{
		{Rel: "related", Href: "/b", Title: "B"},
		{Rel: "up", Href: "/c", Title: "C"},
	}
	got := LinkHeader(links)
	assert.Contains(t, got, `</b>; rel="related"; title="B"`)
	assert.Contains(t, got, `</c>; rel="up"; title="C"`)
	assert.Contains(t, got, ", ")
}

func TestLinkHeader_EmptyTitle(t *testing.T) {
	links := []LinkRelation{{Rel: "related", Href: "/b", Title: ""}}
	got := LinkHeader(links)
	assert.Equal(t, `</b>; rel="related"`, got)
	assert.NotContains(t, got, "title")
}

func TestLinkHeader_EscapesQuotesInTitle(t *testing.T) {
	links := []LinkRelation{{Rel: "related", Href: "/b", Title: `She said "hello"`}}
	got := LinkHeader(links)
	assert.Equal(t, `</b>; rel="related"; title="She said \"hello\""`, got)
}

func TestLinkHeader_EscapesBackslashInTitle(t *testing.T) {
	links := []LinkRelation{{Rel: "related", Href: "/b", Title: `path\to\file`}}
	got := LinkHeader(links)
	assert.Equal(t, `</b>; rel="related"; title="path\\to\\file"`, got)
}

func TestRelConstants(t *testing.T) {
	assert.Equal(t, "related", RelRelated)
	assert.Equal(t, "up", RelUp)
	assert.Equal(t, "self", RelSelf)
	assert.Equal(t, "next", RelNext)
	assert.Equal(t, "prev", RelPrev)
}

// ---------------------------------------------------------------------------
// Rel
// ---------------------------------------------------------------------------

func TestRel(t *testing.T) {
	entry := Rel("/demo/widgets", "Widgets")
	assert.Equal(t, "/demo/widgets", entry.Path)
	assert.Equal(t, "Widgets", entry.Title)
}

// ---------------------------------------------------------------------------
// Ring
// ---------------------------------------------------------------------------

func TestRing_ThreeMembers(t *testing.T) {
	resetLinks(t)

	Ring("group1",
		Rel("/a", "A"),
		Rel("/b", "B"),
		Rel("/c", "C"),
	)

	all := AllLinks()
	// 3 members, each links to 2 others = 6 total links
	total := 0
	for _, v := range all {
		total += len(v)
	}
	assert.Equal(t, 6, total)

	// Verify group name
	for _, links := range all {
		for _, l := range links {
			assert.Equal(t, "group1", l.Group)
		}
	}

	// No self-links
	for path, links := range all {
		for _, l := range links {
			assert.NotEqual(t, path, l.Href, "ring must not create self-links")
		}
	}
}

func TestRing_NoDuplicatesOnRepeat(t *testing.T) {
	resetLinks(t)

	members := []RelEntry{Rel("/a", "A"), Rel("/b", "B")}
	Ring("g", members...)
	Ring("g", members...)

	aLinks := LinksFor("/a")
	assert.Len(t, aLinks, 1, "duplicate Ring call must not create duplicate links")
}

// ---------------------------------------------------------------------------
// Hub
// ---------------------------------------------------------------------------

func TestHub_CenterToSpokes(t *testing.T) {
	resetLinks(t)

	Hub("/center", "Center",
		Rel("/s1", "Spoke One"),
		Rel("/s2", "Spoke Two"),
	)

	centerLinks := LinksFor("/center", "related")
	require.Len(t, centerLinks, 2)
	hrefs := map[string]bool{}
	for _, l := range centerLinks {
		hrefs[l.Href] = true
		assert.Equal(t, "Center", l.Group)
	}
	assert.True(t, hrefs["/s1"])
	assert.True(t, hrefs["/s2"])
}

func TestHub_SpokesToCenter(t *testing.T) {
	resetLinks(t)

	Hub("/center", "Center",
		Rel("/s1", "Spoke One"),
		Rel("/s2", "Spoke Two"),
	)

	for _, spoke := range []string{"/s1", "/s2"} {
		upLinks := LinksFor(spoke, "up")
		require.Len(t, upLinks, 1, "spoke %s must have exactly one up link", spoke)
		assert.Equal(t, "/center", upLinks[0].Href)
		assert.Equal(t, "Center", upLinks[0].Title)
		assert.Equal(t, "Center", upLinks[0].Group)
	}
}

func TestHub_SpokesDoNotLinkToEachOther(t *testing.T) {
	resetLinks(t)

	Hub("/center", "Center",
		Rel("/s1", "Spoke One"),
		Rel("/s2", "Spoke Two"),
	)

	s1Links := LinksFor("/s1")
	for _, l := range s1Links {
		assert.NotEqual(t, "/s2", l.Href, "spokes must not link to each other")
	}
}

func TestHub_CenterInHubsMap(t *testing.T) {
	resetLinks(t)

	Hub("/center", "Center", Rel("/s1", "S1"))

	hubs := Hubs()
	require.Len(t, hubs, 1)
	assert.Equal(t, "/center", hubs[0].Path)
	assert.Equal(t, "Center", hubs[0].Title)
}

// ---------------------------------------------------------------------------
// AllLinks
// ---------------------------------------------------------------------------

func TestAllLinks_ReturnsCopy(t *testing.T) {
	resetLinks(t)

	Link("/a", "related", "/b", "B")

	result := AllLinks()
	result["/a"][0].Title = "MUTATED"
	delete(result, "/a")

	original := AllLinks()
	require.Contains(t, original, "/a")
	assert.Equal(t, "B", original["/a"][0].Title,
		"modifying returned map must not affect registry")
}

// ---------------------------------------------------------------------------
// SortedPaths
// ---------------------------------------------------------------------------

func TestSortedPaths(t *testing.T) {
	resetLinks(t)

	Link("/z", "related", "/a", "A")
	Link("/m", "up", "/a", "A")

	all := AllLinks()
	paths := sortedPaths(all)
	// /a gets auto-registered from symmetric "related"
	assert.True(t, len(paths) >= 3)
	for i := 1; i < len(paths); i++ {
		assert.True(t, paths[i-1] <= paths[i], "paths must be sorted: %s > %s", paths[i-1], paths[i])
	}
}

// ---------------------------------------------------------------------------
// Hubs
// ---------------------------------------------------------------------------

func TestHubs_SortedByPath(t *testing.T) {
	resetLinks(t)

	Hub("/z-hub", "Z Hub", Rel("/zs", "ZS"))
	Hub("/a-hub", "A Hub", Rel("/as", "AS"))

	hubs := Hubs()
	require.Len(t, hubs, 2)
	assert.Equal(t, "/a-hub", hubs[0].Path)
	assert.Equal(t, "/z-hub", hubs[1].Path)
}

func TestHubs_SpokesSortedByTitle(t *testing.T) {
	resetLinks(t)

	Hub("/h", "H",
		Rel("/z", "Zebra"),
		Rel("/a", "Apple"),
		Rel("/m", "Mango"),
	)

	hubs := Hubs()
	require.Len(t, hubs, 1)
	require.Len(t, hubs[0].Spokes, 3)
	assert.Equal(t, "Apple", hubs[0].Spokes[0].Title)
	assert.Equal(t, "Mango", hubs[0].Spokes[1].Title)
	assert.Equal(t, "Zebra", hubs[0].Spokes[2].Title)
}

// ---------------------------------------------------------------------------
// BreadcrumbsFromLinks
// ---------------------------------------------------------------------------

func TestBreadcrumbsFromLinks_SpokeToHub(t *testing.T) {
	resetLinks(t)

	Hub("/hub", "Hub Page", Rel("/hub/spoke", "Spoke Page"))

	crumbs := BreadcrumbsFromLinks("/hub/spoke")
	require.NotNil(t, crumbs)
	// [Home, Hub Page, Spoke]
	require.Len(t, crumbs, 3)
	assert.Equal(t, BreadcrumbLabelHome, crumbs[0].Label)
	assert.Equal(t, "/", crumbs[0].Href)
	assert.Equal(t, "Hub Page", crumbs[1].Label)
	assert.Equal(t, "/hub", crumbs[1].Href)
	assert.Equal(t, "Spoke", crumbs[2].Label)
	assert.Empty(t, crumbs[2].Href, "current page should not be a link")
}

func TestBreadcrumbsFromLinks_NoUpLinks(t *testing.T) {
	resetLinks(t)

	crumbs := BreadcrumbsFromLinks("/standalone")
	assert.Nil(t, crumbs)
}

func TestBreadcrumbsFromLinks_CycleDetection(t *testing.T) {
	resetLinks(t)

	// Create a cycle: /a -> /b -> /a
	linksMu.Lock()
	linksMap["/a"] = []LinkRelation{{Rel: "up", Href: "/b", Title: "B"}}
	linksMap["/b"] = []LinkRelation{{Rel: "up", Href: "/a", Title: "A"}}
	linksMu.Unlock()

	// Must not hang — cycle detection should break the loop
	crumbs := BreadcrumbsFromLinks("/a")
	require.NotNil(t, crumbs)
	// Should terminate; exact length depends on where cycle is broken
	assert.True(t, len(crumbs) > 0, "breadcrumbs must not be empty after cycle")
}

func TestBreadcrumbsFromLinks_IncludesHomeAndCurrent(t *testing.T) {
	resetLinks(t)

	linksMu.Lock()
	linksMap["/child"] = []LinkRelation{{Rel: "up", Href: "/parent", Title: "Parent"}}
	linksMu.Unlock()

	crumbs := BreadcrumbsFromLinks("/child")
	require.NotNil(t, crumbs)
	assert.Equal(t, BreadcrumbLabelHome, crumbs[0].Label, "first crumb should be Home")
	last := crumbs[len(crumbs)-1]
	assert.Equal(t, "Child", last.Label, "last crumb should be current page title")
	assert.Empty(t, last.Href, "last crumb should not be a link")
}

// ---------------------------------------------------------------------------
// BreadcrumbsFromLinks — path walk-up (#13)
// ---------------------------------------------------------------------------

func TestBreadcrumbsFromLinks_WalkUpOneSegment(t *testing.T) {
	resetLinks(t)

	Hub("/admin", "Admin",
		Rel("/admin/groups", "Groups"),
	)

	crumbs := BreadcrumbsFromLinks("/admin/groups/new")
	require.NotNil(t, crumbs)
	// [Home, Admin, Groups, New]
	require.Len(t, crumbs, 4)
	assert.Equal(t, BreadcrumbLabelHome, crumbs[0].Label)
	assert.Equal(t, "/", crumbs[0].Href)
	assert.Equal(t, "Admin", crumbs[1].Label)
	assert.Equal(t, "/admin", crumbs[1].Href)
	assert.Equal(t, "Groups", crumbs[2].Label)
	assert.Equal(t, "/admin/groups", crumbs[2].Href)
	assert.Equal(t, "New", crumbs[3].Label)
	assert.Empty(t, crumbs[3].Href, "terminal segment should have no href")
}

func TestBreadcrumbsFromLinks_WalkUpNumericSegment(t *testing.T) {
	resetLinks(t)

	Hub("/admin", "Admin",
		Rel("/admin/groups", "Groups"),
	)

	crumbs := BreadcrumbsFromLinks("/admin/groups/1")
	require.NotNil(t, crumbs)
	// [Home, Admin, Groups, 1]
	require.Len(t, crumbs, 4)
	assert.Equal(t, "Groups", crumbs[2].Label)
	assert.Equal(t, "/admin/groups", crumbs[2].Href)
	assert.Equal(t, "1", crumbs[3].Label)
	assert.Empty(t, crumbs[3].Href)
}

func TestBreadcrumbsFromLinks_WalkUpMultipleSegments(t *testing.T) {
	resetLinks(t)

	Hub("/admin", "Admin",
		Rel("/admin/groups", "Groups"),
	)

	crumbs := BreadcrumbsFromLinks("/admin/groups/1/edit")
	require.NotNil(t, crumbs)
	// [Home, Admin, Groups, 1, Edit]
	require.Len(t, crumbs, 5)
	assert.Equal(t, "Groups", crumbs[2].Label)
	assert.Equal(t, "/admin/groups", crumbs[2].Href)
	assert.Equal(t, "1", crumbs[3].Label)
	assert.Equal(t, "/admin/groups/1", crumbs[3].Href)
	assert.Equal(t, "Edit", crumbs[4].Label)
	assert.Empty(t, crumbs[4].Href)
}

func TestBreadcrumbsFromLinks_WalkUpNoAncestorReturnsNil(t *testing.T) {
	resetLinks(t)

	crumbs := BreadcrumbsFromLinks("/totally/unknown/path")
	assert.Nil(t, crumbs)
}

// ---------------------------------------------------------------------------
// ResolveFromMaskWithPath (#14)
// ---------------------------------------------------------------------------

func TestResolveFromMaskWithPath_CombinesMaskAndPath(t *testing.T) {
	resetLinks(t)

	RegisterFrom(FromDashboard, Breadcrumb{Label: "Dashboard", Href: "/dashboard"})
	Hub("/admin", "Admin",
		Rel("/admin/groups", "Groups"),
	)

	crumbs := ResolveFromMaskWithPath(FromHome|FromDashboard, "/admin/groups", "3")
	require.NotNil(t, crumbs)
	// Mask crumbs: [Home, Dashboard]
	// Path crumbs: [Home, Admin, Groups] — Home is deduplicated
	// Result: [Home, Dashboard, Admin, Groups]
	require.Len(t, crumbs, 4)
	assert.Equal(t, "Home", crumbs[0].Label)
	assert.Equal(t, "Dashboard", crumbs[1].Label)
	assert.Equal(t, "Admin", crumbs[2].Label)
	assert.Contains(t, crumbs[2].Href, "from=3", "intermediate crumbs should have from param")
	assert.Equal(t, "Groups", crumbs[3].Label)
	assert.Empty(t, crumbs[3].Href, "terminal crumb should have no href")
}

func TestResolveFromMaskWithPath_DeduplicatesByBasePath(t *testing.T) {
	resetLinks(t)

	RegisterFrom(FromDashboard, Breadcrumb{Label: "Dashboard", Href: "/dashboard?tab=overview"})
	Hub("/dashboard", "Dashboard",
		Rel("/dashboard/settings", "Settings"),
	)

	crumbs := ResolveFromMaskWithPath(FromHome|FromDashboard, "/dashboard/settings", "")
	// Mask: [Home, Dashboard (with query)]
	// Path: [Home, Dashboard, Settings] — Home and Dashboard deduplicated by base path
	// Result: [Home, Dashboard (with query), Settings]
	require.Len(t, crumbs, 3)
	assert.Equal(t, "Home", crumbs[0].Label)
	assert.Equal(t, "Dashboard", crumbs[1].Label)
	assert.Equal(t, "/dashboard?tab=overview", crumbs[1].Href, "mask crumb should keep its query params")
	assert.Equal(t, "Settings", crumbs[2].Label)
}

func TestResolveFromMaskWithPath_FallsBackToPathBreadcrumbs(t *testing.T) {
	resetLinks(t)

	// No links registered — should fall back to BreadcrumbsFromPath
	crumbs := ResolveFromMaskWithPath(FromHome, "/users/42/edit", "")
	require.NotNil(t, crumbs)
	// Mask: [Home]
	// Fallback path: [Home, users, 42, edit] — Home deduplicated
	require.Len(t, crumbs, 4)
	assert.Equal(t, "Home", crumbs[0].Label)
	assert.Equal(t, "users", crumbs[1].Label)
	assert.Equal(t, "42", crumbs[2].Label)
	assert.Equal(t, "edit", crumbs[3].Label)
}

func TestResolveFromMaskWithPath_NoFromParam(t *testing.T) {
	resetLinks(t)

	Hub("/admin", "Admin",
		Rel("/admin/groups", "Groups"),
	)

	crumbs := ResolveFromMaskWithPath(FromHome, "/admin/groups", "")
	require.NotNil(t, crumbs)
	// Should not contain from= in any href
	for _, c := range crumbs {
		assert.NotContains(t, c.Href, "from=", "no from param should be appended when from is empty")
	}
}

// ---------------------------------------------------------------------------
// LoadStoredLink
// ---------------------------------------------------------------------------

func TestLoadStoredLink_AddsLink(t *testing.T) {
	resetLinks(t)

	LoadStoredLink("/src", LinkRelation{Rel: "related", Href: "/tgt", Title: "Target"})

	links := LinksFor("/src")
	require.Len(t, links, 1)
	assert.Equal(t, "/tgt", links[0].Href)
}

func TestLoadStoredLink_SkipsDuplicates(t *testing.T) {
	resetLinks(t)

	lr := LinkRelation{Rel: "related", Href: "/tgt", Title: "Target"}
	LoadStoredLink("/src", lr)
	LoadStoredLink("/src", lr)

	links := LinksFor("/src")
	assert.Len(t, links, 1)
}

// ---------------------------------------------------------------------------
// RemoveLink
// ---------------------------------------------------------------------------

func TestRemoveLink_RemovesExisting(t *testing.T) {
	resetLinks(t)

	LoadStoredLink("/src", LinkRelation{Rel: "related", Href: "/tgt", Title: "T"})

	ok := RemoveLink("/src", "/tgt", "related")
	assert.True(t, ok)
	assert.Empty(t, LinksFor("/src"))
}

func TestRemoveLink_ReturnsFalseWhenNotFound(t *testing.T) {
	resetLinks(t)

	ok := RemoveLink("/src", "/tgt", "related")
	assert.False(t, ok)
}

func TestRemoveLink_CleansUpEmptyEntry(t *testing.T) {
	resetLinks(t)

	LoadStoredLink("/src", LinkRelation{Rel: "related", Href: "/tgt", Title: "T"})
	RemoveLink("/src", "/tgt", "related")

	all := AllLinks()
	_, exists := all["/src"]
	assert.False(t, exists, "empty map entries must be cleaned up")
}

// ---------------------------------------------------------------------------
// LinksFor — parent path walk-up (#19)
// ---------------------------------------------------------------------------

func TestLinksFor_ChildPath(t *testing.T) {
	resetLinks(t)

	Ring("admin-apps",
		Rel("/admin/apps", "Apps"),
		Rel("/admin/users", "Users"),
	)

	// /admin/apps/1 is not registered — should inherit links from /admin/apps
	links := LinksFor("/admin/apps/1")
	require.NotEmpty(t, links, "child path should return parent's links")
	hrefs := make(map[string]bool)
	for _, l := range links {
		hrefs[l.Href] = true
	}
	assert.True(t, hrefs["/admin/users"], "expected related link to /admin/users via parent")
}

func TestLinksFor_DeepChildPath(t *testing.T) {
	resetLinks(t)

	Ring("admin-apps",
		Rel("/admin/apps", "Apps"),
		Rel("/admin/users", "Users"),
	)

	// /admin/apps/1/edit — two segments deep, should walk up to /admin/apps
	links := LinksFor("/admin/apps/1/edit")
	require.NotEmpty(t, links, "deeply nested child path should return ancestor links")
	hrefs := make(map[string]bool)
	for _, l := range links {
		hrefs[l.Href] = true
	}
	assert.True(t, hrefs["/admin/users"], "expected related link to /admin/users via ancestor")
}

func TestLinksFor_ExactMatchStillWorks(t *testing.T) {
	resetLinks(t)

	Ring("admin-apps",
		Rel("/admin/apps", "Apps"),
		Rel("/admin/users", "Users"),
	)

	// Exact match for /admin/apps should return its own links, not a parent's
	links := LinksFor("/admin/apps")
	require.NotEmpty(t, links)
	for _, l := range links {
		assert.NotEqual(t, "/admin/apps", l.Href, "should not contain self-link")
	}
}

// ---------------------------------------------------------------------------
// RelatedLinksFor — parent path walk-up (#19)
// ---------------------------------------------------------------------------

func TestRelatedLinksFor_ChildPath(t *testing.T) {
	resetLinks(t)

	Ring("admin-apps",
		Rel("/admin/apps", "Apps"),
		Rel("/admin/users", "Users"),
	)

	// /admin/apps/1 is not registered — should inherit related links from /admin/apps
	related := RelatedLinksFor("/admin/apps/1")
	require.NotEmpty(t, related, "child path should return parent's related links")
	hrefs := make(map[string]bool)
	for _, l := range related {
		hrefs[l.Href] = true
	}
	assert.True(t, hrefs["/admin/users"], "expected /admin/users in related links via parent")
}
