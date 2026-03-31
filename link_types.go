package linkwell

import (
	"fmt"
	"strings"
)

// LinkRelation represents a relationship between two resources.
type LinkRelation struct {
	Rel   string // IANA link relation (e.g., "related", "collection", "up")
	Href  string // Target URL
	Title string // Human-readable label
	Group string // Optional group name (e.g., ring name) for UI grouping
}

// RelEntry is a path+title pair for use with Ring and Hub.
type RelEntry struct {
	Path  string
	Title string
}

// Rel creates a RelEntry for use with Ring and Hub.
func Rel(path, title string) RelEntry {
	return RelEntry{Path: path, Title: title}
}

// HubEntry represents a hub center with its spoke links for site map rendering.
type HubEntry struct {
	Path   string
	Title  string
	Spokes []LinkRelation
}

// TitleFromPath extracts a title from the last segment of a URL path.
// "/demo/inventory" -> "Inventory", "/admin/error-traces" -> "Error Traces"
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

// LinkHeader formats link relations as an RFC 8288 Link header value.
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
