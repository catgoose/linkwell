package linkwell

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
)

// DefaultOriginTrailParam is the query parameter carrying an encoded dynamic
// origin breadcrumb trail.
//
// A dynamic origin trail forwards runtime breadcrumb labels and hrefs through a
// link so a destination reached from several places can show where the user came
// from. It complements the static ?from= bitmask (see FromNav), which replays
// breadcrumbs registered at startup. Like ?from=, an origin trail is display
// context, never a redirect authorization — use ReturnTarget for safe returns.
const DefaultOriginTrailParam = "origin"

// Bounds on a decoded or encoded origin trail so hostile query values cannot
// force unbounded work. The encoded cap is checked against the base64 value.
const (
	maxOriginTrailEncodedLen = 2048
	maxOriginCrumbCount      = 16
	maxOriginLabelLen        = 128
	maxOriginHrefLen         = 512
)

// OriginCrumb is one segment of a dynamic origin breadcrumb trail. Label is
// display text; Href is a safe same-origin local path validated with the same
// rules as ReturnTarget. Field order is preserved end to end.
type OriginCrumb struct {
	Label string `json:"l"`
	Href  string `json:"h"`
}

// OriginTrailParam encodes an ordered origin trail into a URL-safe query
// parameter value. It returns "" when the trail is empty or any crumb is invalid
// (empty label or an href that is not a safe same-origin path), so anything it
// emits round-trips through OriginTrailFromValue.
func OriginTrailParam(trail []OriginCrumb) string {
	if len(trail) == 0 {
		return ""
	}
	clean, ok := sanitizeOriginTrail(trail)
	if !ok {
		return ""
	}
	data, err := json.Marshal(clean)
	if err != nil {
		return ""
	}
	encoded := base64.RawURLEncoding.EncodeToString(data)
	if len(encoded) > maxOriginTrailEncodedLen {
		return ""
	}
	return encoded
}

// OriginTrailFromValue decodes and validates an already-extracted origin trail
// parameter value. It returns the ordered crumbs and true only when decoding
// succeeds and every crumb has a non-empty label and a safe same-origin href;
// otherwise it returns (nil, false). Use when the value is read outside
// net/http.
func OriginTrailFromValue(raw string) ([]OriginCrumb, bool) {
	if raw == "" || len(raw) > maxOriginTrailEncodedLen {
		return nil, false
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, false
	}
	var trail []OriginCrumb
	if err := json.Unmarshal(data, &trail); err != nil {
		return nil, false
	}
	return sanitizeOriginTrail(trail)
}

// OriginTrailFromRequest decodes the origin trail carried in the request's
// DefaultOriginTrailParam query parameter. A nil request or an absent or invalid
// value returns (nil, false).
func OriginTrailFromRequest(r *http.Request) ([]OriginCrumb, bool) {
	if r == nil {
		return nil, false
	}
	return OriginTrailFromValue(r.URL.Query().Get(DefaultOriginTrailParam))
}

// OriginNav appends an encoded origin trail to href as the
// DefaultOriginTrailParam query parameter, preserving any existing query string
// and fragment. href is returned unchanged when the trail is empty or cannot be
// encoded safely.
func OriginNav(href string, trail []OriginCrumb) string {
	value := OriginTrailParam(trail)
	if value == "" {
		return href
	}
	return appendQueryParam(href, DefaultOriginTrailParam, value)
}

// OriginCrumbsToBreadcrumbs converts a decoded origin trail into a breadcrumb
// trail for rendering or for use as a BreadcrumbResolver source. Every segment
// keeps its href; linkwell does not load entity labels.
func OriginCrumbsToBreadcrumbs(trail []OriginCrumb) []Breadcrumb {
	if len(trail) == 0 {
		return nil
	}
	crumbs := make([]Breadcrumb, len(trail))
	for i, c := range trail {
		crumbs[i] = Breadcrumb(c)
	}
	return crumbs
}

// sanitizeOriginTrail trims labels and validates every crumb against the origin
// trail bounds and the safe-local-path rules. It rejects the whole trail when it
// is empty, too long, or any crumb is invalid.
func sanitizeOriginTrail(trail []OriginCrumb) ([]OriginCrumb, bool) {
	if len(trail) == 0 || len(trail) > maxOriginCrumbCount {
		return nil, false
	}
	out := make([]OriginCrumb, 0, len(trail))
	for _, c := range trail {
		label := strings.TrimSpace(c.Label)
		if label == "" || len(label) > maxOriginLabelLen {
			return nil, false
		}
		if len(c.Href) > maxOriginHrefLen {
			return nil, false
		}
		href, ok := safeLocalPath(c.Href)
		if !ok {
			return nil, false
		}
		out = append(out, OriginCrumb{Label: label, Href: href})
	}
	return out, true
}

// appendQueryParam appends key=value to href, keeping any existing query string
// and moving the result before a fragment. value must already be URL-safe.
func appendQueryParam(href, key, value string) string {
	target, fragment := href, ""
	if i := strings.IndexByte(href, '#'); i >= 0 {
		target, fragment = href[:i], href[i:]
	}
	sep := "?"
	if strings.IndexByte(target, '?') >= 0 {
		sep = "&"
	}
	return target + sep + key + "=" + value + fragment
}
