
package linkwell

import (
	"sort"
	"strings"
	"sync"
)

// Thread safety
//
// All registry operations (Link, Ring, Hub, LinksFor, AllLinks, Hubs,
// LoadStoredLink, RemoveLink) are protected by sync.RWMutex and are safe for
// concurrent use. The typical usage pattern is init-time registration: call
// Link, Ring, and Hub during route setup (before the server starts accepting
// requests), then read with LinksFor, AllLinks, Hubs, etc. at request time.
// Concurrent registration from multiple goroutines is supported but unusual.
//
// ResetForTesting clears the registries and is intended for test
// setup/teardown only. It must not be called concurrently with request
// handlers. In parallel tests, use t.Cleanup(ResetForTesting) within each
// subtest to avoid cross-test pollution.
var (
	linksMu  sync.RWMutex
	linksMap = make(map[string][]LinkRelation)

	hubsMu  sync.RWMutex
	hubsMap = make(map[string]string) // path -> title
)

// Link registers a directional relationship from a source path to a target.
// The rel parameter should be an IANA link relation type (e.g., RelRelated, RelUp,
// "collection"). For rel="related", the inverse link is automatically created
// so the relationship is symmetric — both pages will see each other in their
// link sets. Registration is safe for concurrent use.
func Link(source, rel, target, title string) {
	linksMu.Lock()
	defer linksMu.Unlock()

	linksMap[source] = append(linksMap[source], LinkRelation{
		Rel:   rel,
		Href:  target,
		Title: title,
	})

	// rel="related" is symmetric — auto-create the inverse
	if rel == RelRelated {
		// Derive the inverse title from the source path
		// e.g., "/demo/inventory" -> "Inventory"
		inverseTitle := TitleFromPath(source)
		linksMap[target] = append(linksMap[target], LinkRelation{
			Rel:   RelRelated,
			Href:  source,
			Title: inverseTitle,
		})
	}
}

// LinksFor returns all registered link relations for the given path. If one or
// more rel types are provided, only relations matching those types are returned.
// Returns a copy of the internal slice, safe to modify.
//
// When no links are registered for the exact path, LinksFor walks up the path
// hierarchy by stripping the last segment and retrying until a registered path
// is found or the root is reached. This allows child paths like
// /admin/apps/1 to inherit links from /admin/apps.
func LinksFor(path string, rels ...string) []LinkRelation {
	linksMu.RLock()
	defer linksMu.RUnlock()

	// Walk up the path hierarchy until a registered path is found.
	current := path
	var all []LinkRelation
	for {
		if links, ok := linksMap[current]; ok {
			all = links
			break
		}
		idx := strings.LastIndex(current, "/")
		if idx <= 0 {
			// Reached root with no match.
			break
		}
		current = current[:idx]
	}

	if len(rels) == 0 {
		result := make([]LinkRelation, len(all))
		copy(result, all)
		return result
	}

	relSet := make(map[string]bool, len(rels))
	for _, r := range rels {
		relSet[r] = true
	}

	var filtered []LinkRelation
	for _, l := range all {
		if relSet[l.Rel] {
			filtered = append(filtered, l)
		}
	}
	return filtered
}

// RelatedLinksFor returns only rel="related" links for a path, excluding the
// path itself and deduplicating by href. Useful for rendering context bars or
// "See also" panels where self-links would be redundant.
func RelatedLinksFor(path string) []LinkRelation {
	links := LinksFor(path, RelRelated)
	// Deduplicate by href (symmetric registration can create dupes)
	seen := make(map[string]bool)
	var unique []LinkRelation
	for _, l := range links {
		if l.Href == path || seen[l.Href] {
			continue
		}
		seen[l.Href] = true
		unique = append(unique, l)
	}
	return unique
}

