package linkwell_test

import (
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/catgoose/linkwell"
)

// --- ComputeTotalPages ---

func TestComputeTotalPages_ExactDivision(t *testing.T) {
	if got := linkwell.ComputeTotalPages(20, 10); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
}

func TestComputeTotalPages_WithRemainder(t *testing.T) {
	if got := linkwell.ComputeTotalPages(21, 10); got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
}

func TestComputeTotalPages_ZeroItems(t *testing.T) {
	if got := linkwell.ComputeTotalPages(0, 10); got != 1 {
		t.Errorf("expected minimum 1, got %d", got)
	}
}

func TestComputeTotalPages_ZeroPerPage(t *testing.T) {
	if got := linkwell.ComputeTotalPages(100, 0); got != 1 {
		t.Errorf("expected minimum 1 for zero perPage, got %d", got)
	}
}

// --- PageInfo.URLForPage ---

func TestURLForPage_CleanURL(t *testing.T) {
	info := linkwell.PageInfo{BaseURL: "/users", Target: "#tc"}
	u := info.URLForPage(2)
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("invalid URL: %v", err)
	}
	if parsed.Query().Get(linkwell.ParamPage) != "2" {
		t.Errorf("expected page=2, got %q", parsed.Query().Get(linkwell.ParamPage))
	}
}

func TestURLForPage_PreservesExistingParams(t *testing.T) {
	info := linkwell.PageInfo{BaseURL: "/users?q=foo&role=admin", Target: "#tc"}
	u := info.URLForPage(3)
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("invalid URL: %v", err)
	}
	q := parsed.Query()
	if q.Get("q") != "foo" {
		t.Errorf("expected q=foo, got %q", q.Get("q"))
	}
	if q.Get("role") != "admin" {
		t.Errorf("expected role=admin, got %q", q.Get("role"))
	}
	if q.Get(linkwell.ParamPage) != "3" {
		t.Errorf("expected page=3, got %q", q.Get(linkwell.ParamPage))
	}
}

func TestURLForPage_CustomPageParam(t *testing.T) {
	info := linkwell.PageInfo{BaseURL: "/items", PageParam: "p", Target: "#tc"}
	u := info.URLForPage(5)
	parsed, _ := url.Parse(u)
	if parsed.Query().Get("p") != "5" {
		t.Errorf("expected custom param p=5, got %q", parsed.Query().Get("p"))
	}
	if parsed.Query().Get(linkwell.ParamPage) != "" {
		t.Errorf("expected no default 'page' param when custom is used")
	}
}

func TestURLForPage_Page1(t *testing.T) {
	info := linkwell.PageInfo{BaseURL: "/users", Target: "#tc"}
	u := info.URLForPage(1)
	parsed, _ := url.Parse(u)
	if parsed.Query().Get(linkwell.ParamPage) != "1" {
		t.Errorf("expected page=1, got %q", parsed.Query().Get(linkwell.ParamPage))
	}
}

// --- SortableCol ---

func TestSortableCol_NonMatchingKey(t *testing.T) {
	col := linkwell.SortableCol("name", "Name", "email", "asc", "/users", "#tc", "#filter-form")
	if col.SortDir != linkwell.SortNone {
		t.Errorf("expected SortNone for non-matching key, got %q", col.SortDir)
	}
	parsed, err := url.Parse(col.SortURL)
	if err != nil {
		t.Fatalf("invalid SortURL: %v", err)
	}
	q := parsed.Query()
	if q.Get(linkwell.ParamSort) != "name" {
		t.Errorf("expected sort=name, got %q", q.Get(linkwell.ParamSort))
	}
	if q.Get(linkwell.ParamDir) != "asc" {
		t.Errorf("expected dir=asc for non-matching key, got %q", q.Get(linkwell.ParamDir))
	}
}

