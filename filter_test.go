package linkwell_test

import (
	"testing"

	"github.com/catgoose/linkwell"
)

func TestSearchField_Kind(t *testing.T) {
	f := linkwell.SearchField("q", "Search…", "foo")
	if f.Kind != linkwell.FilterKindSearch {
		t.Errorf("expected FilterKindSearch, got %q", f.Kind)
	}
	if f.Name != "q" {
		t.Errorf("expected Name=%q, got %q", "q", f.Name)
	}
	if f.Placeholder != "Search…" {
		t.Errorf("expected Placeholder=%q, got %q", "Search…", f.Placeholder)
	}
	if f.Value != "foo" {
		t.Errorf("expected Value=%q, got %q", "foo", f.Value)
	}
}

func TestSelectField_KindAndOptions(t *testing.T) {
	opts := []linkwell.FilterOption{{Value: "a", Label: "A"}}
	f := linkwell.SelectField("role", "Role", "a", opts)
	if f.Kind != linkwell.FilterKindSelect {
		t.Errorf("expected FilterKindSelect, got %q", f.Kind)
	}
	if f.Label != "Role" {
		t.Errorf("expected Label=%q, got %q", "Role", f.Label)
	}
	if len(f.Options) != 1 {
		t.Fatalf("expected 1 option, got %d", len(f.Options))
	}
	if f.Options[0].Value != "a" {
		t.Errorf("expected option Value=%q, got %q", "a", f.Options[0].Value)
	}
}

func TestRangeField_MinMaxStep(t *testing.T) {
	f := linkwell.RangeField("age", "Max age", "30", "18", "100", "1")
	if f.Kind != linkwell.FilterKindRange {
		t.Errorf("expected FilterKindRange, got %q", f.Kind)
	}
	if f.Min != "18" {
		t.Errorf("expected Min=%q, got %q", "18", f.Min)
	}
	if f.Max != "100" {
		t.Errorf("expected Max=%q, got %q", "100", f.Max)
	}
	if f.Step != "1" {
		t.Errorf("expected Step=%q, got %q", "1", f.Step)
	}
	if f.Value != "30" {
		t.Errorf("expected Value=%q, got %q", "30", f.Value)
	}
}

func TestCheckboxField_ValueTrue(t *testing.T) {
	f := linkwell.CheckboxField("active", "Active", "true")
	if f.Kind != linkwell.FilterKindCheckbox {
		t.Errorf("expected FilterKindCheckbox, got %q", f.Kind)
	}
	if f.Value != "true" {
		t.Errorf("expected Value=%q, got %q", "true", f.Value)
	}
}

func TestCheckboxField_ValueEmpty(t *testing.T) {
	f := linkwell.CheckboxField("active", "Active", "")
	if f.Value != "" {
		t.Errorf("expected empty Value, got %q", f.Value)
	}
}

func TestDateField_Kind(t *testing.T) {
	f := linkwell.DateField("from", "From", "2024-01-01")
	if f.Kind != linkwell.FilterKindDate {
		t.Errorf("expected FilterKindDate, got %q", f.Kind)
	}
	if f.Value != "2024-01-01" {
		t.Errorf("expected Value=%q, got %q", "2024-01-01", f.Value)
	}
}

func TestSelectOptions_CorrectOptionSelected(t *testing.T) {
	opts := linkwell.SelectOptions("b", "a", "A", "b", "B", "c", "C")
	if len(opts) != 3 {
		t.Fatalf("expected 3 options, got %d", len(opts))
	}
	if opts[0].Selected {
		t.Error("option 'a' should not be selected")
	}
	if !opts[1].Selected {
		t.Error("option 'b' should be selected")
	}
	if opts[2].Selected {
		t.Error("option 'c' should not be selected")
	}
}

func TestSelectOptions_EmptyCurrentSelectsNothing(t *testing.T) {
	opts := linkwell.SelectOptions("", "a", "A", "b", "B")
	for _, opt := range opts {
		if opt.Selected {
			t.Errorf("no option should be selected when current is empty, but %q is", opt.Value)
		}
	}
}

func TestSelectOptions_OddPairsHandledSafely(t *testing.T) {
	// 3 args = 1 pair + 1 trailing — trailing should be ignored.
	opts := linkwell.SelectOptions("", "a", "A", "orphan")
	if len(opts) != 1 {
		t.Fatalf("expected 1 option from odd pairs, got %d", len(opts))
	}
	if opts[0].Value != "a" {
		t.Errorf("expected Value=%q, got %q", "a", opts[0].Value)
	}
}

func TestNewFilterBar_DefaultID(t *testing.T) {
	bar := linkwell.NewFilterBar("/users", "#table-container",
		linkwell.SearchField("q", "Search", ""),
	)
	if bar.ID != linkwell.DefaultFilterFormID {
		t.Errorf("expected ID=%q, got %q", linkwell.DefaultFilterFormID, bar.ID)
	}
	if bar.Action != "/users" {
		t.Errorf("expected Action=%q, got %q", "/users", bar.Action)
	}
	if bar.Target != "#table-container" {
		t.Errorf("expected Target=%q, got %q", "#table-container", bar.Target)
	}
	if len(bar.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(bar.Fields))
	}
}
