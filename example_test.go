package linkwell_test

import (
	"fmt"

	"github.com/catgoose/linkwell"
)

func ExampleRel() {
	entry := linkwell.Rel("/inventory", "Inventory")
	fmt.Println(entry.Path, entry.Title)
	// Output: /inventory Inventory
}

func ExampleTitleFromPath() {
	fmt.Println(linkwell.TitleFromPath("/demo/inventory"))
	fmt.Println(linkwell.TitleFromPath("/admin/error-traces"))
	// Output:
	// Inventory
	// Error Traces
}

func ExampleLink() {
	linkwell.ResetForTesting()

	// rel="related" is symmetric: both sides are registered automatically.
	linkwell.Link("/inventory", "related", "/warehouses", "Warehouses")

	fmt.Println(linkwell.LinksFor("/inventory")[0].Title)
	fmt.Println(linkwell.LinksFor("/warehouses")[0].Title)
	// Output:
	// Warehouses
	// Inventory
}

func ExampleRing() {
	linkwell.ResetForTesting()

	linkwell.Ring("Logistics",
		linkwell.Rel("/inventory", "Inventory"),
		linkwell.Rel("/warehouses", "Warehouses"),
		linkwell.Rel("/shipments", "Shipments"),
	)

	links := linkwell.LinksFor("/inventory")
	for _, l := range links {
		fmt.Printf("%s -> %s (group=%s)\n", l.Rel, l.Title, l.Group)
	}
	// Output:
	// related -> Warehouses (group=Logistics)
	// related -> Shipments (group=Logistics)
}

func ExampleHub() {
	linkwell.ResetForTesting()

	linkwell.Hub("/admin", "Admin",
		linkwell.Rel("/admin/users", "Users"),
		linkwell.Rel("/admin/roles", "Roles"),
	)

	// Center links to spokes via rel="related"
	centerLinks := linkwell.LinksFor("/admin", "related")
	fmt.Printf("center has %d spokes\n", len(centerLinks))

	// Spokes link back to center via rel="up"
	upLinks := linkwell.LinksFor("/admin/users", "up")
	fmt.Printf("spoke -> %s (%s)\n", upLinks[0].Href, upLinks[0].Title)
	// Output:
	// center has 2 spokes
	// spoke -> /admin (Admin)
}

func ExampleHubs() {
	linkwell.ResetForTesting()

	linkwell.Hub("/admin", "Admin",
		linkwell.Rel("/admin/users", "Users"),
		linkwell.Rel("/admin/roles", "Roles"),
	)

	for _, hub := range linkwell.Hubs() {
		fmt.Printf("%s (%s)\n", hub.Title, hub.Path)
		for _, spoke := range hub.Spokes {
			fmt.Printf("  %s %s\n", spoke.Title, spoke.Href)
		}
	}
	// Output:
	// Admin (/admin)
	//   Roles /admin/roles
	//   Users /admin/users
}

func ExampleLinksFor() {
	linkwell.ResetForTesting()

	linkwell.Link("/page", "related", "/other", "Other")
	linkwell.Link("/page", "up", "/parent", "Parent")

	// All links for a path
	all := linkwell.LinksFor("/page")
	fmt.Printf("all: %d\n", len(all))

	// Filter by rel type
	upOnly := linkwell.LinksFor("/page", "up")
	fmt.Printf("up: %s\n", upOnly[0].Href)
	// Output:
	// all: 2
	// up: /parent
}

func ExampleRelatedLinksFor() {
	linkwell.ResetForTesting()

	linkwell.Ring("group",
		linkwell.Rel("/a", "A"),
		linkwell.Rel("/b", "B"),
		linkwell.Rel("/c", "C"),
	)

	peers := linkwell.RelatedLinksFor("/a")
	for _, l := range peers {
		fmt.Println(l.Title)
	}
	// Output:
	// B
	// C
}

func ExampleLinkHeader() {
	links := []linkwell.LinkRelation{
		{Rel: "related", Href: "/warehouses", Title: "Warehouses"},
		{Rel: "up", Href: "/admin", Title: "Admin"},
	}
	fmt.Println(linkwell.LinkHeader(links))
	// Output: </warehouses>; rel="related"; title="Warehouses", </admin>; rel="up"; title="Admin"
}

func ExampleLoadStoredLink() {
	linkwell.ResetForTesting()

	linkwell.LoadStoredLink("/projects/42", linkwell.LinkRelation{
		Rel: "related", Href: "/teams/7", Title: "Backend Team",
	})

	links := linkwell.LinksFor("/projects/42")
	fmt.Println(links[0].Title)
	// Output: Backend Team
}

