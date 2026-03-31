package linkwell

import (
	"fmt"
	"strings"
)

// LinkRelation represents a typed relationship between two resources, following
// the IANA link relation model defined in RFC 8288. Each relation carries a
// relation type (Rel), target URL (Href), human-readable label (Title), and an
// optional group name for UI rendering (e.g., the ring or hub name).
type LinkRelation struct {
	// Rel is the IANA link relation type (e.g., "related", "collection", "up").
	// See https://www.iana.org/assignments/link-relations/ for registered values.
	Rel string
	// Href is the target URL of the linked resource.
	Href string
	// Title is a human-readable label for the link, suitable for rendering in
	// navigation bars, context panels, or breadcrumb trails.
	Title string
	// Group is an optional grouping label set by Ring or Hub registration.
	// Templates can use this to visually cluster related links.
	Group string
}

// RelEntry is a path and title pair used as input to Ring and Hub registration.
// Create instances with the Rel helper function.
type RelEntry struct {
	Path  string
	Title string
}

// Rel creates a RelEntry for use with Ring and Hub. This is the preferred
// constructor for building the member lists passed to those functions.
func Rel(path, title string) RelEntry {
	return RelEntry{Path: path, Title: title}
}

// HubEntry represents a hub center with its spoke links, returned by Hubs
// for site map or navigation tree rendering. Spokes are sorted alphabetically
// by title.
type HubEntry struct {
	Path   string
	Title  string
	Spokes []LinkRelation
}

// TitleFromPath derives a human-readable title from the last segment of a URL
// path. Hyphens are replaced with spaces and each word is title-cased.
// Examples: "/demo/inventory" -> "Inventory", "/admin/error-traces" -> "Error Traces".
func TitleFromPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return path
	}
	seg := path[idx+1:]
	seg = strings.ReplaceAll(seg, "-", " ")
	// Title case
	words := strings.Fields(seg)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// LinkHeader formats a slice of LinkRelation values as an RFC 8288 Link header
// string. Each relation is rendered as `<href>; rel="type"; title="label"`,
// joined by commas. Returns an empty string if links is empty.
func LinkHeader(links []LinkRelation) string {
	if len(links) == 0 {
		return ""
	}
	parts := make([]string, len(links))
	for i, l := range links {
		parts[i] = fmt.Sprintf("<%s>; rel=\"%s\"; title=\"%s\"", l.Href, l.Rel, l.Title)
	}
	return strings.Join(parts, ", ")
}
