package linkwell

// LinkIssue describes a structural problem detected in the link registry.
type LinkIssue struct {
	// Path is the registry path where the issue was detected.
	Path string
	// Message is a human-readable description of the problem.
	Message string
	// Kind classifies the issue for programmatic filtering.
	// Values: "orphan", "broken_up", "dead_spoke", "unregistered_route", "missing_route".
	Kind string
}

// ValidateGraph checks the link registry for structural issues and returns a
// slice of problems found. An empty slice means no issues were detected.
//
// Detected issues:
//   - orphan: a registered path has no inbound links from any other path.
//   - broken_up: a rel="up" link targets a path that has no registered links.
//   - dead_spoke: a hub spoke path has no links registered (empty link set).
func ValidateGraph() []LinkIssue {
	all := AllLinks()
	hubs := Hubs()

	var issues []LinkIssue

	// Build set of all paths that are link targets (inbound links).
	inbound := make(map[string]bool)
	for _, links := range all {
		for _, l := range links {
			inbound[l.Href] = true
		}
	}

	// Check each registered path for issues.
	for path := range all {
		// Orphan: no other path links to this one.
		if !inbound[path] {
			issues = append(issues, LinkIssue{
				Path:    path,
				Message: "no inbound links from any other registered path",
				Kind:    "orphan",
			})
		}

		// Broken rel="up": target path is not registered.
		for _, l := range all[path] {
			if l.Rel == RelUp {
				if _, ok := all[l.Href]; !ok {
					issues = append(issues, LinkIssue{
						Path:    path,
						Message: "rel=\"up\" target " + l.Href + " is not registered",
						Kind:    "broken_up",
					})
				}
			}
		}
	}

	// Dead hub spokes: spoke path has no links registered.
	for _, hub := range hubs {
		for _, spoke := range hub.Spokes {
			if _, ok := all[spoke.Href]; !ok {
				issues = append(issues, LinkIssue{
					Path:    spoke.Href,
					Message: "hub spoke for " + hub.Title + " (" + hub.Path + ") has no registered links",
					Kind:    "dead_spoke",
				})
			}
		}
	}

	return issues
}

// ValidateAgainstRoutes checks that all registered link paths exist in the
// provided route set, and that all routes have a link graph presence. Returns
// issues for any mismatches.
//
// Detected issues:
//   - unregistered_route: a path in the link registry has no matching route.
//   - missing_route: a route has no presence in the link registry (neither as a
//     source nor as a target).
func ValidateAgainstRoutes(routes []string) []LinkIssue {
	all := AllLinks()

	var issues []LinkIssue

	// Build route set for fast lookup.
	routeSet := make(map[string]bool, len(routes))
	for _, r := range routes {
		routeSet[r] = true
	}

	// Build set of all paths present in the link graph (sources + targets).
	graphPaths := make(map[string]bool)
	for path, links := range all {
		graphPaths[path] = true
		for _, l := range links {
			graphPaths[l.Href] = true
		}
	}

	// Registered paths with no matching route.
	for path := range all {
		if !routeSet[path] {
			issues = append(issues, LinkIssue{
				Path:    path,
				Message: "registered path has no matching route",
				Kind:    "unregistered_route",
			})
		}
	}

	// Routes with no link graph presence.
	for _, route := range routes {
		if !graphPaths[route] {
			issues = append(issues, LinkIssue{
				Path:    route,
				Message: "route has no link graph presence",
				Kind:    "missing_route",
			})
		}
	}

	return issues
}
