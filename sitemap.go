package linkwell

import "sort"

// SitemapEntry represents a single page in the site hierarchy, derived from the
// link registry. Each entry captures the page's path, title, parent (via
// rel="up"), children (hub spokes), and optional ring group membership.
type SitemapEntry struct {
	// Path is the URL path of the page.
	Path string
	// Title is a human-readable label for the page.
	Title string
	// Parent is the rel="up" parent path, empty for root entries.
	Parent string
	// Children holds spoke paths when this entry is a hub center.
	Children []string
	// Group is the ring group name, if any.
	Group string
}

// Sitemap derives a structured sitemap from the link registry. It collects all
// registered paths, resolves parent relationships from rel="up" links, populates
// children from hub registrations, and captures ring group membership. Entries
// are sorted alphabetically by path.
func Sitemap() []SitemapEntry {
	all := AllLinks()
	hubs := Hubs()

	// Build a set of hub centers for quick lookup, and map center -> spoke paths.
	hubChildren := make(map[string][]string, len(hubs))
	for _, h := range hubs {
		children := make([]string, 0, len(h.Spokes))
		for _, s := range h.Spokes {
			children = append(children, s.Href)
		}
		hubChildren[h.Path] = children
	}

	// Build hub title lookup.
	hubTitles := make(map[string]string, len(hubs))
	for _, h := range hubs {
		hubTitles[h.Path] = h.Title
	}

	// Collect all unique paths.
	pathSet := make(map[string]bool, len(all))
	for p := range all {
		pathSet[p] = true
	}

	paths := make([]string, 0, len(pathSet))
	for p := range pathSet {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	entries := make([]SitemapEntry, 0, len(paths))
	for _, p := range paths {
		links := all[p]

		entry := SitemapEntry{
			Path: p,
		}

		// Resolve title: prefer hub title, then derive from path.
		if t, ok := hubTitles[p]; ok {
			entry.Title = t
		} else {
			entry.Title = TitleFromPath(p)
		}

		// Find parent from rel="up" links.
		for _, l := range links {
			if l.Rel == RelUp {
				entry.Parent = l.Href
				break
			}
		}

		// Populate children from hub spokes.
		if children, ok := hubChildren[p]; ok {
			entry.Children = children
		}

		// Capture group from any link that has one set.
		for _, l := range links {
			if l.Group != "" {
				entry.Group = l.Group
				break
			}
		}

		entries = append(entries, entry)
	}

	return entries
}

// SitemapRoots returns only sitemap entries that have no parent (empty Parent
// field). These are the top-level pages in the site hierarchy.
func SitemapRoots() []SitemapEntry {
	all := Sitemap()
	var roots []SitemapEntry
	for _, e := range all {
		if e.Parent == "" {
			roots = append(roots, e)
		}
	}
	return roots
}
