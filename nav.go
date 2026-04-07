package linkwell

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Renderable is a minimal rendering interface satisfied by templ.Component and
// any type that can render itself to an io.Writer. Using this interface instead
// of templ.Component directly allows consumers that don't use templ to avoid
// pulling in the templ module.
type Renderable interface {
	Render(ctx context.Context, w io.Writer) error
}

// NavConfig holds the app-controlled parts of the navigation layout. Zero
// values are safe defaults: no promoted item, all items visible, no custom
// brand or topbar slots.
type NavConfig struct {
	// Items is the ordered list of navigation entries.
	Items []NavItem
	// Promoted is an optional floating action button item shown on mobile.
	Promoted *NavItem
	// MaxVisible limits how many items are shown before overflowing into a
	// "more" menu. Zero means show all items.
	MaxVisible int
	// AppName is plain text displayed in the brand area of the navigation bar.
	AppName string
	// Brand is an optional component that replaces the default AppName text in
	// the brand slot. Any templ.Component satisfies Renderable. Set to nil to
	// use the plain text AppName.
	Brand Renderable
	// Topbar is an optional component that replaces the default mobile topbar
	// content. Any templ.Component satisfies Renderable. Set to nil to use the
	// default layout.
	Topbar Renderable
}

// NavItem is a server-computed navigation entry. Active state is determined by
// the handler (via SetActiveNavItem or SetActiveNavItemPrefix), not by
// client-side JavaScript. Items may have children for dropdown/flyout menus.
type NavItem struct {
	// HTMXAttrs holds optional HTMX attributes (e.g., from HxRequestConfig.Attrs())
	// for items that trigger HTMX requests instead of full navigation.
	HTMXAttrs map[string]string
	// Label is the user-visible text for this navigation entry.
	Label string
	// Href is the navigation target URL.
	Href string
	// Icon is the icon name rendered alongside the label.
	Icon string
	// Children are nested sub-items for dropdown menus.
	Children []NavItem
	// Active indicates this item matches the current page. Set by SetActiveNavItem
	// or SetActiveNavItemPrefix, or manually by the handler.
	Active bool
}

// Breadcrumb is one segment of a breadcrumb trail. When Href is empty, the
// segment represents the current page and should be rendered as plain text
// rather than a clickable link.
type Breadcrumb struct {
	// Label is the display text for this breadcrumb segment.
	Label string
	// Href is the navigation target. Empty for the terminal (current page) segment.
	Href string
}

// BreadcrumbLabelHome is the default label for the root breadcrumb segment.
const BreadcrumbLabelHome = "Home"

// SetActiveNavItem sets the Active flag on nav items using exact path matching.
// A parent item is also marked active if any of its children match. Returns a
// new slice; the input is not modified.
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

// SetActiveNavItemPrefix sets the Active flag using longest-prefix matching.
// An item is active if the current path equals its Href or starts with Href
// followed by "/". Use for section-level navigation where /users should be
// active when the path is /users/42/edit. A parent is marked active if any
// child matches. Returns a new slice; the input is not modified.
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

// NavItemFromControl converts a Control into a NavItem, mapping Label, Href,
// Icon, and HxRequest attributes. Useful when a control needs to appear in the
// navigation bar.
func NavItemFromControl(ctrl Control) NavItem {
	return NavItem{
		Label:     ctrl.Label,
		Href:      ctrl.Href,
		Icon:      string(ctrl.Icon),
		HTMXAttrs: ctrl.HxRequest.Attrs(),
	}
}

// BreadcrumbsFromPath generates a breadcrumb trail from a URL path by splitting
// on "/" and title-casing each segment. The labels map overrides auto-generated
// labels by segment index (0-based, not counting the Home crumb) — use this to
// replace opaque IDs with human-readable names. The terminal segment always has
// an empty Href (rendered as text, not a link).
func BreadcrumbsFromPath(path string, labels map[int]string) []Breadcrumb {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return []Breadcrumb{{Label: BreadcrumbLabelHome, Href: "/"}}
	}
	segments := strings.Split(trimmed, "/")
	crumbs := make([]Breadcrumb, 0, len(segments)+1)
	crumbs = append(crumbs, Breadcrumb{Label: BreadcrumbLabelHome, Href: "/"})
	for i, seg := range segments {
		label := TitleFromPath("/" + seg)
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

// FromBit is a bitmask position for a registered breadcrumb origin. Lower bits
// render earlier in the trail. Bit 0 is reserved for Home and is always
// included. Use RegisterFrom to associate a Breadcrumb with a bit position,
// then pass a bitmask via the ?from= query parameter to reconstruct the trail.
type FromBit = uint64

// Well-known breadcrumb bit positions. Bit 0 (Home) is always included.
// Register additional bits via RegisterFrom. Bits 2-7 are available for
// application-specific origins.
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

// Protected by fromMu. RegisterFrom and ResolveFromMask are safe for
// concurrent use. Registration is expected at init time; reads happen at
// request time.
var (
	fromMu      sync.RWMutex
	fromEntries []fromEntry
)

func init() {
	// Home is always registered at bit 0.
	RegisterFrom(FromHome, Breadcrumb{Label: BreadcrumbLabelHome, Href: "/"})
}

// RegisterFrom registers a breadcrumb at the given bit position. Call during
// route initialization. Bit 0 is pre-registered as Home. If the bit is already
// registered, the previous entry is replaced.
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

// ResolveFromMask decodes a bitmask into an ordered breadcrumb trail by
// checking each registered bit position. Only registered bits are included;
// unregistered bits are silently ignored. Home (bit 0) is always included
// regardless of the mask value. Entries are ordered by bit position (lowest
// first).
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

// ParseFromParam parses the ?from= query parameter string as a uint64 bitmask.
// Returns 0 if the input is empty or not a valid unsigned integer.
func ParseFromParam(raw string) uint64 {
	if raw == "" {
		return 0
	}
	v, _ := strconv.ParseUint(raw, 10, 64)
	return v
}

// FromParam formats a bitmask as a decimal string suitable for ?from= query
// parameter values.
func FromParam(mask uint64) string {
	return strconv.FormatUint(mask, 10)
}

// FromQueryString returns "from=N" formatted for inclusion in URL query
// strings. Returns an empty string if mask is 0 (no origin context).
func FromQueryString(mask uint64) string {
	if mask == 0 {
		return ""
	}
	return fmt.Sprintf("from=%s", FromParam(mask))
}

// FromNav appends the ?from= parameter to a href, preserving any existing
// query string. Returns href unchanged if from is empty. Use in templates to
// forward breadcrumb context to outbound links.
func FromNav(href, from string) string {
	if from == "" {
		return href
	}
	if strings.Contains(href, "?") {
		return href + "&from=" + from
	}
	return href + "?from=" + from
}