func ExampleRemoveLink() {
	linkwell.ResetForTesting()

	linkwell.LoadStoredLink("/a", linkwell.LinkRelation{
		Rel: "related", Href: "/b", Title: "B",
	})

	ok := linkwell.RemoveLink("/a", "/b", "related")
	fmt.Println(ok)
	fmt.Println(len(linkwell.LinksFor("/a")))
	// Output:
	// true
	// 0
}

func ExampleBreadcrumbsFromLinks() {
	linkwell.ResetForTesting()

	linkwell.Hub("/admin", "Admin",
		linkwell.Rel("/admin/users", "Users"),
	)

	crumbs := linkwell.BreadcrumbsFromLinks("/admin/users")
	for _, c := range crumbs {
		if c.Href == "" {
			fmt.Printf("[%s]\n", c.Label)
		} else {
			fmt.Printf("%s (%s)\n", c.Label, c.Href)
		}
	}
	// Output:
	// Home (/)
	// Admin (/admin)
	// [Users]
}

func ExampleBreadcrumbsFromPath() {
	crumbs := linkwell.BreadcrumbsFromPath("/users/42/edit", map[int]string{
		1: "Jane Doe",
	})
	for _, c := range crumbs {
		if c.Href == "" {
			fmt.Printf("[%s]\n", c.Label)
		} else {
			fmt.Printf("%s (%s)\n", c.Label, c.Href)
		}
	}
	// Output:
	// Home (/)
	// Users (/users)
	// Jane Doe (/users/42)
	// [Edit]
}

func ExampleRegisterFrom() {
	linkwell.RegisterFrom(linkwell.FromDashboard, linkwell.Breadcrumb{
		Label: "Dashboard", Href: "/dashboard",
	})

	crumbs := linkwell.ResolveFromMask(linkwell.FromHome | linkwell.FromDashboard)
	for _, c := range crumbs {
		fmt.Println(c.Label)
	}
	// Output:
	// Home
	// Dashboard
}

func ExampleParseFromParam() {
	mask := linkwell.ParseFromParam("3")
	fmt.Println(mask)
	fmt.Println(linkwell.ParseFromParam(""))
	// Output:
	// 3
	// 0
}

func ExampleFromNav() {
	fmt.Println(linkwell.FromNav("/users/42", "3"))
	fmt.Println(linkwell.FromNav("/users?q=foo", "3"))
	fmt.Println(linkwell.FromNav("/users/42", ""))
	// Output:
	// /users/42?from=3
	// /users?q=foo&from=3
	// /users/42
}

func ExampleFromQueryString() {
	fmt.Println(linkwell.FromQueryString(3))
	fmt.Println(linkwell.FromQueryString(0))
	// Output:
	// from=3
	//
}

func ExampleSetActiveNavItem() {
	items := []linkwell.NavItem{
		{Label: "Dashboard", Href: "/dashboard"},
		{Label: "Users", Href: "/users"},
	}
	result := linkwell.SetActiveNavItem(items, "/users")
	for _, item := range result {
		fmt.Printf("%s active=%v\n", item.Label, item.Active)
	}
	// Output:
	// Dashboard active=false
	// Users active=true
}

func ExampleSetActiveNavItemPrefix() {
	items := []linkwell.NavItem{
		{Label: "Users", Href: "/users"},
		{Label: "Settings", Href: "/settings"},
	}
	// /users is active when the path is /users/42/edit
	result := linkwell.SetActiveNavItemPrefix(items, "/users/42/edit")
	for _, item := range result {
		fmt.Printf("%s active=%v\n", item.Label, item.Active)
	}
	// Output:
	// Users active=true
	// Settings active=false
}

func ExampleNavItemFromControl() {
	ctrl := linkwell.Control{
		Label:     "Dashboard",
		Href:      "/dashboard",
		Icon:      linkwell.IconHome,
		HxRequest: linkwell.HxGet("/dashboard", "body"),
	}
	nav := linkwell.NavItemFromControl(ctrl)
	fmt.Printf("%s %s %s\n", nav.Label, nav.Href, nav.Icon)
	// Output: Dashboard /dashboard home
}

func ExampleRetryButton() {
	ctrl := linkwell.RetryButton("Retry", linkwell.HxMethodGet, "/api/data", "#content")
	fmt.Printf("kind=%s variant=%s\n", ctrl.Kind, ctrl.Variant)
	// Output: kind=retry variant=primary
}

func ExampleConfirmAction() {
	ctrl := linkwell.ConfirmAction("Delete", linkwell.HxMethodDelete, "/users/42", "#list", "Delete?")
	fmt.Printf("kind=%s variant=%s confirm=%s\n", ctrl.Kind, ctrl.Variant, ctrl.Confirm)
	// Output: kind=htmx variant=danger confirm=Delete?
}

