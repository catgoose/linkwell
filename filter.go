package linkwell

// FilterKind identifies the input type for a FilterField.
type FilterKind string

// Filter kind constants for FilterField.Kind.
const (
	FilterKindSearch   FilterKind = "search"
	FilterKindSelect   FilterKind = "select"
	FilterKindRange    FilterKind = "range"
	FilterKindCheckbox FilterKind = "checkbox"
	FilterKindDate     FilterKind = "date"
)

// FilterOption is a single option in a select dropdown.
type FilterOption struct {
	Value    string
	Label    string
	Selected bool
}

// FilterField is a pure-data descriptor for a single filter input.
// Value always holds the current serialized value (string) regardless of Kind.
// For checkboxes: Value == "true" means checked; "" means unchecked (param absent).
// For range: Value is the current numeric string; Min/Max/Step bound the slider.
type FilterField struct {
	HTMXAttrs   map[string]string
	Kind        FilterKind
	Name        string
	Label       string
	Placeholder string
	Value       string
	Min         string
	Max         string
	Step        string
	Options     []FilterOption
	Disabled    bool
}

// FilterBar is the descriptor for the full filter form.
// The form element has ID so pagination/sort links can use hx-include="#filter-form".
type FilterBar struct {
	ID     string // HTML form id (default "filter-form")
	Action string // hx-get endpoint (e.g., "/users")
	Target string // hx-target CSS selector (e.g., "#table-container")
	Fields []FilterField
}

// DefaultFilterFormID is the HTML form id used by NewFilterBar.
const DefaultFilterFormID = "filter-form"

// NewFilterBar creates a FilterBar with ID=DefaultFilterFormID.
func NewFilterBar(action, target string, fields ...FilterField) FilterBar {
	return FilterBar{
		ID:     DefaultFilterFormID,
		Action: action,
		Target: target,
		Fields: fields,
	}
}

// SearchField creates a text search input.
func SearchField(name, placeholder, value string) FilterField {
	return FilterField{
		Kind:        FilterKindSearch,
		Name:        name,
		Placeholder: placeholder,
		Value:       value,
	}
}

// SelectField creates a <select> dropdown.
func SelectField(name, label, value string, options []FilterOption) FilterField {
	return FilterField{
		Kind:    FilterKindSelect,
		Name:    name,
		Label:   label,
		Value:   value,
		Options: options,
	}
}

// RangeField creates a range slider.
func RangeField(name, label, value, min, max, step string) FilterField {
	return FilterField{
		Kind:  FilterKindRange,
		Name:  name,
		Label: label,
		Value: value,
		Min:   min,
		Max:   max,
		Step:  step,
	}
}

// CheckboxField creates a boolean toggle.
// value should be c.QueryParam(name) — "true" if checked, "" if unchecked/absent.
func CheckboxField(name, label, value string) FilterField {
	return FilterField{
		Kind:  FilterKindCheckbox,
		Name:  name,
		Label: label,
		Value: value,
	}
}

// DateField creates a date input.
func DateField(name, label, value string) FilterField {
	return FilterField{
		Kind:  FilterKindDate,
		Name:  name,
		Label: label,
		Value: value,
	}
}

// SelectOptions builds a []FilterOption from flat pairs ["val","Label",...].
// current is matched against Value to set Selected=true.
// Odd pairs are handled safely by ignoring the trailing unpaired value.
func SelectOptions(current string, pairs ...string) []FilterOption {
	options := make([]FilterOption, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		val := pairs[i]
		label := pairs[i+1]
		options = append(options, FilterOption{
			Value:    val,
			Label:    label,
			Selected: val == current,
		})
	}
	return options
}

// FilterGroup is a configuration group of filter controls that can produce
// OOB swap fragments when the available options change. It wraps a FilterBar
// and adds the ability to update select field options and render OOB updates
// using data-filter attributes instead of element IDs.
type FilterGroup struct {
	Bar FilterBar
}

// NewFilterGroup creates a FilterGroup with the given fields.
func NewFilterGroup(action, target string, fields ...FilterField) FilterGroup {
	return FilterGroup{
		Bar: NewFilterBar(action, target, fields...),
	}
}

// UpdateOptions replaces the options for the named select field.
func (g *FilterGroup) UpdateOptions(name string, options []FilterOption) {
	for i := range g.Bar.Fields {
		if g.Bar.Fields[i].Name == name && g.Bar.Fields[i].Kind == FilterKindSelect {
			g.Bar.Fields[i].Options = options
			return
		}
	}
}

// SelectFields returns only the select-type fields for OOB rendering.
func (g *FilterGroup) SelectFields() []FilterField {
	var fields []FilterField
	for _, f := range g.Bar.Fields {
		if f.Kind == FilterKindSelect {
			fields = append(fields, f)
		}
	}
	return fields
}
