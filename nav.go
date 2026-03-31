package linkwell

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/a-h/templ"
)

// NavConfig holds the app-controlled parts of the nav layout.
// Zero values are safe: no promoted item, all items visible, no custom slots.
type NavConfig struct {
	Items      []NavItem       // Navigation items
	Promoted   *NavItem        // Optional promoted FAB item (mobile)
	MaxVisible int             // Items visible before overflow (0 = show all)
	AppName    string          // Brand text
	Brand      templ.Component // Custom brand slot (replaces default appName text)
	Topbar     templ.Component // Custom mobile topbar content (replaces default)
}

// NavItem is a server-computed navigation affordance.
// Active state is set by the handler (or SetActiveNavItem), not by JavaScript.
type NavItem struct {
	HTMXAttrs map[string]string
	Label     string
	Href      string
	Icon      string
	Children  []NavItem
	Active    bool
}

// Breadcrumb is one segment of a breadcrumb trail.
// Href empty means this is the current page (rendered as text, not an anchor).
type Breadcrumb struct {
	Label string
	Href  string
}

// BreadcrumbLabelHome is the default label for the root breadcrumb segment.
const BreadcrumbLabelHome = "Home"

// SetActiveNavItem performs exact-match active state setting.
// A parent item is marked active if any of its children match.
func SetActiveNavItem(items []NavItem, currentPath string) []NavItem {
	result := make([]NavItem, len(items))
	for i, item := range items {
		item.Children = SetActiveNavItem(item.Children, currentPath)
		childActive := false
		for _, child := range item.Children {
			if child.Active {
				childActive = true
				break
			}
		}
		item.Active = item.Href == currentPath || childActive
		result[i] = item
	}
	return result
}

// SetActiveNavItemPrefix performs longest-prefix match active state setting.
// Use for section-level nav: /users is active when path is /users/42/edit.
// A parent item is marked active if any of its children match.
func SetActiveNavItemPrefix(items []NavItem, currentPath string) []NavItem {
	result := make([]NavItem, len(items))
	for i, item := range items {
		item.Children = SetActiveNavItemPrefix(item.Children, currentPath)
		childActive := false
		for _, child := range item.Children {
			if child.Active {
				childActive = true
				break
			}
		}
		// Require href+separator to avoid "/" matching every path.
		isActive := item.Href != "" &&
			(currentPath == item.Href || strings.HasPrefix(currentPath, item.Href+"/"))
		item.Active = isActive || childActive
		result[i] = item
	}
	return result
}

// NavItemFromControl bridges a Control to a NavItem (Label, Href, Icon, HTMXAttrs).
func NavItemFromControl(ctrl Control) NavItem {
	return NavItem{
		Label:     ctrl.Label,
		Href:      ctrl.Href,
		Icon:      string(ctrl.Icon),
		HTMXAttrs: ctrl.HxRequest.Attrs(),
	}
}

// BreadcrumbsFromPath generates crumbs from a URL path.
// labels overrides auto-generated labels by segment index (0-based, not counting the Home crumb).
// The terminal segment always has an empty Href (rendered as plain text).
func BreadcrumbsFromPath(path string, labels map[int]string) []Breadcrumb {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return []Breadcrumb{{Label: BreadcrumbLabelHome, Href: "/"}}
	}
	segments := strings.Split(trimmed, "/")
	crumbs := make([]Breadcrumb, 0, len(segments)+1)
	crumbs = append(crumbs, Breadcrumb{Label: BreadcrumbLabelHome, Href: "/"})
	for i, seg := range segments {
		label := seg
		if l, ok := labels[i]; ok {
			label = l
		}
		href := "/" + strings.Join(segments[:i+1], "/")
		if i == len(segments)-1 {
			href = "" // terminal segment has no href
		}
		crumbs = append(crumbs, Breadcrumb{Label: label, Href: href})
	}
	return crumbs
}

// FromBit is a bitmask position for a registered breadcrumb origin.
// Lower bits render earlier in the trail. Bit 0 is reserved for Home.
type FromBit = uint64

// Well-known breadcrumb bit positions. Bit 0 (Home) is always included.
// Register additional bits via RegisterFrom.
const (
	FromHome      FromBit = 1 << iota // bit 0 — always shown
	FromDashboard                     // bit 1
	FromBit2                          // bit 2 — available for registration
	FromBit3                          // bit 3
	FromBit4                          // bit 4
	FromBit5                          // bit 5
	FromBit6                          // bit 6
	FromBit7                          // bit 7
)

// fromEntry is a registered breadcrumb with its bit position.
type fromEntry struct {
	bit   FromBit
	crumb Breadcrumb
}

var (
	fromMu      sync.RWMutex
	fromEntries []fromEntry
)

func init() {
	// Home is always registered at bit 0.
	RegisterFrom(FromHome, Breadcrumb{Label: BreadcrumbLabelHome, Href: "/"})
}

// RegisterFrom registers a breadcrumb at the given bit position.
// Call during route initialization. Bit 0 is pre-registered as Home.
func RegisterFrom(bit FromBit, crumb Breadcrumb) {
	fromMu.Lock()
	// Replace if bit already registered.
	for i, e := range fromEntries {
		if e.bit == bit {
			fromEntries[i].crumb = crumb
			fromMu.Unlock()
			return
		}
	}
	fromEntries = append(fromEntries, fromEntry{bit: bit, crumb: crumb})
	sort.Slice(fromEntries, func(i, j int) bool {
		return fromEntries[i].bit < fromEntries[j].bit
	})
	fromMu.Unlock()
}

// ResolveFromMask decodes a bitmask into an ordered breadcrumb trail.
// Only registered bits are included. Unregistered bits are silently ignored.
// Home (bit 0) is always included regardless of the mask value.
func ResolveFromMask(mask uint64) []Breadcrumb {
	mask |= FromHome // always include Home
	fromMu.RLock()
	defer fromMu.RUnlock()

	var crumbs []Breadcrumb
	for _, e := range fromEntries {
		if mask&e.bit != 0 {
			crumbs = append(crumbs, e.crumb)
		}
	}
	return crumbs
}

// ParseFromParam parses the ?from= query parameter as a uint64 bitmask.
// Returns 0 if empty or invalid.
func ParseFromParam(raw string) uint64 {
	if raw == "" {
		return 0
	}
	v, _ := strconv.ParseUint(raw, 10, 64)
	return v
}

// FromParam formats a bitmask as a string suitable for ?from= query values.
func FromParam(mask uint64) string {
	return strconv.FormatUint(mask, 10)
}

// FromQueryString returns "from=N" for use in URL query strings.
// Returns empty string if mask is 0.
func FromQueryString(mask uint64) string {
	if mask == 0 {
		return ""
	}
	return fmt.Sprintf("from=%s", FromParam(mask))
}

// FromNav appends the from parameter to a href if non-empty.
// Use in templates to forward breadcrumb context to outbound links.
func FromNav(href, from string) string {
	if from == "" {
		return href
	}
	if strings.Contains(href, "?") {
		return href + "&from=" + from
	}
	return href + "?from=" + from
}