func TestSortableCol_MatchingKeyAsc(t *testing.T) {
	col := linkwell.SortableCol("name", "Name", "name", "asc", "/users", "#tc", "#filter-form")
	if col.SortDir != linkwell.SortAsc {
		t.Errorf("expected SortAsc, got %q", col.SortDir)
	}
	parsed, _ := url.Parse(col.SortURL)
	if parsed.Query().Get("dir") != "desc" {
		t.Errorf("expected dir=desc toggle from asc, got %q", parsed.Query().Get("dir"))
	}
}

func TestSortableCol_MatchingKeyDesc(t *testing.T) {
	col := linkwell.SortableCol("name", "Name", "name", "desc", "/users", "#tc", "#filter-form")
	if col.SortDir != linkwell.SortDesc {
		t.Errorf("expected SortDesc, got %q", col.SortDir)
	}
	parsed, _ := url.Parse(col.SortURL)
	if parsed.Query().Get("dir") != "asc" {
		t.Errorf("expected dir=asc toggle from desc, got %q", parsed.Query().Get("dir"))
	}
}

func TestSortableCol_IncludeSet(t *testing.T) {
	col := linkwell.SortableCol("name", "Name", "name", "asc", "/users", "#tc", "#filter-form")
	if col.Include != "#filter-form" {
		t.Errorf("expected Include=#filter-form, got %q", col.Include)
	}
}

func TestSortableCol_SortURLIsValidURL(t *testing.T) {
	col := linkwell.SortableCol("email", "Email", "", "", "/users?q=foo", "#tc", "")
	if _, err := url.Parse(col.SortURL); err != nil {
		t.Errorf("SortURL is not a valid URL: %v", err)
	}
}

func TestSortableCol_Sortable(t *testing.T) {
	col := linkwell.SortableCol("name", "Name", "", "", "/users", "#tc", "")
	if !col.Sortable {
		t.Error("SortableCol should set Sortable=true")
	}
}

// --- PaginationControls ---

func TestPaginationControls_NilWhenTotalPagesOne(t *testing.T) {
	info := linkwell.PageInfo{Page: 1, TotalPages: 1, BaseURL: "/users", Target: "#tc"}
	if controls := linkwell.PaginationControls(info); controls != nil {
		t.Errorf("expected nil for TotalPages=1, got %v", controls)
	}
}

func TestPaginationControls_NilWhenTotalPagesZero(t *testing.T) {
	info := linkwell.PageInfo{Page: 1, TotalPages: 0, BaseURL: "/users", Target: "#tc"}
	if controls := linkwell.PaginationControls(info); controls != nil {
		t.Errorf("expected nil for TotalPages=0, got %v", controls)
	}
}

func TestPaginationControls_CurrentPageIsDisabledPrimary(t *testing.T) {
	info := linkwell.PageInfo{Page: 3, TotalPages: 10, BaseURL: "/users", Target: "#tc"}
	controls := linkwell.PaginationControls(info)
	if controls == nil {
		t.Fatal("expected controls, got nil")
	}
	found := false
	for _, ctrl := range controls {
		if ctrl.Label == "3" && ctrl.Disabled && ctrl.Variant == linkwell.VariantPrimary {
			found = true
			break
		}
	}
	if !found {
		t.Error("current page (3) should be Disabled+VariantPrimary")
	}
}

func TestPaginationControls_Page1_FirstAndPrevDisabled(t *testing.T) {
	info := linkwell.PageInfo{Page: 1, TotalPages: 5, BaseURL: "/users", Target: "#tc"}
	controls := linkwell.PaginationControls(info)
	// First two controls should be « and ‹, both disabled.
	if len(controls) < 2 {
		t.Fatalf("expected at least 2 controls, got %d", len(controls))
	}
	if controls[0].Label != linkwell.PaginationFirst || !controls[0].Disabled {
		t.Errorf("first control should be disabled «, got label=%q disabled=%v", controls[0].Label, controls[0].Disabled)
	}
	if controls[1].Label != linkwell.PaginationPrev || !controls[1].Disabled {
		t.Errorf("second control should be disabled ‹, got label=%q disabled=%v", controls[1].Label, controls[1].Disabled)
	}
}

