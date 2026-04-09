package linkwell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ValidateGraph
// ---------------------------------------------------------------------------

func TestValidateGraph_Clean(t *testing.T) {
	resetLinks(t)

	Hub("/admin", "Admin",
		Rel("/admin/users", "Users"),
		Rel("/admin/groups", "Groups"),
	)

	issues := ValidateGraph()
	assert.Empty(t, issues, "well-formed hub should produce no issues")
}

func TestValidateGraph_Orphan(t *testing.T) {
	resetLinks(t)

	// /orphan links out but nothing links to it
	Link("/orphan", "up", "/parent", "Parent")
	// Register /parent so it's not flagged as broken_up target
	Link("/parent", "related", "/other", "Other")

	issues := ValidateGraph()
	var orphans []LinkIssue
	for _, i := range issues {
		if i.Kind == "orphan" && i.Path == "/orphan" {
			orphans = append(orphans, i)
		}
	}
	require.Len(t, orphans, 1)
	assert.Contains(t, orphans[0].Message, "no inbound links")
}

func TestValidateGraph_SelfLinkIsOrphan(t *testing.T) {
	resetLinks(t)

	// /foo's only "inbound" link is a self-link — it should still be
	// flagged as an orphan because self-links don't count as inbound.
	linksMu.Lock()
	linksMap["/foo"] = []LinkRelation{
		{Rel: RelRelated, Href: "/foo", Title: "Foo"},
	}
	linksMu.Unlock()

	issues := ValidateGraph()
	var orphans []LinkIssue
	for _, i := range issues {
		if i.Kind == "orphan" && i.Path == "/foo" {
			orphans = append(orphans, i)
		}
	}
	require.Len(t, orphans, 1, "self-linked page should be flagged as orphan")
	assert.Contains(t, orphans[0].Message, "no inbound links")
}

func TestValidateGraph_BrokenUp(t *testing.T) {
	resetLinks(t)

	// /child has rel="up" pointing to /missing which is not registered
	linksMu.Lock()
	linksMap["/child"] = []LinkRelation{
		{Rel: RelUp, Href: "/missing", Title: "Missing"},
	}
	linksMu.Unlock()

	issues := ValidateGraph()
	var broken []LinkIssue
	for _, i := range issues {
		if i.Kind == "broken_up" {
			broken = append(broken, i)
		}
	}
	require.Len(t, broken, 1)
	assert.Equal(t, "/child", broken[0].Path)
	assert.Contains(t, broken[0].Message, "/missing")
}

func TestValidateGraph_DeadSpoke(t *testing.T) {
	resetLinks(t)

	Hub("/hub", "Hub", Rel("/spoke", "Spoke"))

	// Remove all links from /spoke to simulate a dead spoke
	linksMu.Lock()
	delete(linksMap, "/spoke")
	linksMu.Unlock()

	issues := ValidateGraph()
	var deadSpokes []LinkIssue
	for _, i := range issues {
		if i.Kind == "dead_spoke" {
			deadSpokes = append(deadSpokes, i)
		}
	}
	require.Len(t, deadSpokes, 1)
	assert.Equal(t, "/spoke", deadSpokes[0].Path)
	assert.Contains(t, deadSpokes[0].Message, "Hub")
}

func TestValidateGraph_EmptyRegistry(t *testing.T) {
	resetLinks(t)

	issues := ValidateGraph()
	assert.Empty(t, issues, "empty registry should produce no issues")
}

func TestValidateGraph_Ring_NoOrphans(t *testing.T) {
	resetLinks(t)

	Ring("group",
		Rel("/a", "A"),
		Rel("/b", "B"),
		Rel("/c", "C"),
	)

	issues := ValidateGraph()
	var orphans []LinkIssue
	for _, i := range issues {
		if i.Kind == "orphan" {
			orphans = append(orphans, i)
		}
	}
	assert.Empty(t, orphans, "ring members should all have inbound links")
}

func TestValidateGraph_MultipleIssues(t *testing.T) {
	resetLinks(t)

	// Orphan with broken up chain
	linksMu.Lock()
	linksMap["/orphan-broken"] = []LinkRelation{
		{Rel: RelUp, Href: "/nonexistent", Title: "Gone"},
	}
	linksMu.Unlock()

	issues := ValidateGraph()
	require.True(t, len(issues) >= 2, "expected at least orphan + broken_up issues")

	kinds := make(map[string]bool)
	for _, i := range issues {
		kinds[i.Kind] = true
	}
	assert.True(t, kinds["orphan"])
	assert.True(t, kinds["broken_up"])
}

