package linkwell

import (
	"net/url"
	"strconv"
)

// SortDir indicates the current sort direction for a table column. The zero
// value (SortNone) means the column is not currently sorted.
type SortDir string

// Sort direction constants for TableCol.SortDir.
const (
	SortNone SortDir = ""
	SortAsc  SortDir = "asc"
	SortDesc SortDir = "desc"
)

// Query parameter keys used by sort and pagination helpers.
const (
	ParamSort = "sort"
	ParamDir  = "dir"
	ParamPage = "page"
)

// Pagination navigation labels.
const (
	PaginationFirst = "«"
	PaginationPrev  = "‹"
	PaginationNext  = "›"
	PaginationLast  = "»"
)

// TableCol is a pure-data column descriptor for a table header. Sort metadata
// (direction, toggle URL) is precomputed by the handler so templates render
// without logic. Non-sortable columns leave SortURL empty.
type TableCol struct {
	// Key identifies the column (maps to the "sort" query parameter value).
	Key string
	// Label is the visible column header text.
	Label string
	// SortDir is the current sort direction for this column.
	SortDir SortDir
	// SortURL is the HTMX request URL that toggles the sort direction. Built by
	// SortableCol with the next direction pre-encoded.
	SortURL string
	// Target is the CSS selector for hx-target on the sort link.
	Target string
	// Include is the CSS selector for hx-include on the sort link (e.g.,
	// "#filter-form" to forward filter state).
	Include string
	// Width is an optional CSS width hint (e.g., "120px", "20%").
	Width string
	// Sortable indicates whether the column header renders as a clickable sort link.
	Sortable bool
}

// PageInfo carries server-computed pagination state used by PaginationControls
// to generate the page navigation bar. All fields are set by the handler; the
// template just passes PageInfo through.
type PageInfo struct {
	// BaseURL is the current URL (with existing query params preserved).
	// URLForPage appends/replaces the page parameter.
	BaseURL string
	// PageParam overrides the query parameter name for the page number.
	// Defaults to "page" when empty.
	PageParam string
	// Target is the CSS selector for hx-target on pagination links.
	Target string
	// Include is the CSS selector for hx-include (e.g., "#filter-form").
	Include string
	// Page is the current 1-based page number.
	Page int
	// PerPage is the number of items per page.
	PerPage int
	// TotalItems is the total number of items across all pages.
	TotalItems int
	// TotalPages is the total number of pages. Use ComputeTotalPages to calculate.
	TotalPages int
}

// pageParam returns the PageParam field, defaulting to "page" when empty.
func (pi PageInfo) pageParam() string {
	if pi.PageParam == "" {
		return ParamPage
	}
	return pi.PageParam
}

// URLForPage builds the URL for a specific page number by appending or
// replacing the page query parameter in BaseURL.
func (pi PageInfo) URLForPage(page int) string {
	u, err := url.Parse(pi.BaseURL)
	if err != nil {
		return pi.BaseURL
	}
	q := u.Query()
	q.Set(pi.pageParam(), strconv.Itoa(page))
	u.RawQuery = q.Encode()
	return u.String()
}

// ComputeTotalPages calculates the number of pages needed to display
// totalItems at the given perPage size. Always returns at least 1.
func ComputeTotalPages(totalItems, perPage int) int {
	if perPage <= 0 {
		return 1
	}
	if totalItems <= 0 {
		return 1
	}
	pages := (totalItems + perPage - 1) / perPage
	if pages < 1 {
		return 1
	}
	return pages
}

// SortableCol creates a TableCol with precomputed sort state and toggle URL.
// Pass the current sort key and direction from the request's query parameters.
// The baseURL should have "sort" and "dir" params already stripped. Toggle
// logic: clicking an unsorted column sorts ascending, clicking the active
// ascending column toggles to descending, and clicking the active descending
// column toggles back to ascending.
func SortableCol(key, label, currentSortKey, currentSortDir, baseURL, target, include string) TableCol {
	col := TableCol{
		Key:      key,
		Label:    label,
		Sortable: true,
		Target:   target,
		Include:  include,
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		u, _ = url.Parse(baseURL)
		if u == nil {
			u = &url.URL{}
		}
	}
	q := u.Query()

	if key != currentSortKey {
		// Not the active sort column: show as unsorted, clicking will sort asc.
		col.SortDir = SortNone
		q.Set(ParamSort, key)
		q.Set(ParamDir, string(SortAsc))
	} else {
		switch SortDir(currentSortDir) {
		case SortAsc:
			col.SortDir = SortAsc
			q.Set(ParamSort, key)
			q.Set(ParamDir, string(SortDesc))
		default:
			// SortDesc or any unexpected value → show desc, toggle back to asc.
			col.SortDir = SortDesc
			q.Set(ParamSort, key)
			q.Set(ParamDir, string(SortAsc))
		}
	}

	u.RawQuery = q.Encode()
	col.SortURL = u.String()
	return col
}

// PaginationControls generates the control slice for a pagination bar. Returns
// nil if TotalPages <= 1 (no pagination needed). The returned controls follow
// a [First][Prev][page window][Next][Last] layout. The current page is rendered
// as a disabled primary-variant control; boundary controls (First/Prev at page
// 1, Next/Last at last page) are disabled.
func PaginationControls(info PageInfo) []Control {
	if info.TotalPages <= 1 {
		return nil
	}

	buildReq := func(pageURL string) HxRequestConfig {
		return HxRequestConfig{
			Method:  HxMethodGet,
			URL:     pageURL,
			Target:  info.Target,
			Include: info.Include,
		}
	}

	disabledCtrl := func(label string) Control {
		return Control{
			Kind:     ControlKindHTMX,
			Label:    label,
			Disabled: true,
		}
	}

	navCtrl := func(label string, page int) Control {
		return Control{
			Kind:      ControlKindHTMX,
			Label:     label,
			HxRequest: buildReq(info.URLForPage(page)),
		}
	}

	var controls []Control

	// First (««) and Prev (‹)
	if info.Page <= 1 {
		controls = append(controls, disabledCtrl(PaginationFirst), disabledCtrl(PaginationPrev))
	} else {
		controls = append(controls, navCtrl(PaginationFirst, 1), navCtrl(PaginationPrev, info.Page-1))
	}

	// Page window: page-2 … page+2, capped at boundaries.
	windowStart := info.Page - 2
	windowEnd := info.Page + 2
	if windowStart < 1 {
		windowStart = 1
	}
	if windowEnd > info.TotalPages {
		windowEnd = info.TotalPages
	}

	for p := windowStart; p <= windowEnd; p++ {
		if p == info.Page {
			controls = append(controls, Control{
				Kind:     ControlKindHTMX,
				Label:    strconv.Itoa(p),
				Disabled: true,
				Variant:  VariantPrimary,
			})
		} else {
			controls = append(controls, navCtrl(strconv.Itoa(p), p))
		}
	}

	// Next (›) and Last (»)
	if info.Page >= info.TotalPages {
		controls = append(controls, disabledCtrl(PaginationNext), disabledCtrl(PaginationLast))
	} else {
		controls = append(controls, navCtrl(PaginationNext, info.Page+1), navCtrl(PaginationLast, info.TotalPages))
	}

	return controls
}
