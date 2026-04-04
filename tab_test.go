package linkwell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewTabConfig
// ---------------------------------------------------------------------------

func TestNewTabConfig_SetsFields(t *testing.T) {
	tabs := NewTabConfig("user-tabs", "#tab-content",
		TabItem{Label: "Overview", Href: "/users/42/overview"},
		TabItem{Label: "Activity", Href: "/users/42/activity"},
	)
	require.Equal(t, "user-tabs", tabs.ID)
	require.Equal(t, "#tab-content", tabs.Target)
	require.Len(t, tabs.Items, 2)
	require.Equal(t, "Overview", tabs.Items[0].Label)
	require.Equal(t, "Activity", tabs.Items[1].Label)
}

func TestNewTabConfig_NoItems(t *testing.T) {
	tabs := NewTabConfig("empty", "#panel")
	require.Equal(t, "empty", tabs.ID)
	require.Equal(t, "#panel", tabs.Target)
	require.Empty(t, tabs.Items)
}

// ---------------------------------------------------------------------------
// SetActiveTab
// ---------------------------------------------------------------------------

func TestSetActiveTab_ExactMatch(t *testing.T) {
	tabs := []TabItem{
		{Label: "Overview", Href: "/users/42/overview"},
		{Label: "Activity", Href: "/users/42/activity"},
		{Label: "Settings", Href: "/users/42/settings"},
	}
	result := SetActiveTab(tabs, "/users/42/activity")
	require.False(t, result[0].Active, "Overview should not be active")
	require.True(t, result[1].Active, "Activity should be active")
	require.False(t, result[2].Active, "Settings should not be active")
}

func TestSetActiveTab_NoMatch(t *testing.T) {
	tabs := []TabItem{
		{Label: "Overview", Href: "/users/42/overview"},
		{Label: "Activity", Href: "/users/42/activity"},
	}
	result := SetActiveTab(tabs, "/other")
	for _, tab := range result {
		require.False(t, tab.Active, "%s should not be active", tab.Label)
	}
}

func TestSetActiveTab_DoesNotMutateOriginal(t *testing.T) {
	tabs := []TabItem{
		{Label: "Overview", Href: "/users/42/overview"},
	}
	result := SetActiveTab(tabs, "/users/42/overview")
	require.True(t, result[0].Active)
	require.False(t, tabs[0].Active, "original slice must not be mutated")
}

func TestSetActiveTab_EmptySlice(t *testing.T) {
	result := SetActiveTab(nil, "/anything")
	require.Empty(t, result)
}

func TestSetActiveTab_PreservesOtherFields(t *testing.T) {
	tabs := []TabItem{
		{
			Label:    "Overview",
			Href:     "/users/42/overview",
			Target:   "#custom-panel",
			Icon:     IconHome,
			Disabled: true,
			Badge:    "3",
			Swap:     SwapOuterHTML,
		},
	}
	result := SetActiveTab(tabs, "/users/42/overview")
	require.True(t, result[0].Active)
	require.Equal(t, "#custom-panel", result[0].Target)
	require.Equal(t, IconHome, result[0].Icon)
	require.True(t, result[0].Disabled)
	require.Equal(t, "3", result[0].Badge)
	require.Equal(t, SwapOuterHTML, result[0].Swap)
}

// ---------------------------------------------------------------------------
// TabItem fields
// ---------------------------------------------------------------------------

func TestTabItem_ZeroValue(t *testing.T) {
	var tab TabItem
	require.Empty(t, tab.Label)
	require.Empty(t, tab.Href)
	require.Empty(t, tab.Target)
	require.Empty(t, tab.Icon)
	require.False(t, tab.Active)
	require.False(t, tab.Disabled)
	require.Empty(t, tab.Badge)
	require.Empty(t, string(tab.Swap))
}

func TestTabConfig_ItemTargetOverridesDefault(t *testing.T) {
	tabs := NewTabConfig("tabs", "#default-panel",
		TabItem{Label: "A", Href: "/a"},
		TabItem{Label: "B", Href: "/b", Target: "#custom"},
	)
	require.Equal(t, "#default-panel", tabs.Target)
	require.Empty(t, tabs.Items[0].Target, "item A uses shared default")
	require.Equal(t, "#custom", tabs.Items[1].Target, "item B has its own target")
}

// ---------------------------------------------------------------------------
// Integration: NewTabConfig + SetActiveTab
// ---------------------------------------------------------------------------

func TestNewTabConfig_WithSetActiveTab(t *testing.T) {
	tabs := NewTabConfig("user-tabs", "#tab-content",
		TabItem{Label: "Overview", Href: "/users/42/overview", Icon: "user"},
		TabItem{Label: "Activity", Href: "/users/42/activity", Icon: "clock"},
		TabItem{Label: "Settings", Href: "/users/42/settings", Icon: "cog"},
	)
	tabs.Items = SetActiveTab(tabs.Items, "/users/42/activity")

	require.Equal(t, "user-tabs", tabs.ID)
	require.Equal(t, "#tab-content", tabs.Target)
	require.Len(t, tabs.Items, 3)
	require.False(t, tabs.Items[0].Active)
	require.True(t, tabs.Items[1].Active)
	require.False(t, tabs.Items[2].Active)
	require.Equal(t, Icon("clock"), tabs.Items[1].Icon)
}
