package linkwell

import "strings"

// BreadcrumbPolicy is a declarative, request-local description of how a mounted
// app's routes render as breadcrumbs. It is decoupled from the link graph and
// sitemap: those describe durable site topology, while a policy describes route
// display only. It is pure data: build one once (even as a package-level
// value) and call Resolve per request. Builder methods return a new policy and
// never mutate the receiver, so shared reuse across requests is safe.
type BreadcrumbPolicy struct {
	prefix    string
	rootLabel string
	rootHref  string
	rules     []crumbRule
}

// crumbRule is a route pattern (relative to the policy prefix) paired with a
// display label. Segments beginning with ":" match any single path segment.
type crumbRule struct {
	segs   []string
	label  string
	exacts int // non-":param" segment count, for specificity tiebreak
}

// CrumbOption is a per-call runtime label, typically a DB/entity name resolved
// by the handler. Runtime labels override policy labels and are not stored on
// the policy.
type CrumbOption struct {
	rule crumbRule
}

// Breadcrumbs starts a new empty BreadcrumbPolicy.
func Breadcrumbs() BreadcrumbPolicy { return BreadcrumbPolicy{} }

// Prefix sets the mounted-app path prefix trimmed before matching. Returned
// crumb Hrefs remain full absolute paths.
func (p BreadcrumbPolicy) Prefix(prefix string) BreadcrumbPolicy {
	p.prefix = normalizeCrumbPath(prefix)
	return p
}

// Root sets the root crumb label. With a prefix, its Href defaults to that
// prefix. Without a prefix, its Href defaults to "/". Override with RootHref.
func (p BreadcrumbPolicy) Root(label string) BreadcrumbPolicy {
	p.rootLabel = label
	return p
}

// RootHref overrides the app-local root crumb Href, which otherwise defaults to
// the prefix.
func (p BreadcrumbPolicy) RootHref(href string) BreadcrumbPolicy {
	p.rootHref = href
	return p
}

// Crumb registers a stable policy label for a route pattern relative to the
// prefix. Patterns support exact segments and ":param" segments that each match
// one path segment.
func (p BreadcrumbPolicy) Crumb(pattern, label string) BreadcrumbPolicy {
	rules := make([]crumbRule, len(p.rules), len(p.rules)+1)
	copy(rules, p.rules)
	p.rules = append(rules, newCrumbRule(pattern, label))
	return p
}

// CrumbLabel builds a per-call runtime label for a route pattern. Pass to
// Resolve to override the policy label for the matching segment, e.g. to supply
// an entity name for a ":id" segment.
func CrumbLabel(pattern, label string) CrumbOption {
	return CrumbOption{rule: newCrumbRule(pattern, label)}
}

// Resolve renders breadcrumbs for path under this policy. Parent crumbs carry
// full absolute Hrefs; the terminal (current page) crumb has an empty Href.
// Runtime labels override policy labels; unmatched segments fall back to
// TitleFromPath.
func (p BreadcrumbPolicy) Resolve(path string, runtime ...CrumbOption) []Breadcrumb {
	path = normalizeCrumbPath(path)

	rel := path
	hasPrefix := p.prefix == ""
	if p.prefix != "" {
		switch {
		case path == p.prefix:
			rel, hasPrefix = "", true
		case strings.HasPrefix(path, p.prefix+"/"):
			rel, hasPrefix = path[len(p.prefix):], true
		}
	}

	relSegs := splitCrumbSegs(rel)

	base := ""
	if hasPrefix {
		base = p.prefix
	}

	var crumbs []Breadcrumb
	if hasPrefix && p.rootLabel != "" {
		href := p.rootHref
		if href == "" {
			href = "/"
			if p.prefix != "" {
				href = p.prefix
			}
		}
		if len(relSegs) == 0 {
			href = "" // the root itself is the current page
		}
		crumbs = append(crumbs, Breadcrumb{Label: p.rootLabel, Href: href})
	}

	for i := range relSegs {
		abs := base + "/" + strings.Join(relSegs[:i+1], "/")
		label := p.labelFor(relSegs[:i+1], runtime)
		if label == "" {
			label = TitleFromPath(abs)
		}
		href := abs
		if i == len(relSegs)-1 {
			href = "" // terminal crumb is the current page
		}
		crumbs = append(crumbs, Breadcrumb{Label: label, Href: href})
	}
	return crumbs
}

// labelFor picks the best label for a cumulative path. Runtime labels beat
// policy labels; within a source, more exact (fewer ":param") patterns win;
// ties keep the first registered.
func (p BreadcrumbPolicy) labelFor(pathSegs []string, runtime []CrumbOption) string {
	best := ""
	bestScore := -1
	consider := func(r crumbRule, isRuntime bool) {
		if !matchCrumbSegs(r.segs, pathSegs) {
			return
		}
		score := r.exacts
		if isRuntime {
			score += 1000 // runtime overrides any policy label
		}
		if score > bestScore {
			bestScore = score
			best = r.label
		}
	}
	for _, r := range p.rules {
		consider(r, false)
	}
	for _, o := range runtime {
		consider(o.rule, true)
	}
	return best
}

func newCrumbRule(pattern, label string) crumbRule {
	segs := splitCrumbSegs(pattern)
	exacts := 0
	for _, s := range segs {
		if !strings.HasPrefix(s, ":") {
			exacts++
		}
	}
	return crumbRule{segs: segs, label: label, exacts: exacts}
}

// matchCrumbSegs reports whether pattern matches path. ":param" segments match
// any single segment; lengths must be equal, so a param never spans segments.
func matchCrumbSegs(pattern, path []string) bool {
	if len(pattern) != len(path) {
		return false
	}
	for i, ps := range pattern {
		if strings.HasPrefix(ps, ":") {
			continue
		}
		if ps != path[i] {
			return false
		}
	}
	return true
}

func splitCrumbSegs(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

// normalizeCrumbPath strips any query/fragment and trailing slash so matching
// is stable. The root "/" is preserved.
func normalizeCrumbPath(path string) string {
	if i := strings.IndexAny(path, "?#"); i >= 0 {
		path = path[:i]
	}
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}
	return path
}