func TestPaginationControls_LastPage_NextAndLastDisabled(t *testing.T) {
	info := linkwell.PageInfo{Page: 5, TotalPages: 5, BaseURL: "/users", Target: "#tc"}
	controls := linkwell.PaginationControls(info)
	last := controls[len(controls)-1]
	prev := controls[len(controls)-2]
	if last.Label != linkwell.PaginationLast || !last.Disabled {
		t.Errorf("last control should be disabled », got label=%q disabled=%v", last.Label, last.Disabled)
	}
	if prev.Label != linkwell.PaginationNext || !prev.Disabled {
		t.Errorf("second-to-last should be disabled ›, got label=%q disabled=%v", prev.Label, prev.Disabled)
	}
}

func TestPaginationControls_WindowClipsAtBoundaries(t *testing.T) {
	// Page 1 of 3: window 1..3 (capped; page+2=3)
	info := linkwell.PageInfo{Page: 1, TotalPages: 3, BaseURL: "/users", Target: "#tc"}
	controls := linkwell.PaginationControls(info)
	// Find all numeric page labels.
	var pages []string
	for _, ctrl := range controls {
		if _, err := strconv.Atoi(ctrl.Label); err == nil {
			pages = append(pages, ctrl.Label)
		}
	}
	// Should contain 1, 2, 3 but not 0 or negative.
	for _, p := range pages {
		n, _ := strconv.Atoi(p)
		if n < 1 || n > 3 {
			t.Errorf("page %d out of range [1,3]", n)
		}
	}
}

func TestPaginationControls_NormalPagesHaveHxRequest(t *testing.T) {
	info := linkwell.PageInfo{Page: 3, TotalPages: 10, BaseURL: "/users", Target: "#tc"}
	controls := linkwell.PaginationControls(info)
	for _, ctrl := range controls {
		if ctrl.Disabled {
			continue
		}
		if ctrl.HxRequest.URL == "" {
			t.Errorf("non-disabled control %q should have HxRequest.URL, got none", ctrl.Label)
		}
		if ctrl.HxRequest.Target != "#tc" {
			t.Errorf("non-disabled control %q should have target=#tc, got %q", ctrl.Label, ctrl.HxRequest.Target)
		}
	}
}

func TestPaginationControls_IncludePropagated(t *testing.T) {
	info := linkwell.PageInfo{
		Page: 3, TotalPages: 10,
		BaseURL: "/users", Target: "#tc", Include: "#filter-form",
	}
	controls := linkwell.PaginationControls(info)
	for _, ctrl := range controls {
		if ctrl.Disabled {
			continue
		}
		if ctrl.HxRequest.Include != "#filter-form" {
			t.Errorf("non-disabled control %q should have include=#filter-form, got %q",
				ctrl.Label, ctrl.HxRequest.Include)
		}
	}
}

func TestPaginationControls_NoIncludeWhenEmpty(t *testing.T) {
	info := linkwell.PageInfo{Page: 2, TotalPages: 5, BaseURL: "/users", Target: "#tc"}
	controls := linkwell.PaginationControls(info)
	for _, ctrl := range controls {
		if ctrl.Disabled {
			continue
		}
		if ctrl.HxRequest.Include != "" {
			t.Errorf("control %q should not have include when PageInfo.Include is empty", ctrl.Label)
		}
	}
}

func TestPaginationControls_GetURLContainsPageNumber(t *testing.T) {
	info := linkwell.PageInfo{Page: 3, TotalPages: 10, BaseURL: "/users", Target: "#tc"}
	controls := linkwell.PaginationControls(info)
	// Find the control for page 4 (should be in the window).
	for _, ctrl := range controls {
		if ctrl.Label == "4" && !ctrl.Disabled {
			if !strings.Contains(ctrl.HxRequest.URL, "page=4") {
				t.Errorf("page 4 control URL should contain page=4, got %q", ctrl.HxRequest.URL)
			}
			return
		}
	}
	t.Error("did not find a non-disabled control for page 4")
}