// Ring registers symmetric rel="related" links between all members, creating a
// fully-connected group where every member links to every other member. The name
// parameter is stored as the Group field on each LinkRelation for UI grouping.
// Duplicate links are skipped. Registration is safe for concurrent use.
func Ring(name string, members ...RelEntry) {
	linksMu.Lock()
	defer linksMu.Unlock()

	for i, a := range members {
		for j, b := range members {
			if i == j {
				continue
			}
			if !hasLink(linksMap[a.Path], b.Path, RelRelated) {
				linksMap[a.Path] = append(linksMap[a.Path], LinkRelation{
					Rel:   RelRelated,
					Href:  b.Path,
					Title: b.Title,
					Group: name,
				})
			}
		}
	}
}

// Hub registers a star topology where a center page links to all spokes via
// rel="related", and each spoke links back to the center via rel="up". Spokes
// do not link to each other. The centerTitle is used as the Group field on all
// links and as the hub label returned by Hubs. Registration is safe for
// concurrent use.
func Hub(centerPath, centerTitle string, spokes ...RelEntry) {
	hubsMu.Lock()
	hubsMap[centerPath] = centerTitle
	hubsMu.Unlock()

	linksMu.Lock()
	defer linksMu.Unlock()

	for _, spoke := range spokes {
		// Center -> spoke
		if !hasLink(linksMap[centerPath], spoke.Path, RelRelated) {
			linksMap[centerPath] = append(linksMap[centerPath], LinkRelation{
				Rel:   RelRelated,
				Href:  spoke.Path,
				Title: spoke.Title,
				Group: centerTitle,
			})
		}
		// Spoke -> center (uses rel="up" to indicate parent)
		if !hasLink(linksMap[spoke.Path], centerPath, RelUp) {
			linksMap[spoke.Path] = append(linksMap[spoke.Path], LinkRelation{
				Rel:   RelUp,
				Href:  centerPath,
				Title: centerTitle,
				Group: centerTitle,
			})
		}
	}
}

// linksForExact returns all registered link relations for exactly the given
// path, with optional rel filtering. It does NOT walk up the path hierarchy.
func linksForExact(path string, rels ...string) []LinkRelation {
	linksMu.RLock()
	defer linksMu.RUnlock()

	all := linksMap[path]
	if len(rels) == 0 {
		result := make([]LinkRelation, len(all))
		copy(result, all)
		return result
	}
	relSet := make(map[string]bool, len(rels))
	for _, r := range rels {
		relSet[r] = true
	}
	var filtered []LinkRelation
	for _, l := range all {
		if relSet[l.Rel] {
			filtered = append(filtered, l)
		}
	}
	return filtered
}

// registeredTitleFor returns the title registered for targetPath by checking
// its parent's links first (via rel="up"), then scanning all registry entries.
// Returns "" if no registered title is found.
func registeredTitleFor(targetPath string) string {
	linksMu.RLock()
	defer linksMu.RUnlock()

	// Check the immediate parent (via rel="up") for a link targeting this path.
	for _, l := range linksMap[targetPath] {
		if l.Rel == RelUp {
			for _, pl := range linksMap[l.Href] {
				if pl.Href == targetPath && pl.Title != "" {
					return pl.Title
				}
			}
		}
	}

	// Fallback: scan all entries for any link targeting this path with a title.
	for _, links := range linksMap {
		for _, l := range links {
			if l.Href == targetPath && l.Title != "" {
				return l.Title
			}
		}
	}

	return ""
}

// hasLink checks if a link with the given href and rel already exists.
func hasLink(links []LinkRelation, href, rel string) bool {
	for _, l := range links {
		if l.Href == href && l.Rel == rel {
			return true
		}
	}
	return false
}

// AllLinks returns a snapshot of all registered link relations grouped by
// source path. The returned map and slices are copies, safe to modify. Useful
// for admin dashboards or debug pages that inspect the link graph.
func AllLinks() map[string][]LinkRelation {
	linksMu.RLock()
	defer linksMu.RUnlock()

	result := make(map[string][]LinkRelation, len(linksMap))
	for k, v := range linksMap {
		copied := make([]LinkRelation, len(v))
		copy(copied, v)
		result[k] = copied
	}
	return result
}

