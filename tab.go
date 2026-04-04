package linkwell

// TabItem is a server-computed tab entry for in-page tabbed navigation. Active
// state is determined by the handler (via SetActiveTab), not by client-side
// JavaScript. Each tab typically maps to an HTMX lazy-loaded content panel.
type TabItem struct {
	// Label is the user-visible text for this tab.
	Label string
	// Href is the content endpoint loaded when the tab is selected.
	Href string
	// Target is the hx-target CSS selector for this tab's content panel.
	// Overrides TabConfig.Target when set.
	Target string
	// Icon is the icon name rendered alongside the label.
	Icon Icon
	// Active indicates this tab matches the current page. Set by SetActiveTab
	// or manually by the handler.
	Active bool
	// Disabled renders the tab in a non-interactive state.
	Disabled bool
	// Badge is an optional count or status indicator displayed on the tab.
	Badge string
	// Swap is the HTMX swap strategy for this tab's content. Defaults to
	// innerHTML when empty.
	Swap SwapMode
}

// TabConfig holds the configuration for an in-page tabbed navigation component.
// The server decides which tabs exist and which is active — templates consume
// this to render the tab bar and content panel.
type TabConfig struct {
	// ID is a unique identifier for this tab group (e.g., "user-tabs").
	ID string
	// Items is the ordered list of tab entries.
	Items []TabItem
	// Target is the shared default hx-target CSS selector for all tabs.
	// Individual TabItem.Target values override this.
	Target string
}

// NewTabConfig creates a TabConfig with the given ID, shared target selector,
// and tab items. Items with an empty Target inherit the shared target at
// render time.
func NewTabConfig(id, target string, items ...TabItem) TabConfig {
	return TabConfig{
		ID:     id,
		Items:  items,
		Target: target,
	}
}

// SetActiveTab sets the Active flag on tab items using exact path matching.
// Returns a new slice; the input is not modified.
func SetActiveTab(tabs []TabItem, currentPath string) []TabItem {
	result := make([]TabItem, len(tabs))
	for i, tab := range tabs {
		tab.Active = tab.Href == currentPath
		result[i] = tab
	}
	return result
}
