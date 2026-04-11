# linkwell

![image](https://github.com/catgoose/screenshots/blob/main/linkwell/linkwell.png)

<!--toc:start-->

- [linkwell](#linkwell)
  - [The Problem](#the-problem)
  - [The Fix](#the-fix)
  - [Install](#install)
  - [Link Registry](#link-registry)
    - [Registering Links](#registering-links)
    - [Ring (Symmetric Group)](#ring-symmetric-group)
    - [Hub (Star Topology)](#hub-star-topology)
    - [Querying Links](#querying-links)
    - [RFC 8288 Link Header](#rfc-8288-link-header)
    - [Title From Path](#title-from-path)
    - [Dynamic Links](#dynamic-links)
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
  - [Tabs](#tabs)
  - [Filters](#filters)
    - [FilterBar](#filterbar)
    - [Field Types](#field-types)
    - [FilterGroup](#filtergroup)
  - [Tables and Pagination](#tables-and-pagination)
    - [Sortable Columns](#sortable-columns)
    - [Pagination](#pagination)
  - [Modals](#modals)
  - [Toasts](#toasts)
  - [Stepper](#stepper)
  - [Action Patterns](#action-patterns)
    - [Resource Actions](#resource-actions)
    - [Row Actions](#row-actions)
    - [Form Actions](#form-actions)
    - [Bulk Actions](#bulk-actions)
    - [Empty State and Catalog](#empty-state-and-catalog)
  - [Error Controls](#error-controls)
  - [Graph Validation](#graph-validation)
  - [Sitemap](#sitemap)
  - [Speculation Rules](#speculation-rules)
  - [Thread Safety](#thread-safety)
  - [Testing](#testing)
  - [Recipes](#recipes)
  - [Philosophy](#philosophy)
  - [Architecture](#architecture)
    - [How linkwell drives navigation](#how-linkwell-drives-navigation)
  - [License](#license)
  <!--toc:end-->

[![Go Reference](https://pkg.go.dev/badge/github.com/catgoose/linkwell.svg)](https://pkg.go.dev/github.com/catgoose/linkwell)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![linkwell](https://raw.githubusercontent.com/catgoose/screenshots/main/linkwell/linkwell.png)

> WHERE ARE THE LINKS, KEVIN?
>
> -- The PENTAVERB

A Go library for HATEOAS-style hypermedia controls, link relations ([RFC 8288](https://www.rfc-editor.org/rfc/rfc8288)), and navigation primitives. Designed for server-rendered HTML apps using HTMX, but the data types are framework-agnostic.

Your JSON endpoint returns data. Your HTML page returns data AND WHAT TO DO WITH IT. linkwell is the library that makes "what to do with it" a first-class concern.

## The Problem

> Big Brain Developer say "I have built a micro-frontend architecture with seventeen independently deployable SPAs, each with its own state management solution, communicating through a custom event bus with schema validation."
>
> Grug say "what it do"
>
> Big Brain Developer say "it renders a table of users."
>
> -- The Recorded Sayings of Layman Grug

Every server-rendered app reinvents the same things: nav bars with active states, breadcrumbs stitched together by hand, pagination controls, filter forms, modal configs, action buttons with confirmation dialogs. Each handler builds its own structs. Each project copies the last one's approach and changes just enough to break it.

```go
// Hardcoded nav, scattered across handlers
type NavItem struct {
	Label, Href string
	Active      bool
}
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

You have constructed an ENORMOUS and MAGNIFICENT cathedral of boilerplate for the sole purpose of avoiding a shared vocabulary for hypermedia controls.

## The Fix

> "before enlightenment: fetch JSON, parse JSON, validate JSON, transform JSON, store JSON in client state, derive view from client state, diff virtual DOM, reconcile DOM."
>
> "and after enlightenment?"
>
> "`hx-get`"
>
> -- Layman Grug

Declare your link graph once. Query it at request time. Let the server drive the state.

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
- **Hypermedia controls** -- pure-data descriptors for buttons, actions, and navigation affordances
- **Tabs** for in-page tabbed navigation with HTMX lazy-loading
- **Filter, table, and pagination** primitives for data-heavy views
- **Modal** configuration with preset button sets
- **Toast** notifications for success/info/warning/error feedback
- **Stepper** for multi-step wizard flows with auto-generated navigation controls
- **Navigation** types with active-state computation
- **Error controls** dispatched by HTTP status code
- **Graph validation** for catching link registration bugs at test time
- **Sitemap** generation from the link registry

No rendering logic. No framework coupling. Just data structures that your templates consume. The server decides what actions are available, and the controls carry that decision to the HTML.

## Install

```bash
go get github.com/catgoose/linkwell
```

Import with an alias if you prefer:

```go
import hypermedia "github.com/catgoose/linkwell"
```

## Link Registry

> The server sends a representation. The representation contains links and forms. The client follows them. THAT IS THE ENTIRE INTERACTION MODEL.
>
> -- The Wisdom of the Uniform Interface

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

`LinksFor` walks parent path segments when the exact path has no registered links, so `/admin/apps/1` inherits links from `/admin/apps`.

```go
// All links for a path (walks parent paths if needed)
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

Format links as a standard `Link` HTTP header. Titles are properly escaped per RFC 7230 quoted-string rules.

```go
links := linkwell.LinksFor("/inventory")
header := linkwell.LinkHeader(links)
// </warehouses>; rel="related"; title="Warehouses", ...
```

Named constants cover common IANA relation types: `RelRelated`, `RelUp`, `RelSelf`, `RelAlternate`, `RelCanonical`, `RelFirst`, `RelLast`, `RelNext`, `RelPrev`, `RelCollection`, `RelItem`.

### Title From Path

> A resource is any concept important enough to be named. And naming is the first and most dangerous act of computing.
>
> -- The Wisdom of the Uniform Interface

`TitleFromPath` derives a human-readable title from the last segment of a URL path. Hyphens become spaces and each word is title-cased. Useful for auto-generating labels from route paths when no explicit title is registered.

```go
linkwell.TitleFromPath("/demo/inventory")     // "Inventory"
linkwell.TitleFromPath("/admin/error-traces")  // "Error Traces"
```

This is the same logic used internally by `BreadcrumbsFromPath` to label segments.

### Dynamic Links

Load and remove links at runtime (e.g., from a database):

```go
linkwell.LoadStoredLink("/projects/42", linkwell.LinkRelation{
	Rel: "related", Href: "/teams/7", Title: "Backend Team",
})

linkwell.RemoveLink("/projects/42", "/teams/7", "related")
```

## Breadcrumbs

> The links are RIGHT THERE. In the HTML. They have been there this whole time. You have been stepping over them to get to your OpenAPI generator.
>
> -- The Wisdom of the Uniform Interface

### From Link Graph

Walk the `rel="up"` chain from a path to build a breadcrumb trail. Requires links registered via `Hub` or `Link` with `rel="up"`. Registered titles are preserved -- if a spoke was registered with a custom title, that title appears in the breadcrumb instead of the path-derived label.

```go
// Given: Hub("/admin", "Admin", Rel("/admin/users", "User Directory"))
crumbs := linkwell.BreadcrumbsFromLinks("/admin/users")
// [{Label:"Home" Href:"/"}, {Label:"Admin" Href:"/admin"}, {Label:"User Directory" Href:""}]
```

The terminal breadcrumb has an empty `Href` (current page, rendered as text).

### From URL Path

Generate breadcrumbs from URL segments. Each segment is title-cased (hyphens
become spaces, first letter of each word capitalised). Override labels by
segment index.

```go
crumbs := linkwell.BreadcrumbsFromPath("/users/42/edit", map[int]string{
	1: "Jane Doe", // override "42" with a name
})
// [{Label:"Home" Href:"/"}, {Label:"Users" Href:"/users"},
//  {Label:"Jane Doe" Href:"/users/42"}, {Label:"Edit" Href:""}]
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
> -- The Wisdom of the Uniform Interface

A `Control` is a pure-data descriptor for a hypermedia affordance (button, link, action). Templates consume controls to render the appropriate HTML element -- the control itself has no rendering logic.

The `<button>` with an `hx-delete` says "here is something you can destroy, and here is exactly how to destroy it, and here is where the confirmation dialog will appear, and none of this required a README." linkwell makes your server produce those controls as data, so your templates just render them.

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

## Tabs

> _How many web browsers know the difference between a banking application and a wiki? None of them. And yet they operate both. Your browser speaks HTTP, understands media types, and follows links. Three things._
>
> -- The Wisdom of the Uniform Interface

`TabConfig` and `TabItem` describe in-page tabbed navigation. Structurally similar to `NavItem` but scoped to a content panel with HTMX lazy-load semantics. The server decides which tabs exist and which is active.

```go
tabs := linkwell.NewTabConfig("user-tabs", "#tab-content",
	linkwell.TabItem{Label: "Overview", Href: "/users/42/overview", Icon: "user"},
	linkwell.TabItem{Label: "Activity", Href: "/users/42/activity", Icon: "clock"},
	linkwell.TabItem{Label: "Settings", Href: "/users/42/settings", Icon: "cog"},
)
tabs.Items = linkwell.SetActiveTab(tabs.Items, "/users/42/activity")
```

Each `TabItem` supports:

- `Target` -- per-tab hx-target override (falls back to `TabConfig.Target`)
- `Badge` -- optional count/status indicator
- `Swap` -- HTMX swap strategy (defaults to innerHTML)
- `Disabled` -- non-interactive state

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

> "You been repeating yourself since 2013. Every year new hook. Every year same table."
>
> -- The Recorded Sayings of Layman Grug

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

| Set                    | Buttons                  |
| ---------------------- | ------------------------ |
| `ModalOK`              | OK                       |
| `ModalYesNo`           | No, Yes                  |
| `ModalSaveCancel`      | Cancel, Save             |
| `ModalSaveCancelReset` | Reset, Cancel, Save      |
| `ModalSubmitCancel`    | Cancel, Submit           |
| `ModalConfirmCancel`   | Cancel, Confirm (danger) |
| `ModalDeleteCancel`    | Cancel, Delete (danger)  |

Report issue modal shortcut:

```go
modal := linkwell.ReportIssueModal(requestID)
```

## Toasts

> _past is already past -- don't debug it_
>
> _future not here yet -- don't optimize for it_
>
> _server return html -- this present moment_
>
> -- The Recorded Sayings of Layman Grug, [The Dothog Manifesto](https://github.com/catgoose/dothog/blob/main/MANIFESTO.md)

`Toast` is the success/info/warning complement to `ErrorContext`. The server decides what feedback to show and the template renders the appropriate notification. Toasts are value types -- use the `With*` methods to derive modified copies.

```go
// Factory functions for each variant
toast := linkwell.SuccessToast("User created")
toast := linkwell.InfoToast("Export queued")
toast := linkwell.WarningToast("API rate limit approaching")
toast := linkwell.ErrorToast("Upload failed")
```

Builder chain for a full notification with an undo action:

```go
toast := linkwell.SuccessToast("User deleted").
	WithControls(linkwell.HTMXAction("Undo", linkwell.HxPost("/users/42/restore", "#user-table"))).
	WithAutoDismiss(5).
	WithOOB("#toast-container", "afterbegin")
```

| Field         | Purpose                                              |
| ------------- | ---------------------------------------------------- |
| `Message`     | User-visible notification text                       |
| `Variant`     | Visual style: success, info, warning, error          |
| `Controls`    | Optional action affordances (e.g., Undo button)      |
| `AutoDismiss` | Seconds before auto-close (0 = sticky)               |
| `OOBTarget`   | CSS selector for HTMX OOB swap target                |
| `OOBSwap`     | hx-swap-oob strategy (e.g., "afterbegin")            |

## Stepper

> _The client receives a page, sees what it can do, and does it. Like a person. Using a website._
>
> -- The Wisdom of the Uniform Interface

`StepperConfig` describes a multi-step wizard flow where the server knows the full step sequence, current position, and completion state. Navigation controls are auto-generated.

```go
stepper := linkwell.NewStepper(1, // currently on step 2 (0-indexed)
	linkwell.Step{Label: "Account", Href: "/onboard/account", Icon: "user"},
	linkwell.Step{Label: "Profile", Href: "/onboard/profile", Icon: "id-card"},
	linkwell.Step{Label: "Preferences", Href: "/onboard/prefs", Icon: "sliders"},
	linkwell.Step{Label: "Review", Href: "/onboard/review", Icon: "check-circle"},
)
```

`NewStepper` auto-computes:

- Steps before the current index are marked `StepComplete`
- The current step is marked `StepActive`
- Steps after are marked `StepPending`
- Pre-set statuses like `StepSkipped` are preserved
- `Prev` control points to the previous step (nil on first)
- `Next` control points to the next step (nil on last)
- `Submit` control appears only on the final step (with `VariantPrimary`)

Invalid `currentIndex` values are clamped into the valid range `[0, len(steps)-1]`. Empty steps return an empty config with no navigation controls.

## Action Patterns

> HATEOAS: Hypermedia As The Engine Of Application State. Yes, it is an ugly acronym. The truth is not always beautiful. Sometimes the truth is an ugly acronym that you should have tattooed on the inside of your eyelids.
>
> -- The Wisdom of the Uniform Interface

Action patterns are HATEOAS made concrete -- the server decides what actions are available, and the controls carry that decision to the template. The representation tells the client what is possible RIGHT NOW. When what is possible changes, the representation changes, and the client adapts.

All pattern functions conditionally include controls based on which URLs are provided -- omit a URL to hide that action.

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

> Guilt is a client-side state and the server does not care about client-side state. The server has never cared about client-side state. This is the First Lesson and also the Last Lesson.
>
> -- The PENTAVERB

Status-code-specific control sets for error pages:

```go
// Dispatch by status code
controls := linkwell.ErrorControlsForStatus(404, linkwell.ErrorControlOpts{
	HomeURL: "/",
})

// Or use individual builders
controls := linkwell.NotFoundControls("/")          // [Back, GoHome]
controls := linkwell.ServiceErrorControls(opts)     // [Retry?, Dismiss]
controls := linkwell.UnauthorizedControls("/login") // [Log In?, Dismiss]
controls := linkwell.ForbiddenControls()            // [Back, Dismiss]
controls := linkwell.InternalErrorControls(opts)    // [Retry?, Dismiss]
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

## Graph Validation

> If your client must read your API docs to know which URL to `POST` to, that is out-of-band. If your client must be recompiled when you rename a resource, you have coupled the client to the server's URI structure and you will maintain this coupling in blood and tears until one of you is decommissioned.
>
> -- The Wisdom of the Uniform Interface

Link registration bugs are silent -- a typo in a path means a breadcrumb chain breaks or a related link disappears. `ValidateGraph` catches structural issues at test time before they reach production.

```go
func TestLinkGraphIntegrity(t *testing.T) {
	setupRoutes(e)

	issues := linkwell.ValidateGraph()
	for _, issue := range issues {
		t.Errorf("link graph: %s -- %s (%s)", issue.Path, issue.Message, issue.Kind)
	}
}
```

Detected issues:

| Kind               | Description                                         |
| ------------------ | --------------------------------------------------- |
| `orphan`           | Path with no inbound links from any other path      |
| `broken_up`        | `rel="up"` target is not registered                 |
| `dead_spoke`       | Hub spoke path has no registered links               |

Validate that registered paths match your router's route set:

```go
func TestAllLinksMatchRoutes(t *testing.T) {
	router := setupRouter()
	routes := extractRoutes(router)

	issues := linkwell.ValidateAgainstRoutes(routes)
	for _, issue := range issues {
		t.Errorf("route mismatch: %s -- %s", issue.Path, issue.Message)
	}
}
```

| Kind                 | Description                                        |
| -------------------- | -------------------------------------------------- |
| `unregistered_route` | Link graph path (source or target) has no matching route |
| `missing_route`      | Route has no link graph presence                    |

## Sitemap

Derive a structured sitemap from the link registry. The hub/spoke/ring topology already contains the page hierarchy -- `Sitemap` exposes it as a queryable data type without maintaining a separate definition. Target-only pages (paths that appear only as link targets, never as source keys) are included automatically, so every reachable page in the graph has a sitemap entry.

```go
// Full sitemap sorted by path
entries := linkwell.Sitemap()

// Only top-level entries (no parent)
roots := linkwell.SitemapRoots()

// HTML sitemap page
for _, entry := range entries {
	fmt.Printf("%s (%s)\n", entry.Path, entry.Title)
	if entry.Parent != "" {
		fmt.Printf("  parent: %s\n", entry.Parent)
	}
	for _, child := range entry.Children {
		fmt.Printf("  child: %s\n", child)
	}
}

// XML sitemap generation
for _, entry := range roots {
	fmt.Fprintf(w, "<url><loc>%s%s</loc></url>\n", baseURL, entry.Path)
}
```

Each `SitemapEntry` provides:

| Field      | Source                                              |
| ---------- | --------------------------------------------------- |
| `Path`     | Registered path (includes target-only pages)        |
| `Title`    | Hub title, registered link title, or derived from path |
| `Parent`   | `rel="up"` target (empty for roots)                 |
| `Children` | Hub spoke paths                                     |
| `Group`    | Ring group name                                     |

## Speculation Rules

The [Speculation Rules API](https://developer.mozilla.org/en-US/docs/Web/API/Speculation_Rules_API) lets browsers prefetch or prerender pages before the user navigates, declared as JSON inside a `<script type="speculationrules">` block. No JavaScript required -- the browser handles it natively.

linkwell's link registry already describes the navigation graph. Hub spokes are natural prefetch candidates (list-to-detail navigation), and Ring members are moderate candidates (peer navigation). The same data that drives breadcrumbs and sitemaps can drive speculation rules.

**Why this lives in application code, not linkwell:** linkwell is a data library with no rendering logic and no browser-specific concerns. Speculation rules are an output format -- the same way linkwell doesn't generate HTML but provides data for templates to consume, it doesn't generate `<script>` tags but provides the navigation data you need to build them.

Build speculation rules from the registry in your application:

```go
// speculationRules builds a Speculation Rules JSON string from linkwell's
// link registry. Hub spokes get prefetched on hover (moderate eagerness),
// giving users near-instant navigation to detail pages.
func speculationRules() string {
    var patterns []string
    for _, hub := range linkwell.Hubs() {
        for _, spoke := range hub.Spokes {
            patterns = append(patterns, spoke.Href+"*")
        }
    }
    if len(patterns) == 0 {
        return ""
    }
    rules := map[string]any{
        "prefetch": []map[string]any{{
            "where":    map[string]any{"href_matches": patterns},
            "eagerness": "moderate",
        }},
    }
    b, _ := json.Marshal(rules)
    return string(b)
}
```

Emit it in your base layout template:

```html
<!-- In your base layout template -->
{{ if $rules := speculationRules }}
<script type="speculationrules">{{ $rules }}</script>
{{ end }}
```

**Browser support:** Chrome 121+, Edge 121+. Other browsers ignore the `<script type="speculationrules">` tag entirely, so this is progressive enhancement -- free to add, zero cost when unsupported.

## Thread Safety

All registry operations (`Link`, `Ring`, `Hub`, `LinksFor`, `AllLinks`, `Hubs`,
`LoadStoredLink`, `RemoveLink`) are protected by `sync.RWMutex` and are safe for
concurrent use. The typical pattern is init-time registration (call `Link`,
`Ring`, and `Hub` during route setup before the server starts accepting
requests), then read with `LinksFor`, `AllLinks`, `Hubs`, etc. at request time.

`RegisterFrom` and `ResolveFromMask` are similarly protected and safe for
concurrent use.

## Testing

Use `ResetForTesting` to clear all registries (links, hubs, and
breadcrumb-origin registrations) between tests. The Home breadcrumb (bit 0) is
re-registered automatically. It is intended for test setup/teardown only and
must not be called concurrently with request handlers. In parallel tests, call
it at the start of each subtest and register it with `t.Cleanup`:

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

## Recipes

See [recipes.md](recipes.md) for integration patterns showing linkwell and [tavern](https://github.com/catgoose/tavern) working together in server-rendered HTMX apps. Recipes cover live dashboards, real-time tables, delete-with-broadcast, scoped notifications, lifecycle-aware publishing, and multi-step wizard flows.

## Philosophy

> _THE FOOL asked: "What is HATEOAS?" Hypermedia As The Engine Of Application State. The server sends a representation. The representation contains links and forms. The client follows them. THAT IS THE ENTIRE INTERACTION MODEL._
>
> -- The Wisdom of the Uniform Interface

linkwell is that interaction model, made concrete in Go. The server decides what controls to present. The representation carries them. The client follows them. linkwell follows the [dothog design philosophy](https://github.com/catgoose/dothog/blob/main/PHILOSOPHY.md).

## Architecture

### How linkwell drives navigation

```
  startup                          request time
  ───────                          ────────────

  Hub("/admin", "Admin",           links := LinksFor("/admin/users")
    Rel("/admin/users", ...),      crumbs := BreadcrumbsFromLinks(path)
    Rel("/admin/roles", ...),      controls := ResourceActions(cfg)
  )                                         │
       │                                    v
       v                              ┌───────────┐
  ┌──────────┐                        │  template  │
  │ registry │ ◄── query at ──────►   │  renders   │
  │ (links)  │     request time       │  controls  │
  └──────────┘                        └───────────┘
```

> Enter the application with a single URI and a set of standardized media types. Follow the links. Submit the forms. Let the server drive the state. That is all.
>
> -- The Wisdom of the Uniform Interface

That is linkwell's entire design. There is no conclusion. There is only the next request.

## License

MIT