// sortedPaths returns the keys of a link map in alphabetical order.
func sortedPaths(links map[string][]LinkRelation) []string {
	paths := make([]string, 0, len(links))
	for k := range links {
		paths = append(paths, k)
	}
	sort.Strings(paths)
	return paths
}

// Hubs returns all registered hub centers with their spoke links, sorted by
// center path. Spokes within each hub are sorted alphabetically by title. Use
// for rendering site maps or navigation trees.
func Hubs() []HubEntry {
	hubsMu.RLock()
	paths := make([]string, 0, len(hubsMap))
	titles := make(map[string]string, len(hubsMap))
	for p, t := range hubsMap {
		paths = append(paths, p)
		titles[p] = t
	}
	hubsMu.RUnlock()

	sort.Strings(paths)

	entries := make([]HubEntry, 0, len(paths))
	for _, p := range paths {
		spokes := linksForExact(p, RelRelated)
		sort.Slice(spokes, func(i, j int) bool {
			return spokes[i].Title < spokes[j].Title
		})
		entries = append(entries, HubEntry{
			Path:   p,
			Title:  titles[p],
			Spokes: spokes,
		})
	}
	return entries
}

// BreadcrumbsFromLinks walks the rel="up" chain from path to build a
// breadcrumb trail. The trail starts with Home, includes each ancestor found
// via rel="up" links, and ends with the current page (empty Href). Returns nil
// if no rel="up" links are registered for the path.
//
// When the exact path has no rel="up" links, BreadcrumbsFromLinks strips the
// last path segment and retries up the hierarchy until a registered path is
// found. Intermediate segments between the matched ancestor and the original
// path are included as breadcrumbs. This allows child pages like
// /admin/groups/new to inherit breadcrumbs from /admin/groups. Cycle-safe.
func BreadcrumbsFromLinks(path string) []Breadcrumb {
	// Find the nearest ancestor with rel="up" links by walking up the path.
	// Use linksForExact so the walk-up is controlled here, not inside LinksFor.
	matchedPath := path
	var tailSegments []string
	for {
		upLinks := linksForExact(matchedPath, RelUp)
		if len(upLinks) > 0 {
			break
		}
		// Strip the last segment and try the parent path.
		idx := strings.LastIndex(matchedPath, "/")
		if idx <= 0 {
			// Reached root with no match.
			return nil
		}
		tailSegments = append([]string{matchedPath[idx+1:]}, tailSegments...)
		matchedPath = matchedPath[:idx]
	}

	// Walk the rel="up" chain from the matched path.
	var crumbs []Breadcrumb
	visited := map[string]bool{}
	current := matchedPath

	for !visited[current] {
		visited[current] = true

		upLinks := linksForExact(current, RelUp)
		if len(upLinks) == 0 {
			break
		}

		parent := upLinks[0]
		crumbs = append([]Breadcrumb{{Label: parent.Title, Href: parent.Href}}, crumbs...)
		current = parent.Href
	}

	// Add home at the start, unless the walked chain already includes it.
	if len(crumbs) == 0 || crumbs[0].Href != "/" {
		crumbs = append([]Breadcrumb{{Label: BreadcrumbLabelHome, Href: "/"}}, crumbs...)
	}

	// Add the matched path itself as a breadcrumb (it has rel="up" links,
	// so it is a known page). Prefer the registered title from the parent's
	// link targeting this path (e.g., Hub spoke titles) over TitleFromPath.
	label := registeredTitleFor(matchedPath)
	if label == "" {
		label = TitleFromPath(matchedPath)
	}
	crumbs = append(crumbs, Breadcrumb{
		Label: label,
		Href:  matchedPath,
	})

	// Add intermediate tail segments (derived from path walk-up).
	built := matchedPath
	for i, seg := range tailSegments {
		built += "/" + seg
		href := built
		if i == len(tailSegments)-1 {
			href = "" // terminal segment has no href
		}
		crumbs = append(crumbs, Breadcrumb{Label: TitleFromPath("/" + seg), Href: href})
	}

	// If no tail segments were added, the matched path IS the original path,
	// so set the last crumb's Href to empty (terminal).
	if len(tailSegments) == 0 {
		crumbs[len(crumbs)-1].Href = ""
	}

	return crumbs
}

