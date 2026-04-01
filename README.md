# linkwell

[![Go Reference](https://pkg.go.dev/badge/github.com/catgoose/linkwell.svg)](https://pkg.go.dev/github.com/catgoose/linkwell)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![linkwell](https://raw.githubusercontent.com/catgoose/screenshots/main/linkwell/linkwell.png)

> WHERE ARE THE LINKS, KEVIN?
>
> -- The PENTAVERB, Dothog Manifesto

A Go library for HATEOAS-style hypermedia controls, link relations ([RFC 8288](https://www.rfc-editor.org/rfc/rfc8288)), and navigation primitives. Designed for server-rendered HTML apps using HTMX, but the data types are framework-agnostic.

## Why

**Without linkwell:**

```go
// Hardcoded nav, scattered across handlers
type NavItem struct{ Label, Href string; Active bool }
nav := []NavItem{
    {"Users", "/admin/users", path == "/admin/users"},
    {"Roles", "/admin/roles", path == "/admin/roles"},
    {"Settings", "/admin/settings", path == "/admin/settings"},
}

// Breadcrumbs built by hand, per handler
crumbs := []Breadcrumb{
    {Label: "Home", Href: "/"},
    {Label: "Admin", Href: "/admin"},
    {Label: "Users", Href: ""},
}

// Pagination, filters, sort columns, modals, error controls --
// all custom structs, repeated in every project.
```

**With linkwell:**

```go
// Register once at startup
linkwell.Hub("/admin", "Admin",
    linkwell.Rel("/admin/users", "Users"),
    linkwell.Rel("/admin/roles", "Roles"),
    linkwell.Rel("/admin/settings", "Settings"),
)

// At request time
crumbs := linkwell.BreadcrumbsFromLinks("/admin/users")
related := linkwell.RelatedLinksFor("/admin/users")
controls := linkwell.ResourceActions(linkwell.ResourceActionCfg{
    EditURL: "/admin/users/42/edit", DeleteURL: "/admin/users/42",
    Target: "#content", ConfirmMsg: "Delete this user?",
})
```

linkwell provides:

- A **link registry** for declaring relationships between pages (related, parent/child, hub-and-spoke)
- **Breadcrumb** generation from link graphs or URL paths
- **Hypermedia controls** — pure-data descriptors for buttons, actions, and navigation affordances
- **Filter, table, and pagination** primitives for data-heavy views
- **Modal** configuration with preset button sets
- **Navigation** types with active-state computation
- **Error controls** dispatched by HTTP status code

## Install

```bash
go get github.com/catgoose/linkwell
```

Import with an alias if you prefer:

```go
import hypermedia "github.com/catgoose/linkwell"
```

## Table of Contents

- [Link Registry](#link-registry)
  - [Registering Links](#registering-links)
  - [Ring (Symmetric Group)](#ring-symmetric-group)
  - [Hub (Star Topology)](#hub-star-topology)
  - [Querying Links](#querying-links)
  - [RFC 8288 Link Header](#rfc-8288-link-header)
- [Breadcrumbs](#breadcrumbs)
  - [From Link Graph](#from-link-graph)
  - [From URL Path](#from-url-path)
  - [Bitmask Breadcrumbs](#bitmask-breadcrumbs)
- [Controls](#controls)
  - [Factory Functions](#factory-functions)
  - [Control Modifiers](#control-modifiers)
  - [HTMX Request Config](#htmx-request-config)
- [Navigation](#navigation)
  - [NavConfig and NavItem](#navconfig-and-navitem)
  - [Active State](#active-state)
- [Filters](#filters)
  - [FilterBar](#filterbar)
  - [Field Types](#field-types)
  - [FilterGroup](#filtergroup)
- [Tables and Pagination](#tables-and-pagination)
  - [Sortable Columns](#sortable-columns)
  - [Pagination](#pagination)
- [Modals](#modals)
- [Action Patterns](#action-patterns)
  - [Resource Actions](#resource-actions)
  - [Row Actions](#row-actions)
  - [Form Actions](#form-actions)
  - [Bulk Actions](#bulk-actions)
- [Error Controls](#error-controls)
- [Thread Safety](#thread-safety)
- [Testing](#testing)

## Link Registry

> The server sends a representation. The representation contains links and forms. The client follows them. THAT IS THE ENTIRE INTERACTION MODEL.
>
> -- The Wisdom of the Uniform Interface, Dothog Manifesto

The link registry maintains an in-memory graph of relationships between pages. Register links at startup (typically in route initialization), then query them at request time.

### Registering Links

`Link` registers a directional relationship. For `rel="related"`, the inverse is automatically created (symmetric).

```go
// Inventory links to Warehouses
linkwell.Link("/inventory", "related", "/warehouses", "Warehouses")
// The inverse "/warehouses" -> "/inventory" is auto-registered

// One-way parent link
linkwell.Link("/users/42", "up", "/users", "Users")
```

### Ring (Symmetric Group)

`Ring` connects every member to every other member with `rel="related"`. Use for peer pages that should cross-link.

```go
linkwell.Ring("Logistics",
    linkwell.Rel("/inventory", "Inventory"),
    linkwell.Rel("/warehouses", "Warehouses"),
    linkwell.Rel("/shipments", "Shipments"),
)
// Each page now links to the other two with Group="Logistics"
```

### Hub (Star Topology)

`Hub` connects a center page to spokes. Spokes link back to the center with `rel="up"` but do not link to each other.

```go
linkwell.Hub("/admin", "Admin",
    linkwell.Rel("/admin/users", "Users"),
    linkwell.Rel("/admin/roles", "Roles"),
    linkwell.Rel("/admin/settings", "Settings"),
)
```

Query all hubs for site map rendering:

```go
hubs := linkwell.Hubs() // []HubEntry sorted by path
for _, hub := range hubs {
    fmt.Println(hub.Title, hub.Path)
    for _, spoke := range hub.Spokes {
        fmt.Println("  ", spoke.Title, spoke.Href)
    }
}
```

### Querying Links

```go
// All links for a path
links := linkwell.LinksFor("/inventory")

// Only specific relation types
related := linkwell.LinksFor("/inventory", "related")
parents := linkwell.LinksFor("/admin/users", "up")

// Related links excluding self (for context bars)
peers := linkwell.RelatedLinksFor("/inventory")

// Full registry snapshot (for admin/debug)
all := linkwell.AllLinks()
```

### RFC 8288 Link Header

Format links as a standard `Link` HTTP header:

```go
links := linkwell.LinksFor("/inventory")
header := linkwell.LinkHeader(links)
// </warehouses>; rel="related"; title="Warehouses", ...
```

### Dynamic Links

Load and remove links at runtime (e.g., from a database):

```go
linkwell.LoadStoredLink("/projects/42", linkwell.LinkRelation{
    Rel: "related", Href: "/teams/7", Title: "Backend Team",
})

linkwell.RemoveLink("/projects/42", "/teams/7", "related")
```

## Breadcrumbs

> The links are RIGHT THERE. In the HTML. They have been there this whole time.
>
> -- The Wisdom of the Uniform Interface, Dothog Manifesto

### From Link Graph

Walk the `rel="up"` chain from a path to build a breadcrumb trail. Requires links registered via `Hub` or `Link` with `rel="up"`.

```go
// Given: Hub("/admin", "Admin", Rel("/admin/users", "Users"))
crumbs := linkwell.BreadcrumbsFromLinks("/admin/users")
// [{Label:"Home" Href:"/"}, {Label:"Admin" Href:"/admin"}, {Label:"Users" Href:""}]
```

The terminal breadcrumb has an empty `Href` (current page, rendered as text).

### From URL Path

Generate breadcrumbs from URL segments. Override labels by segment index.

```go
crumbs := linkwell.BreadcrumbsFromPath("/users/42/edit", map[int]string{
    1: "Jane Doe", // override "42" with a name
})
// [{Label:"Home" Href:"/"}, {Label:"users" Href:"/users"},
//  {Label:"Jane Doe" Href:"/users/42"}, {Label:"edit" Href:""}]
```

### Bitmask Breadcrumbs

For pages reachable from multiple parents, use a `?from=` bitmask to preserve navigation context.

```go
// Register origins at startup
linkwell.RegisterFrom(linkwell.FromDashboard, linkwell.Breadcrumb{
    Label: "Dashboard", Href: "/dashboard",
})

// In handler: parse the ?from= param and resolve breadcrumbs
mask := linkwell.ParseFromParam(c.QueryParam("from"))
crumbs := linkwell.ResolveFromMask(mask)
// Home is always included at bit 0

// Forward context to outbound links
href := linkwell.FromNav("/users/42", c.QueryParam("from"))
// "/users/42?from=3" if mask was 3
```

## Controls

> Hypertext is the simultaneous presentation of information and controls such that the information BECOMES THE AFFORDANCE through which choices are obtained and actions are selected.
>
> -- The Wisdom of the Uniform Interface, Dothog Manifesto

A `Control` is a pure-data descriptor for a hypermedia affordance (button, link, action). Templates consume controls to render the appropriate HTML element -- the control itself has no rendering logic.

### Factory Functions

```go
// HTMX retry button
ctrl := linkwell.RetryButton("Retry", linkwell.HxMethodGet, "/api/data", "#content")

// Danger button with confirmation dialog
ctrl := linkwell.ConfirmAction("Delete", linkwell.HxMethodDelete, "/users/42", "#user-list", "Delete this user?")

// Browser back button (no server round-trip)
ctrl := linkwell.BackButton("Go Back")

// Home navigation
ctrl := linkwell.GoHomeButton("Go Home", "/", "body")

// Plain anchor link
ctrl := linkwell.RedirectLink("View Profile", "/users/42")

// Arbitrary HTMX action
ctrl := linkwell.HTMXAction("Archive", linkwell.HxPost("/users/42/archive", "#content"))

// Dismiss button (HyperScript close)
ctrl := linkwell.DismissButton("Close")

// Report issue modal trigger
ctrl := linkwell.ReportIssueButton("Report Issue", requestID)
```

### Control Modifiers

Controls are value types. Modifiers return copies.

```go
ctrl := linkwell.RetryButton("Retry", linkwell.HxMethodGet, "/api/data", "#content").
    WithSwap(linkwell.SwapOuterHTML).
    WithVariant(linkwell.VariantDanger).
    WithIcon(linkwell.IconCheck).
    WithConfirm("Are you sure?").
    WithDisabled(true).
    WithErrorTarget("#inline-error")
```

### HTMX Request Config

Build HTMX request attributes programmatically:

```go
req := linkwell.HxGet("/users", "#user-list")
req = req.WithInclude("closest form")

// Or build manually
req := linkwell.HxRequestConfig{
    Method:  linkwell.HxMethodPost,
    URL:     "/users",
    Target:  "#user-list",
    Include: "closest form",
    Vals:    `{"status":"active"}`,
}

// Convert to attribute map for interop
attrs := req.Attrs() // map[string]string{"get": "/users", "target": "#user-list", ...}
```

## Navigation

### NavConfig and NavItem

`NavConfig` holds the app navigation layout. `NavItem` is a single navigation entry with optional HTMX attributes and children.

```go
nav := linkwell.NavConfig{
    AppName:    "My App",
    MaxVisible: 5, // overflow after 5 items
    Items: []linkwell.NavItem{
        {Label: "Dashboard", Href: "/dashboard", Icon: "home"},
        {Label: "Users", Href: "/users", Icon: "users", Children: []linkwell.NavItem{
            {Label: "Active", Href: "/users?status=active"},
            {Label: "Invited", Href: "/users?status=invited"},
        }},
    },
}
```

Bridge a `Control` to a `NavItem`:

```go
navItem := linkwell.NavItemFromControl(ctrl)
```

### Active State

Set the active nav item based on the current request path. Exact match:

```go
items := linkwell.SetActiveNavItem(nav.Items, "/users")
```

Prefix match (section-level navigation):

```go
// "/users" is active when path is "/users/42/edit"
items := linkwell.SetActiveNavItemPrefix(nav.Items, "/users/42/edit")
```

Both functions handle nested children: a parent is marked active if any child matches.

## Filters

### FilterBar

`FilterBar` describes a filter form. The form ID enables `hx-include` from pagination/sort links.

```go
bar := linkwell.NewFilterBar("/users", "#user-table",
    linkwell.SearchField("q", "Search users...", currentQuery),
    linkwell.SelectField("status", "Status", currentStatus, linkwell.SelectOptions(currentStatus,
        "", "All",
        "active", "Active",
        "inactive", "Inactive",
    )),
    linkwell.CheckboxField("verified", "Verified only", verifiedParam),
)
```

### Field Types

```go
linkwell.SearchField(name, placeholder, value)
linkwell.SelectField(name, label, value, options)
linkwell.RangeField(name, label, value, min, max, step)
linkwell.CheckboxField(name, label, value) // value: "true" or ""
linkwell.DateField(name, label, value)
```

Build select options from flat pairs:

```go
opts := linkwell.SelectOptions(currentValue,
    "draft", "Draft",
    "published", "Published",
    "archived", "Archived",
)
```

### FilterGroup

`FilterGroup` wraps a `FilterBar` and supports dynamic option updates and OOB swap fragments.

```go
group := linkwell.NewFilterGroup("/products", "#product-table",
    linkwell.SearchField("q", "Search...", ""),
    linkwell.SelectField("category", "Category", "", categories),
)

// Update options dynamically
group.UpdateOptions("category", newCategories)

// Get only select fields for OOB rendering
selects := group.SelectFields()
```

## Tables and Pagination

### Sortable Columns

`SortableCol` creates a column descriptor with sort state and toggle URL precomputed.

```go
cols := []linkwell.TableCol{
    linkwell.SortableCol("name", "Name", sortKey, sortDir, baseURL, "#table", "#filter-form"),
    linkwell.SortableCol("email", "Email", sortKey, sortDir, baseURL, "#table", "#filter-form"),
    {Key: "actions", Label: "Actions"}, // non-sortable
}
```

Sort direction toggles: unsorted -> asc -> desc -> asc.

### Pagination

```go
totalPages := linkwell.ComputeTotalPages(totalItems, perPage)

info := linkwell.PageInfo{
    BaseURL:    "/users?q=foo",
    Page:       currentPage,
    PerPage:    25,
    TotalItems: totalItems,
    TotalPages: totalPages,
    Target:     "#user-table",
    Include:    "#filter-form",
}

controls := linkwell.PaginationControls(info)
// Returns nil if TotalPages <= 1
// Controls: [First] [Prev] [1] [2] [3] [Next] [Last]
// Current page is disabled with VariantPrimary
```

Generate a URL for a specific page:

```go
url := info.URLForPage(3) // "/users?q=foo&page=3"
```

## Modals

`ModalConfig` describes everything needed to render a modal dialog.

```go
modal := linkwell.ModalConfig{
    ID:       "delete-user-modal",
    Title:    "Delete User",
    Buttons:  linkwell.ModalDeleteCancel,
    HxPost:   "/users/42/delete",
    HxTarget: "#user-list",
    HxSwap:   linkwell.SwapOuterHTML,
}
```

Preset button sets:

| Set | Buttons |
|-----|---------|
| `ModalOK` | OK |
| `ModalYesNo` | No, Yes |
| `ModalSaveCancel` | Cancel, Save |
| `ModalSaveCancelReset` | Reset, Cancel, Save |
| `ModalSubmitCancel` | Cancel, Submit |
| `ModalConfirmCancel` | Cancel, Confirm (danger) |
| `ModalDeleteCancel` | Cancel, Delete (danger) |

Report issue modal shortcut:

```go
modal := linkwell.ReportIssueModal(requestID)
```

## Action Patterns

Pre-built control sets for common CRUD and data-management patterns. All pattern functions conditionally include controls based on which URLs are provided — omit a URL to hide that action.

### Resource Actions

Edit + Delete controls for resource detail pages:

```go
controls := linkwell.ResourceActions(linkwell.ResourceActionCfg{
    EditURL:     "/users/42/edit",
    DeleteURL:   "/users/42",
    ConfirmMsg:  "Delete this user?",
    Target:      "#content",
    ErrorTarget: "#error",
})
```

### Row Actions

Controls for table rows with different swap targets:

```go
// Edit swaps the row, Delete swaps the row
controls := linkwell.RowActions(linkwell.RowActionCfg{
    EditURL:     "/users/42/edit",
    DeleteURL:   "/users/42",
    RowTarget:   "#row-42",
    ConfirmMsg:  "Delete this user?",
    ErrorTarget: "#row-42-error",
})

// Edit swaps the row, Delete replaces the whole table
controls := linkwell.TableRowActions(linkwell.TableRowActionCfg{
    EditURL:     "/users/42/edit",
    DeleteURL:   "/users/42",
    RowTarget:   "#row-42",
    TableTarget: "#user-table",
    ConfirmMsg:  "Delete this user?",
})
```

Inline edit row controls:

```go
// Existing row: Save uses PUT
controls := linkwell.RowFormActions(linkwell.RowFormActionCfg{
    SaveURL:      "/users/42",
    CancelURL:    "/users/42",
    SaveTarget:   "#row-42",
    CancelTarget: "#row-42",
})

// New row: Save uses POST
controls := linkwell.NewRowFormActions(linkwell.RowFormActionCfg{
    SaveURL:      "/users",
    CancelURL:    "/users/new/cancel",
    SaveTarget:   "#new-row",
    CancelTarget: "#new-row",
})
```

### Form Actions

Save + Cancel for form footers. Save uses `hx-include="closest form"`.

```go
controls := linkwell.FormActions("/users")
```

### Bulk Actions

Toolbar controls for batch operations on checkbox-selected rows:

```go
controls := linkwell.BulkActions(linkwell.BulkActionCfg{
    DeleteURL:        "/users/bulk-delete",
    ActivateURL:      "/users/bulk-activate",
    DeactivateURL:    "/users/bulk-deactivate",
    TableTarget:      "#user-table",
    CheckboxSelector: ".user-checkbox",
})
```

### Empty State and Catalog

```go
ctrl := linkwell.EmptyStateAction("Create First User", "/users/new", "#content")
ctrl := linkwell.CatalogRowAction("/products/42/details", "#detail-row-42")
```

## Error Controls

Status-code-specific control sets for error pages:

```go
// Dispatch by status code
controls := linkwell.ErrorControlsForStatus(404, linkwell.ErrorControlOpts{
    HomeURL: "/",
})

// Or use individual builders
controls := linkwell.NotFoundControls("/")           // [Back, GoHome]
controls := linkwell.ServiceErrorControls(opts)       // [Retry?, Dismiss]
controls := linkwell.UnauthorizedControls("/login")   // [Log In?, Dismiss]
controls := linkwell.ForbiddenControls()              // [Back, Dismiss]
controls := linkwell.InternalErrorControls(opts)      // [Retry?, Dismiss]
```

`ErrorContext` carries the full error state through a rendering pipeline:

```go
ec := linkwell.ErrorContext{
    StatusCode: 500,
    Message:    "Database connection failed",
    Route:      "/api/users",
    RequestID:  "abc-123",
    Controls:   linkwell.InternalErrorControls(opts),
    Closable:   true,
}

// Fluent modifiers
ec = ec.WithControls(linkwell.ReportIssueButton("Report", "abc-123"))
ec = ec.WithOOB("#error-status", "innerHTML")

// Wrap as a returnable error
return linkwell.NewHTTPError(ec)
```

## Thread Safety

All registry operations (`Link`, `Ring`, `Hub`, `LinksFor`, `AllLinks`, `Hubs`,
`LoadStoredLink`, `RemoveLink`) are protected by `sync.RWMutex` and are safe for
concurrent use. The typical pattern is init-time registration (call `Link`,
`Ring`, and `Hub` during route setup before the server starts accepting
requests), then read with `LinksFor`, `AllLinks`, `Hubs`, etc. at request time.

`RegisterFrom` and `ResolveFromMask` are similarly protected and safe for
concurrent use.

## Testing

Use `ResetForTesting` to clear all registries between tests. It is intended for
test setup/teardown only and must not be called concurrently with request
handlers. In parallel tests, call it at the start of each subtest and register
it with `t.Cleanup`:

```go
func TestMyHandler(t *testing.T) {
    linkwell.ResetForTesting()
    t.Cleanup(linkwell.ResetForTesting)

    linkwell.Hub("/admin", "Admin",
        linkwell.Rel("/admin/users", "Users"),
    )
    // ... test logic
}
```

## License

MIT