func ExampleBackButton() {
	ctrl := linkwell.BackButton("Go Back")
	fmt.Printf("kind=%s label=%s\n", ctrl.Kind, ctrl.Label)
	// Output: kind=back label=Go Back
}

func ExampleGoHomeButton() {
	ctrl := linkwell.GoHomeButton("Go Home", "/", "body")
	fmt.Printf("kind=%s href=%s pushURL=%s\n", ctrl.Kind, ctrl.Href, ctrl.PushURL)
	// Output: kind=home href=/ pushURL=/
}

func ExampleRedirectLink() {
	ctrl := linkwell.RedirectLink("View Profile", "/users/42")
	fmt.Printf("kind=%s href=%s\n", ctrl.Kind, ctrl.Href)
	// Output: kind=link href=/users/42
}

func ExampleHTMXAction() {
	ctrl := linkwell.HTMXAction("Archive", linkwell.HxPost("/users/42/archive", "#content"))
	fmt.Printf("kind=%s method=%s url=%s\n", ctrl.Kind, ctrl.HxRequest.Method, ctrl.HxRequest.URL)
	// Output: kind=htmx method=post url=/users/42/archive
}

func ExampleDismissButton() {
	ctrl := linkwell.DismissButton("Close")
	fmt.Printf("kind=%s label=%s\n", ctrl.Kind, ctrl.Label)
	// Output: kind=dismiss label=Close
}

func ExampleControl_WithSwap() {
	ctrl := linkwell.RetryButton("Retry", linkwell.HxMethodGet, "/api", "#c").
		WithSwap(linkwell.SwapOuterHTML).
		WithVariant(linkwell.VariantDanger).
		WithIcon(linkwell.IconCheck).
		WithConfirm("Sure?").
		WithDisabled(true).
		WithErrorTarget("#err")
	fmt.Printf("swap=%s variant=%s icon=%s confirm=%s disabled=%v errTarget=%s\n",
		ctrl.Swap, ctrl.Variant, ctrl.Icon, ctrl.Confirm, ctrl.Disabled, ctrl.ErrorTarget)
	// Output: swap=outerHTML variant=danger icon=check confirm=Sure? disabled=true errTarget=#err
}

func ExampleHxGet() {
	req := linkwell.HxGet("/users", "#user-list")
	fmt.Printf("method=%s url=%s target=%s\n", req.Method, req.URL, req.Target)
	// Output: method=get url=/users target=#user-list
}

func ExampleHxRequestConfig_Attrs() {
	req := linkwell.HxPost("/users", "#list").WithInclude("closest form")
	attrs := req.Attrs()
	fmt.Println(attrs["post"])
	fmt.Println(attrs["target"])
	fmt.Println(attrs["include"])
	// Output:
	// /users
	// #list
	// closest form
}

func ExampleNewFilterBar() {
	bar := linkwell.NewFilterBar("/users", "#user-table",
		linkwell.SearchField("q", "Search users...", "alice"),
		linkwell.SelectField("status", "Status", "active", linkwell.SelectOptions("active",
			"", "All",
			"active", "Active",
			"inactive", "Inactive",
		)),
	)
	fmt.Printf("id=%s action=%s fields=%d\n", bar.ID, bar.Action, len(bar.Fields))
	fmt.Printf("search value=%s\n", bar.Fields[0].Value)
	// Output:
	// id=filter-form action=/users fields=2
	// search value=alice
}

func ExampleSelectOptions() {
	opts := linkwell.SelectOptions("published",
		"draft", "Draft",
		"published", "Published",
		"archived", "Archived",
	)
	for _, o := range opts {
		fmt.Printf("%s selected=%v\n", o.Label, o.Selected)
	}
	// Output:
	// Draft selected=false
	// Published selected=true
	// Archived selected=false
}

func ExampleNewFilterGroup() {
	group := linkwell.NewFilterGroup("/products", "#table",
		linkwell.SelectField("cat", "Category", "", nil),
	)
	group.UpdateOptions("cat", linkwell.SelectOptions("",
		"electronics", "Electronics",
		"clothing", "Clothing",
	))
	selects := group.SelectFields()
	fmt.Printf("select fields: %d\n", len(selects))
	fmt.Printf("options: %d\n", len(selects[0].Options))
	// Output:
	// select fields: 1
	// options: 2
}

func ExampleSortableCol() {
	col := linkwell.SortableCol("name", "Name", "name", "asc", "/users", "#table", "#filter-form")
	fmt.Printf("key=%s dir=%s sortable=%v\n", col.Key, col.SortDir, col.Sortable)
	// Output: key=name dir=asc sortable=true
}

