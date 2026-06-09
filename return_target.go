package linkwell

import "net/http"

// Navigation contract (issue #60):
//
//   - Canonical parent/back is structural and server-owned: the durable "up"
//     destination for a page, independent of how the user arrived. Modeled by
//     CanonicalTarget and supplied by the application, so it is trusted data.
//   - Exact return is request-specific: the precise page a journey should return
//     to, carried in a query parameter (default "back_to"). It originates from
//     the request, so it is validated as a same-origin path before use.
//   - The ?from= breadcrumb context (see FromNav) is display/origin hinting, not
//     an exact return target. Never treat a "from" value as a safe redirect.
//
// ReturnTarget collapses these into one destination: an accepted same-origin
// exact return when present and safe, otherwise the canonical fallback.

// DefaultReturnParam is the query parameter consulted for an exact return path
// when ReturnTargetConfig.Param is empty.
const DefaultReturnParam = "back_to"

// DefaultReturnLabel is the label applied to an accepted exact return when
// ReturnTargetConfig.ExactLabel is empty.
const DefaultReturnLabel = "Back"

// CanonicalTarget is a server-owned structural back/up destination. It is
// trusted application data (not derived from the request), so its Href is used
// verbatim and is not subjected to same-origin validation.
type CanonicalTarget struct {
	// Label is the user-visible text for the canonical destination.
	Label string
	// Href is the canonical destination path (server-owned, trusted).
	Href string
}

// ReturnTargetConfig configures how an exact return is read from a request and
// how the canonical fallback is described.
type ReturnTargetConfig struct {
	// Param is the query parameter holding the exact return path. Defaults to
	// DefaultReturnParam ("back_to") when empty.
	Param string
	// ExactLabel is the label for an accepted exact return. Defaults to
	// DefaultReturnLabel ("Back") when empty.
	ExactLabel string
	// Fallback is the canonical target used when no safe exact return is present.
	Fallback CanonicalTarget
}

// ReturnTarget is the resolved destination a "return" affordance should use.
// Exact reports whether Href came from an accepted same-origin exact return
// (true) or from the canonical fallback (false).
type ReturnTarget struct {
	// Label is the user-visible text for the return control.
	Label string
	// Href is the resolved, safe destination path.
	Href string
	// Exact is true when Href is an accepted exact return, false for fallback.
	Exact bool
}

// ReturnTargetFromRequest resolves a ReturnTarget from the request's return
// parameter, falling back to cfg.Fallback when the parameter is absent or is
// not a safe same-origin path. A nil request resolves to the fallback.
func ReturnTargetFromRequest(r *http.Request, cfg ReturnTargetConfig) ReturnTarget {
	raw := ""
	if r != nil {
		raw = r.URL.Query().Get(returnParam(cfg))
	}
	return ReturnTargetFromValue(raw, cfg)
}

// ReturnTargetFromValue resolves a ReturnTarget from an already-extracted raw
// parameter value. Use when the return value is read outside net/http (e.g. a
// framework's own query accessor). raw is accepted only when it is a safe local
// absolute path; otherwise the canonical fallback is returned.
func ReturnTargetFromValue(raw string, cfg ReturnTargetConfig) ReturnTarget {
	if safe, ok := safeLocalPath(raw); ok {
		label := cfg.ExactLabel
		if label == "" {
			label = DefaultReturnLabel
		}
		return ReturnTarget{Label: label, Href: safe, Exact: true}
	}
	return ReturnTarget{Label: cfg.Fallback.Label, Href: cfg.Fallback.Href}
}

func returnParam(cfg ReturnTargetConfig) string {
	if cfg.Param == "" {
		return DefaultReturnParam
	}
	return cfg.Param
}

// safeLocalPath reports whether raw is a same-origin local path safe to use as
// a return/redirect target, returning the value to use. Query and fragment are
// preserved; neither can change the origin. Rejected: empty values, values not
// beginning with a single "/", protocol-relative ("//host") and backslash
// ("/\host", or any "\" a browser may fold to "/") forms, absolute URLs (which
// carry a scheme before the first "/"), and values containing raw or
// percent-encoded control characters or backslashes.
func safeLocalPath(raw string) (string, bool) {
	if raw == "" || raw[0] != '/' {
		return "", false
	}
	if len(raw) > 1 && (raw[1] == '/' || raw[1] == '\\') {
		return "", false
	}
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if c == '\\' || c < 0x20 || c == 0x7f {
			return "", false
		}
		if c == '%' && unsafeEscapedByte(raw[i:]) {
			return "", false
		}
	}
	return raw, true
}

func unsafeEscapedByte(s string) bool {
	if len(s) < 3 {
		return false
	}
	hi, ok := fromHex(s[1])
	if !ok {
		return false
	}
	lo, ok := fromHex(s[2])
	if !ok {
		return false
	}
	b := hi<<4 | lo
	return b == '\\' || b < 0x20 || b == 0x7f
}

func fromHex(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}
