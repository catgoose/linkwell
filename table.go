package linkwell

import (
	"net/url"
	"strconv"
)

// SortDir indicates the current sort direction for a column.
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

// TableCol is a pure-data column descriptor.
// Sort metadata is precomputed by the handler — templates just render it.
type TableCol struct {
	Key      string
	Label    string
	SortDir  SortDir
	SortURL  string
	Target   string
	Include  string
	Width    string
	Sortable bool
}

// PageInfo carries server-computed pagination state.
type PageInfo struct {
	BaseURL    string
	PageParam  string
	Target     string
	Include    string
	Page       int
	PerPage    int
	TotalItems int
	TotalPages int
}

// pageParam returns the PageParam field, defaulting to "page" when empty.
func (pi PageInfo) pageParam() string {
	if pi.PageParam == "" {
		return ParamPage
	}
	return pi.PageParam
}

// URLForPage builds the URL for a specific page using net/url for correct encoding.
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

// ComputeTotalPages returns ceil(totalItems / perPage); minimum 1.
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

// SortableCol creates a TableCol with sort state and toggle URL computed from current request.
// baseURL: current URL with "sort" and "dir" params already stripped.
// Toggle logic:
//   - non-matching key → SortNone, URL points to asc.
//   - matching key + asc → SortAsc, URL toggles to desc.
//   - matching key + desc → SortDesc, URL toggles to asc.
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

// PaginationControls generates the []Control slice for a PaginationBar.
// Exported so it can be tested independently.
// Returns nil if TotalPages <= 1.
// Control semantics:
//
//	Current page:       Kind=ControlKindHTMX, Disabled=true, Variant=VariantPrimary
//	Disabled boundary:  Kind=ControlKindHTMX, Disabled=true
//	Normal page/nav:    Kind=ControlKindHTMX, HxRequest={Method: HxMethodGet, URL: url, Target: ..., Include: ...}
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