func ExampleComputeTotalPages() {
	fmt.Println(linkwell.ComputeTotalPages(100, 25))
	fmt.Println(linkwell.ComputeTotalPages(101, 25))
	fmt.Println(linkwell.ComputeTotalPages(0, 25))
	// Output:
	// 4
	// 5
	// 1
}

func ExamplePaginationControls() {
	info := linkwell.PageInfo{
		BaseURL:    "/users",
		Page:       2,
		PerPage:    25,
		TotalItems: 100,
		TotalPages: 4,
		Target:     "#user-table",
	}
	controls := linkwell.PaginationControls(info)
	for _, c := range controls {
		fmt.Printf("label=%-3s disabled=%v\n", c.Label, c.Disabled)
	}
	// Output:
	// label=«   disabled=false
	// label=‹   disabled=false
	// label=1   disabled=false
	// label=2   disabled=true
	// label=3   disabled=false
	// label=4   disabled=false
	// label=›   disabled=false
	// label=»   disabled=false
}

func ExamplePageInfo_URLForPage() {
	info := linkwell.PageInfo{BaseURL: "/users?q=foo"}
	fmt.Println(info.URLForPage(3))
	// Output: /users?page=3&q=foo
}

func ExampleReportIssueModal() {
	modal := linkwell.ReportIssueModal("req-abc")
	fmt.Printf("id=%s title=%s post=%s swap=%s\n", modal.ID, modal.Title, modal.HxPost, modal.HxSwap)
	// Output: id=report-issue-modal title=Report Issue post=/report-issue/req-abc swap=none
}

func ExampleResourceActions() {
	controls := linkwell.ResourceActions(linkwell.ResourceActionCfg{
		EditURL:    "/users/42/edit",
		DeleteURL:  "/users/42",
		ConfirmMsg: "Delete this user?",
		Target:     "#content",
	})
	for _, c := range controls {
		fmt.Printf("%s kind=%s variant=%s\n", c.Label, c.Kind, c.Variant)
	}
	// Output:
	// Edit kind=htmx variant=secondary
	// Delete kind=htmx variant=danger
}

func ExampleRowActions() {
	controls := linkwell.RowActions(linkwell.RowActionCfg{
		EditURL:    "/users/42/edit",
		DeleteURL:  "/users/42",
		RowTarget:  "#row-42",
		ConfirmMsg: "Delete?",
	})
	for _, c := range controls {
		fmt.Printf("%s swap=%s\n", c.Label, c.Swap)
	}
	// Output:
	// Edit swap=outerHTML
	// Delete swap=outerHTML
}

func ExampleFormActions() {
	controls := linkwell.FormActions("/users")
	for _, c := range controls {
		fmt.Printf("%s kind=%s\n", c.Label, c.Kind)
	}
	// Output:
	// Save kind=htmx
	// Cancel kind=link
}

func ExampleBulkActions() {
	controls := linkwell.BulkActions(linkwell.BulkActionCfg{
		DeleteURL:        "/users/bulk-delete",
		ActivateURL:      "/users/bulk-activate",
		TableTarget:      "#user-table",
		CheckboxSelector: ".user-checkbox",
	})
	for _, c := range controls {
		fmt.Printf("%s variant=%s\n", c.Label, c.Variant)
	}
	// Output:
	// Delete Selected variant=danger
	// Activate variant=secondary
}

func ExampleEmptyStateAction() {
	ctrl := linkwell.EmptyStateAction("Create First User", "/users/new", "#content")
	fmt.Printf("%s kind=%s variant=%s\n", ctrl.Label, ctrl.Kind, ctrl.Variant)
	// Output: Create First User kind=htmx variant=primary
}

func ExampleErrorControlsForStatus() {
	controls := linkwell.ErrorControlsForStatus(404, linkwell.ErrorControlOpts{
		HomeURL: "/",
	})
	for _, c := range controls {
		fmt.Printf("%s kind=%s\n", c.Label, c.Kind)
	}
	// Output:
	// Go Back kind=back
	// Go Home kind=home
}

func ExampleErrorContext_WithControls() {
	ec := linkwell.ErrorContext{
		StatusCode: 500,
		Message:    "Database error",
	}
	ec = ec.WithControls(linkwell.DismissButton("Close"))
	ec = ec.WithOOB("#error-status", "innerHTML")
	fmt.Printf("controls=%d oobTarget=%s\n", len(ec.Controls), ec.OOBTarget)
	// Output: controls=1 oobTarget=#error-status
}

func ExampleNewHTTPError() {
	ec := linkwell.ErrorContext{
		StatusCode: 404,
		Message:    "User not found",
	}
	err := linkwell.NewHTTPError(ec)
	fmt.Println(err.Error())
	// Output: HTTP 404: User not found
}