// ---------------------------------------------------------------------------
// ValidateAgainstRoutes
// ---------------------------------------------------------------------------

func TestValidateAgainstRoutes_AllMatch(t *testing.T) {
	resetLinks(t)

	Hub("/admin", "Admin",
		Rel("/admin/users", "Users"),
	)

	routes := []string{"/admin", "/admin/users"}
	issues := ValidateAgainstRoutes(routes)
	assert.Empty(t, issues, "all paths match routes — no issues expected")
}

func TestValidateAgainstRoutes_UnregisteredRoute(t *testing.T) {
	resetLinks(t)

	// Register a path that won't be in the route list
	linksMu.Lock()
	linksMap["/phantom"] = []LinkRelation{
		{Rel: RelRelated, Href: "/other", Title: "Other"},
	}
	linksMu.Unlock()

	routes := []string{"/other"}
	issues := ValidateAgainstRoutes(routes)

	var unregistered []LinkIssue
	for _, i := range issues {
		if i.Kind == "unregistered_route" {
			unregistered = append(unregistered, i)
		}
	}
	require.Len(t, unregistered, 1)
	assert.Equal(t, "/phantom", unregistered[0].Path)
	assert.Contains(t, unregistered[0].Message, "no matching route")
}

func TestValidateAgainstRoutes_MissingRoute(t *testing.T) {
	resetLinks(t)

	Link("/a", "related", "/b", "B")

	routes := []string{"/a", "/b", "/c"}
	issues := ValidateAgainstRoutes(routes)

	var missing []LinkIssue
	for _, i := range issues {
		if i.Kind == "missing_route" {
			missing = append(missing, i)
		}
	}
	require.Len(t, missing, 1)
	assert.Equal(t, "/c", missing[0].Path)
	assert.Contains(t, missing[0].Message, "no link graph presence")
}

func TestValidateAgainstRoutes_EmptyRoutes(t *testing.T) {
	resetLinks(t)

	Link("/a", "related", "/b", "B")

	issues := ValidateAgainstRoutes(nil)
	var unregistered []LinkIssue
	for _, i := range issues {
		if i.Kind == "unregistered_route" {
			unregistered = append(unregistered, i)
		}
	}
	// Both /a and /b are registered as sources
	assert.True(t, len(unregistered) >= 1, "registered paths with no routes should be flagged")
}

func TestValidateAgainstRoutes_EmptyRegistry(t *testing.T) {
	resetLinks(t)

	routes := []string{"/a", "/b"}
	issues := ValidateAgainstRoutes(routes)

	var missing []LinkIssue
	for _, i := range issues {
		if i.Kind == "missing_route" {
			missing = append(missing, i)
		}
	}
	assert.Len(t, missing, 2, "all routes should be flagged as missing from empty registry")
}

func TestValidateAgainstRoutes_TargetOnlyPathUnregistered(t *testing.T) {
	resetLinks(t)

	// /target is only a link target (not a source) and has no matching route.
	Link("/source", "related", "/target", "Target")

	routes := []string{"/source"}
	issues := ValidateAgainstRoutes(routes)

	var unregistered []LinkIssue
	for _, i := range issues {
		if i.Kind == "unregistered_route" && i.Path == "/target" {
			unregistered = append(unregistered, i)
		}
	}
	require.Len(t, unregistered, 1)
	assert.Contains(t, unregistered[0].Message, "no matching route")
}

func TestValidateAgainstRoutes_TargetOnlyPathRegistered(t *testing.T) {
	resetLinks(t)

	// /target is only a link target but IS in the route set — no issue expected.
	Link("/source", "related", "/target", "Target")

	routes := []string{"/source", "/target"}
	issues := ValidateAgainstRoutes(routes)

	var unregistered []LinkIssue
	for _, i := range issues {
		if i.Kind == "unregistered_route" {
			unregistered = append(unregistered, i)
		}
	}
	assert.Empty(t, unregistered, "target-only path in route set should not be flagged")
}

func TestValidateAgainstRoutes_TargetOnlyPathNotFlagged(t *testing.T) {
	resetLinks(t)

	// /target is only a link target, not a source — but it should still
	// count as present in the graph for route matching purposes.
	Link("/source", "up", "/target", "Target")

	routes := []string{"/source", "/target"}
	issues := ValidateAgainstRoutes(routes)

	var missing []LinkIssue
	for _, i := range issues {
		if i.Kind == "missing_route" {
			missing = append(missing, i)
		}
	}
	assert.Empty(t, missing, "target-only paths should count as graph presence")
}
