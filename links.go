
package linkwell

import (
	"sort"
	"sync"
)

// linkRegistry stores registered link relations keyed by source path.
var (
	linksMu  sync.RWMutex
	linksMap = make(map[string][]LinkRelation)

	hubsMu  sync.RWMutex
	hubsMap = make(map[string]string) // path -> title
)

// Link registers a relationship from a source path to a target.
// For rel="related", the inverse is automatically registered (symmetric).
func Link(source, rel, target, title string) {
	linksMu.Lock()
	defer linksMu.Unlock()

	linksMap[source] = append(linksMap[source], LinkRelation{
		Rel:   rel,
		Href:  target,
		Title: title,
	})

	// rel="related" is symmetric — auto-create the inverse
	if rel == "related" {
		// Derive the inverse title from the source path
		// e.g., "/demo/inventory" -> "Inventory"
		inverseTitle := TitleFromPath(source)
		linksMap[target] = append(linksMap[target], LinkRelation{
			Rel:   "related",
			Href:  source,
			Title: inverseTitle,
		})
	}
}

// LinksFor returns all registered link relations for a path.
// If rels is provided, only relations matching those types are returned.
func LinksFor(path string, rels ...string) []LinkRelation {
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

// RelatedLinksFor returns only rel="related" links for a path,
// excluding the current path itself (for use in context bars).
func RelatedLinksFor(path string) []LinkRelation {
	links := LinksFor(path, "related")
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

// Ring registers symmetric rel="related" links between all members.
// Every member links to every other member. The name is used for UI grouping.
func Ring(name string, members ...RelEntry) {
	linksMu.Lock()
	defer linksMu.Unlock()

	for i, a := range members {
		for j, b := range members {
			if i == j {
				continue
			}
			if !hasLink(linksMap[a.Path], b.Path, "related") {
				linksMap[a.Path] = append(linksMap[a.Path], LinkRelation{
					Rel:   "related",
					Href:  b.Path,
					Title: b.Title,
					Group: name,
				})
			}
		}
	}
}

// Hub registers a center page that links to all spokes, and each spoke
// links back to the center only. Spokes do not link to each other.
// The centerTitle is used as the group name for all links.
func Hub(centerPath, centerTitle string, spokes ...RelEntry) {
	hubsMu.Lock()
	hubsMap[centerPath] = centerTitle
	hubsMu.Unlock()

	linksMu.Lock()
	defer linksMu.Unlock()

	for _, spoke := range spokes {
		// Center -> spoke
		if !hasLink(linksMap[centerPath], spoke.Path, "related") {
			linksMap[centerPath] = append(linksMap[centerPath], LinkRelation{
				Rel:   "related",
				Href:  spoke.Path,
				Title: spoke.Title,
				Group: centerTitle,
			})
		}
		// Spoke -> center (uses rel="up" to indicate parent)
		if !hasLink(linksMap[spoke.Path], centerPath, "up") {
			linksMap[spoke.Path] = append(linksMap[spoke.Path], LinkRelation{
				Rel:   "up",
				Href:  centerPath,
				Title: centerTitle,
				Group: centerTitle,
			})
		}
	}
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

// AllLinks returns all registered link relations grouped by source path.
// Used for admin/debug inspection.
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

// SortedPaths returns all registered source paths in sorted order.
func SortedPaths(links map[string][]LinkRelation) []string {
	paths := make([]string, 0, len(links))
	for k := range links {
		paths = append(paths, k)
	}
	sort.Strings(paths)
	return paths
}

// Hubs returns all registered hub centers with their spokes, sorted by path.
// Spokes within each hub are sorted alphabetically by title.
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
		spokes := LinksFor(p, "related")
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

// BreadcrumbsFromLinks walks the rel="up" chain from path to build a breadcrumb trail.
// Returns nil if no rel="up" links are registered for the path.
func BreadcrumbsFromLinks(path string) []Breadcrumb {
	var crumbs []Breadcrumb
	visited := map[string]bool{}
	current := path

	for {
		if visited[current] {
			break // prevent cycles
		}
		visited[current] = true

		upLinks := LinksFor(current, "up")
		if len(upLinks) == 0 {
			break
		}

		parent := upLinks[0]
		crumbs = append([]Breadcrumb{{Label: parent.Title, Href: parent.Href}}, crumbs...)
		current = parent.Href
	}

	// Add home at the start
	if len(crumbs) > 0 {
		crumbs = append([]Breadcrumb{{Label: BreadcrumbLabelHome, Href: "/"}}, crumbs...)
	}

	// Add current page (no href = current page, not a link)
	if len(crumbs) > 0 {
		crumbs = append(crumbs, Breadcrumb{Label: TitleFromPath(path)})
	}

	return crumbs
}

// LoadStoredLink adds a single link relation from an external source (e.g., database)
// into the in-memory registry. Skips duplicates.
func LoadStoredLink(source string, r LinkRelation) {
	linksMu.Lock()
	defer linksMu.Unlock()
	if !hasLink(linksMap[source], r.Href, r.Rel) {
		linksMap[source] = append(linksMap[source], r)
	}
}

// RemoveLink removes a link relation matching source, href, and rel from the
// in-memory registry. Returns true if a link was removed.
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

// ResetForTesting clears the global link registry. Test use only.
func ResetForTesting() {
	linksMu.Lock()
	linksMap = make(map[string][]LinkRelation)
	linksMu.Unlock()
	hubsMu.Lock()
	hubsMap = make(map[string]string)
	hubsMu.Unlock()
}
