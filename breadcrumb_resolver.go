package linkwell

// BreadcrumbResolveContext carries request-local breadcrumb inputs.
type BreadcrumbResolveContext struct {
	// Path is the current request path used by policy and path fallback sources.
	Path string
	// Explicit is a route-provided breadcrumb trail.
	Explicit []Breadcrumb
	// Parent is an app-provided parent/up breadcrumb trail.
	Parent []Breadcrumb
	// RuntimeLabels are per-request labels passed to a BreadcrumbPolicy source.
	RuntimeLabels []CrumbOption
	// PathLabels override path-derived fallback labels by segment index.
	PathLabels map[int]string
}

// BreadcrumbSource resolves a breadcrumb trail from request-local context.
type BreadcrumbSource func(BreadcrumbResolveContext) []Breadcrumb

// BreadcrumbResolver tries breadcrumb sources in order and returns the first
// non-empty trail.
type BreadcrumbResolver struct {
	sources []BreadcrumbSource
}

// NewBreadcrumbResolver builds an ordered breadcrumb fallback resolver.
func NewBreadcrumbResolver(sources ...BreadcrumbSource) BreadcrumbResolver {
	copied := make([]BreadcrumbSource, len(sources))
	copy(copied, sources)
	return BreadcrumbResolver{sources: copied}
}

// Resolve returns the first non-empty trail produced by the configured sources.
func (r BreadcrumbResolver) Resolve(ctx BreadcrumbResolveContext) []Breadcrumb {
	for _, source := range r.sources {
		if source == nil {
			continue
		}
		crumbs := cloneBreadcrumbs(source(ctx))
		if len(crumbs) > 0 {
			return crumbs
		}
	}
	return nil
}

// ExplicitBreadcrumbs returns route-provided breadcrumbs from the context.
func ExplicitBreadcrumbs() BreadcrumbSource {
	return func(ctx BreadcrumbResolveContext) []Breadcrumb {
		return cloneBreadcrumbs(ctx.Explicit)
	}
}

// ParentBreadcrumbs returns app-provided parent/up breadcrumbs from the context.
func ParentBreadcrumbs() BreadcrumbSource {
	return func(ctx BreadcrumbResolveContext) []Breadcrumb {
		return cloneBreadcrumbs(ctx.Parent)
	}
}

// PolicyBreadcrumbs resolves breadcrumbs from a BreadcrumbPolicy and runtime labels.
func PolicyBreadcrumbs(policy BreadcrumbPolicy) BreadcrumbSource {
	return func(ctx BreadcrumbResolveContext) []Breadcrumb {
		return policy.Resolve(ctx.Path, ctx.RuntimeLabels...)
	}
}

// PathBreadcrumbs resolves breadcrumbs from the current path.
func PathBreadcrumbs() BreadcrumbSource {
	return func(ctx BreadcrumbResolveContext) []Breadcrumb {
		return BreadcrumbsFromPath(normalizeCrumbPath(ctx.Path), ctx.PathLabels)
	}
}

func cloneBreadcrumbs(crumbs []Breadcrumb) []Breadcrumb {
	if len(crumbs) == 0 {
		return nil
	}
	copied := make([]Breadcrumb, len(crumbs))
	copy(copied, crumbs)
	return copied
}