// ResolveFromMaskWithPath combines bitmask-resolved breadcrumbs with
// path-derived crumbs (using the walk-up behavior of BreadcrumbsFromLinks),
// deduplicating by base path (query parameters stripped for comparison). The
// from parameter is forwarded as a ?from= query parameter on intermediate path
// crumb links. Returns nil if path produces no breadcrumbs.
func ResolveFromMaskWithPath(mask uint64, path string, from string) []Breadcrumb {
	maskCrumbs := ResolveFromMask(mask)
	pathCrumbs := BreadcrumbsFromLinks(path)
	if pathCrumbs == nil {
		// Fall back to path-based breadcrumbs.
		pathCrumbs = BreadcrumbsFromPath(path, nil)
	}

	// Build a set of base paths from mask crumbs for deduplication.
	seen := make(map[string]bool, len(maskCrumbs))
	for _, c := range maskCrumbs {
		seen[basePath(c.Href)] = true
	}

	// Start with mask crumbs, then append path crumbs that aren't duplicates.
	merged := make([]Breadcrumb, len(maskCrumbs))
	copy(merged, maskCrumbs)

	for _, c := range pathCrumbs {
		bp := basePath(c.Href)
		if seen[bp] {
			continue
		}
		seen[bp] = true
		// Forward the from param on intermediate links.
		if c.Href != "" && from != "" {
			c.Href = FromNav(c.Href, from)
		}
		merged = append(merged, c)
	}

	return merged
}

// basePath strips query parameters from a URL path for deduplication.
func basePath(href string) string {
	if idx := strings.IndexByte(href, '?'); idx >= 0 {
		return href[:idx]
	}
	return href
}

// LoadStoredLink adds a single link relation from an external source (e.g., a
// database or configuration file) into the in-memory registry. Duplicate links
// (same source, href, and rel) are silently skipped.
func LoadStoredLink(source string, r LinkRelation) {
	linksMu.Lock()
	defer linksMu.Unlock()
	if !hasLink(linksMap[source], r.Href, r.Rel) {
		linksMap[source] = append(linksMap[source], r)
	}
}

// RemoveLink removes the first link relation matching source, href, and rel
// from the in-memory registry. Returns true if a link was found and removed,
// false if no match was found.
func RemoveLink(source, href, rel string) bool {
	linksMu.Lock()
	defer linksMu.Unlock()
	links := linksMap[source]
	for i, l := range links {
		if l.Href == href && l.Rel == rel {
			linksMap[source] = append(links[:i], links[i+1:]...)
			if len(linksMap[source]) == 0 {
				delete(linksMap, source)
			}
			return true
		}
	}
	return false
}

// ResetForTesting clears all entries from the global link, hub, and
// breadcrumb-origin registries. The Home breadcrumb (bit 0) is re-registered
// automatically. Intended for use in test setup/teardown only — not safe to
// call in production while handlers may be reading the registry. In parallel
// tests, call ResetForTesting at the start of each subtest and register it
// with t.Cleanup to avoid cross-test pollution.
func ResetForTesting() {
	linksMu.Lock()
	linksMap = make(map[string][]LinkRelation)
	linksMu.Unlock()
	hubsMu.Lock()
	hubsMap = make(map[string]string)
	hubsMu.Unlock()
	fromMu.Lock()
	fromEntries = []fromEntry{{bit: FromHome, crumb: Breadcrumb{Label: BreadcrumbLabelHome, Href: "/"}}}
	fromMu.Unlock()
}
