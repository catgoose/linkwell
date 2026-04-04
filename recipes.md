# Recipes: linkwell + tavern

linkwell describes **what** the UI should show (controls, navigation, pagination, feedback).
[tavern](https://github.com/catgoose/tavern) pushes **when** to update it (SSE pub/sub, OOB swaps).

These recipes show the two libraries working together in server-rendered HTMX apps. Each recipe focuses on the linkwell side — see tavern's own recipes for the complementary SSE/pub-sub perspective.

---

## Live dashboard with scheduled widgets

linkwell structures the page (nav, breadcrumbs, tabs); tavern's `ScheduledPublisher` ticks widget updates via OOB swaps.

### Page structure

```go
nav := linkwell.NavConfig{
    AppName: "Ops Dashboard",
    Items: linkwell.SetActiveNavItemPrefix([]linkwell.NavItem{
        {Label: "Dashboard", Href: "/dashboard", Icon: "chart-bar"},
        {Label: "Users", Href: "/users", Icon: "users"},
        {Label: "Settings", Href: "/settings", Icon: "cog"},
    }, "/dashboard"),
}

crumbs := []linkwell.Breadcrumb{
    {Label: "Home", Href: "/"},
    {Label: "Dashboard"},
}

tabs := linkwell.NewTabConfig("dashboard-tabs", "#tab-content",
    linkwell.TabItem{Label: "Network", Href: "/dashboard/network", Icon: "wifi"},
    linkwell.TabItem{Label: "Storage", Href: "/dashboard/storage", Icon: "server"},
    linkwell.TabItem{Label: "Compute", Href: "/dashboard/compute", Icon: "cpu"},
)
tabs.Items = linkwell.SetActiveTab(tabs.Items, "/dashboard/network")
```

### Scheduled widget updates (tavern side)

tavern's `ScheduledPublisher` renders each widget section on independent intervals and publishes the rendered HTML as OOB fragments. The browser receives these via SSE and HTMX swaps them into the page — no polling, no JavaScript state.

```go
// tavern publishes OOB fragments on a schedule:
//   tavern.Replace("#network-chart", renderedHTML)
//   tavern.Replace("#storage-gauge", renderedHTML)
//
// The browser connects with:
//   <div hx-ext="sse" sse-connect="/sse?topic=dashboard" sse-swap="message">
```

linkwell's job is done at initial render — nav, tabs, and breadcrumbs are static descriptors. tavern handles the live updates.

---

## Real-time table with filters and pagination

linkwell builds the filter bar, sortable columns, and pagination controls. tavern broadcasts row mutations to all connected viewers.

### Table setup

```go
// Filters
filters := linkwell.NewFilterBar("/users", "#user-table",
    linkwell.SearchField("q", "Search users...", currentQuery),
    linkwell.SelectField("role", "Role", currentRole,
        linkwell.SelectOptions(currentRole, "admin", "Admin", "user", "User", "guest", "Guest"),
    ),
    linkwell.SelectField("status", "Status", currentStatus,
        linkwell.SelectOptions(currentStatus, "active", "Active", "inactive", "Inactive"),
    ),
)

// Sortable columns
cols := []linkwell.TableCol{
    linkwell.SortableCol("name", "Name", sortKey, sortDir, baseURL, "#user-table", "#filter-form"),
    linkwell.SortableCol("email", "Email", sortKey, sortDir, baseURL, "#user-table", "#filter-form"),
    linkwell.SortableCol("created", "Created", sortKey, sortDir, baseURL, "#user-table", "#filter-form"),
    {Key: "actions", Label: "Actions"},
}

// Pagination
pageInfo := linkwell.PageInfo{
    BaseURL:    "/users",
    Page:       page,
    PerPage:    25,
    TotalItems: totalUsers,
    TotalPages: linkwell.ComputeTotalPages(totalUsers, 25),
    Target:     "#user-table",
    Include:    "#filter-form",
}
pagination := linkwell.PaginationControls(pageInfo)
```

### Row actions

```go
actions := linkwell.TableRowActions(linkwell.TableRowActionCfg{
    EditURL:     fmt.Sprintf("/users/%d/edit", user.ID),
    DeleteURL:   fmt.Sprintf("/users/%d", user.ID),
    RowTarget:   fmt.Sprintf("#user-%d", user.ID),
    TableTarget: "#user-table",
    ConfirmMsg:  fmt.Sprintf("Delete %s?", user.Name),
})
```

### Live row broadcasts (tavern side)

When any user creates, updates, or deletes a row, tavern broadcasts the change as OOB fragments to all viewers on the topic. linkwell provides the table structure and controls; tavern ensures every browser sees the mutation.

```go
// After a successful delete handler:
//   tavern.PublishOOB("users", tavern.Delete(fmt.Sprintf("#user-%d", id)))
//
// After a successful create/update handler:
//   tavern.PublishOOB("users", tavern.Replace("#user-table", renderedTable))
```

---

## Delete with broadcast

linkwell provides the delete control and error feedback. tavern removes the row from every connected browser.

### Handler flow

```go
func deleteUser(id int) {
    // linkwell: the delete control was rendered with ConfirmAction
    // The user confirmed and the request arrived here.

    if err := db.DeleteUser(id); err != nil {
        // linkwell: build error feedback
        ec := linkwell.ErrorContext{
            Err:        err,
            Message:    "Could not delete user",
            StatusCode: 500,
            Controls:   linkwell.InternalErrorControls(linkwell.ErrorControlOpts{
                RetryMethod: linkwell.HxMethodDelete,
                RetryURL:    fmt.Sprintf("/users/%d", id),
                RetryTarget: "#user-table",
            }),
        }.WithOOB("#error-status", "innerHTML")

        // render ec to response...
        return
    }

    // Success: notify the requesting browser
    toast := linkwell.SuccessToast("User deleted").
        WithAutoDismiss(5).
        WithOOB("#toast-container", "afterbegin")

    // render toast to response...

    // tavern: broadcast removal to all OTHER connected browsers
    //   tavern.PublishOOB("users", tavern.Delete(fmt.Sprintf("#user-%d", id)))
}
```

### The delete control (rendered earlier)

```go
deleteCtrl := linkwell.ConfirmAction(
    "Delete", linkwell.HxMethodDelete,
    fmt.Sprintf("/users/%d", user.ID),
    "#user-table",
    fmt.Sprintf("Delete %s?", user.Name),
)
```

---

## Scoped notifications

linkwell renders the notification UI (toasts, modals, error panels). tavern delivers events to specific users or resource scopes.

### Per-user toast via scoped SSE

```go
// In the handler that completes a long-running job:
toast := linkwell.SuccessToast("Export ready — download your file").
    WithControls(
        linkwell.RedirectLink("Download", fmt.Sprintf("/exports/%d", exportID)),
    ).
    WithAutoDismiss(0). // sticky until dismissed
    WithOOB("#toast-container", "afterbegin")

// tavern: deliver only to this user's scoped subscription
//   tavern.PublishOOBTo("notifications", userID, tavern.Replace("#toast-container", rendered))
```

### Per-resource modal trigger

```go
// Admin locks a user account — notify viewers of that user's profile
modal := linkwell.ModalConfig{
    ID:      "account-locked-modal",
    Title:   "Account Locked",
    Buttons: linkwell.ModalOK,
}

// tavern: publish to anyone viewing this resource
//   tavern.PublishOOBTo("user:42", "viewers", tavern.Replace("#modal-container", rendered))
```

### Error broadcast to a team channel

```go
ec := linkwell.ErrorContext{
    Message:    "Payment gateway timeout",
    StatusCode: 503,
    Controls:   linkwell.ServiceErrorControls(linkwell.ErrorControlOpts{
        RetryURL:    "/payments/retry",
        RetryTarget: "#payment-status",
    }),
    Closable: true,
}.WithOOB("#error-status", "innerHTML")

// tavern: broadcast to all ops team members
//   tavern.PublishOOBTo("ops-alerts", "team:ops", tavern.Replace("#error-status", rendered))
```

---

## Lifecycle-aware publishing

tavern starts/stops expensive work based on subscriber presence. linkwell controls adapt to whether live data is flowing.

### Dashboard with on-demand metrics

```go
// tavern lifecycle hooks (registered at startup):
//   broker.OnFirstSubscriber("metrics", func() {
//       // start polling the metrics API
//   })
//   broker.OnLastUnsubscribe("metrics", func() {
//       // stop polling — no one is watching
//   })

// linkwell: render controls based on whether the SSE connection is active.
// The initial page render includes the static structure:
tabs := linkwell.NewTabConfig("metrics-tabs", "#metrics-content",
    linkwell.TabItem{Label: "CPU", Href: "/metrics/cpu"},
    linkwell.TabItem{Label: "Memory", Href: "/metrics/memory"},
    linkwell.TabItem{Label: "Disk", Href: "/metrics/disk"},
)

// When tavern detects no subscribers, publish a "paused" state:
//   tavern.PublishOOB("metrics", tavern.Replace("#metrics-content", pausedHTML))
//
// The paused view uses a linkwell control to reconnect:
reconnect := linkwell.HTMXAction("Resume Live Data", linkwell.HxGet("/metrics/cpu", "#metrics-content"))
```

---

## Multi-step form with live progress

linkwell's `StepperConfig` tracks wizard state. tavern can push step completion from background processing.

### Stepper setup

```go
stepper := linkwell.NewStepper(1, // currently on step 2 (0-indexed)
    linkwell.Step{Label: "Upload", Href: "/import/upload", Icon: "cloud-upload"},
    linkwell.Step{Label: "Validate", Href: "/import/validate", Icon: "check-circle"},
    linkwell.Step{Label: "Map Fields", Href: "/import/mapping", Icon: "arrows-right-left"},
    linkwell.Step{Label: "Import", Href: "/import/run", Icon: "play"},
)
// Steps[0] is Complete, Steps[1] is Active, Steps[2-3] are Pending
// stepper.Prev points to /import/upload, stepper.Next points to /import/mapping
```

### Async validation with tavern

```go
// After the user submits the file, the server kicks off async validation.
// tavern pushes progress updates scoped to this import session:
//
//   tavern.PublishOOBTo("import", sessionID,
//       tavern.Replace("#validation-progress", renderedProgress))
//
// When validation completes, tavern pushes the updated stepper:
//   stepper.Steps[1].Status = linkwell.StepComplete
//   // re-render stepper and push via OOB
//   tavern.PublishOOBTo("import", sessionID,
//       tavern.Replace("#stepper", renderedStepper))
```

---

## Division of labor

| Concern | Library |
|---|---|
| Navigation, breadcrumbs, tabs | linkwell |
| Action controls (CRUD, bulk, row, form) | linkwell |
| Filters, sortable columns, pagination | linkwell |
| Error context and controls | linkwell |
| Toast notifications | linkwell |
| Modals | linkwell |
| Stepper / wizard state | linkwell |
| Link graph validation | linkwell |
| Sitemap from link registry | linkwell |
| Real-time push via SSE | [tavern](https://github.com/catgoose/tavern) |
| OOB DOM mutations (replace, append, delete) | [tavern](https://github.com/catgoose/tavern) |
| Scoped per-user/per-resource streams | [tavern](https://github.com/catgoose/tavern) |
| Subscriber-aware publish gating | [tavern](https://github.com/catgoose/tavern) |
| Scheduled / debounced / throttled publishing | [tavern](https://github.com/catgoose/tavern) |
