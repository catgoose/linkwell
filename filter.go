package linkwell

// FilterKind identifies the HTML input type for a FilterField. Templates use
// this to select the appropriate form control (text input, select, range slider,
// checkbox, or date picker).
type FilterKind string

// Filter kind constants for FilterField.Kind.
const (
	FilterKindSearch   FilterKind = "search"
	FilterKindSelect   FilterKind = "select"
	FilterKindRange    FilterKind = "range"
	FilterKindCheckbox FilterKind = "checkbox"
	FilterKindDate     FilterKind = "date"
)

// FilterOption is a single option in a select dropdown. Selected is typically
// set by SelectOptions based on the current query parameter value.
type FilterOption struct {
	Value    string
	Label    string
	Selected bool
}

// FilterField is a pure-data descriptor for a single filter input. Value always
// holds the current serialized value as a string regardless of Kind:
//   - Search/Select/Date: the query parameter value
//   - Checkbox: "true" when checked, "" when unchecked (param absent)
//   - Range: the current numeric string; Min/Max/Step bound the slider
type FilterField struct {
	// HTMXAttrs holds optional HTMX attributes for this field (e.g., for
	// cascading filter updates triggered by hx-get on change).
	HTMXAttrs map[string]string
	// Kind determines the HTML input type rendered by the template.
	Kind FilterKind
	// Name is the form field name (maps to the query parameter key).
	Name string
	// Label is the visible label text. Used by select, range, checkbox, and date.
	Label string
	// Placeholder is hint text for search inputs.
	Placeholder string
	// Value is the current serialized value from the query string.
	Value string
	// Min is the minimum value for range inputs.
	Min string
	// Max is the maximum value for range inputs.
	Max string
	// Step is the increment step for range inputs.
	Step string
	// Options is the list of choices for select inputs.
	Options []FilterOption
	// Disabled renders the field in a non-interactive state.
	Disabled bool
}

// FilterBar is the descriptor for a complete filter form. The form element
// carries an HTML id so that pagination and sort links can use
// hx-include="#filter-form" to forward filter state with their requests.
type FilterBar struct {
	// ID is the HTML form id attribute. Defaults to DefaultFilterFormID ("filter-form")
	// when created via NewFilterBar.
	ID string
	// Action is the hx-get endpoint the form submits to (e.g., "/users").
	Action string
	// Target is the hx-target CSS selector (e.g., "#table-container").
	Target string
	// Fields is the ordered list of filter inputs rendered in the form.
	Fields []FilterField
}

// DefaultFilterFormID is the HTML form id used by NewFilterBar.
const DefaultFilterFormID = "filter-form"

// NewFilterBar creates a FilterBar with the default form ID and the given
// action endpoint, HTMX target, and fields.
func NewFilterBar(action, target string, fields ...FilterField) FilterBar {
	return FilterBar{
		ID:     DefaultFilterFormID,
		Action: action,
		Target: target,
		Fields: fields,
	}
}

// SearchField creates a text search input with the given name, placeholder, and
// current value.
func SearchField(name, placeholder, value string) FilterField {
	return FilterField{
		Kind:        FilterKindSearch,
		Name:        name,
		Placeholder: placeholder,
		Value:       value,
	}
}

// SelectField creates a select dropdown with the given options. Use
// SelectOptions to build the options slice from flat value/label pairs.
func SelectField(name, label, value string, options []FilterOption) FilterField {
	return FilterField{
		Kind:    FilterKindSelect,
		Name:    name,
		Label:   label,
		Value:   value,
		Options: options,
	}
}

// RangeField creates a range slider input bounded by min, max, and step.
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

// CheckboxField creates a boolean toggle. Pass the raw query parameter value:
// "true" when checked, "" when unchecked or absent.
func CheckboxField(name, label, value string) FilterField {
	return FilterField{
		Kind:  FilterKindCheckbox,
		Name:  name,
		Label: label,
		Value: value,
	}
}

// DateField creates a date picker input.
func DateField(name, label, value string) FilterField {
	return FilterField{
		Kind:  FilterKindDate,
		Name:  name,
		Label: label,
		Value: value,
	}
}

// SelectOptions builds a slice of FilterOption from flat value/label pairs.
// The current parameter is matched against each value to set Selected=true on
// the matching option. Unpaired trailing values are safely ignored.
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

// FilterGroup wraps a FilterBar with methods for dynamic option updates and
// OOB (out-of-band) swap rendering. Use when filter options change based on
// other selections (e.g., cascading dropdowns) and the server needs to push
// updated select elements back to the page via HTMX OOB swaps.
type FilterGroup struct {
	// Bar is the underlying filter bar configuration.
	Bar FilterBar
}

// NewFilterGroup creates a FilterGroup with a default-ID FilterBar.
func NewFilterGroup(action, target string, fields ...FilterField) FilterGroup {
	return FilterGroup{
		Bar: NewFilterBar(action, target, fields...),
	}
}

// UpdateOptions replaces the options for the named select field. Non-select
// fields or unrecognized names are silently ignored.
func (g *FilterGroup) UpdateOptions(name string, options []FilterOption) {
	for i := range g.Bar.Fields {
		if g.Bar.Fields[i].Name == name && g.Bar.Fields[i].Kind == FilterKindSelect {
			g.Bar.Fields[i].Options = options
			return
		}
	}
}

// SelectFields returns only the select-type fields from the bar. Use when
// rendering OOB swap fragments that update dropdown options without replacing
// the entire filter form.
func (g *FilterGroup) SelectFields() []FilterField {
	var fields []FilterField
	for _, f := range g.Bar.Fields {
		if f.Kind == FilterKindSelect {
			fields = append(fields, f)
		}
	}
	return fields
}
