package linkwell

// StepStatus represents the state of a step in a multi-step wizard flow.
type StepStatus string

const (
	// StepPending indicates the step has not been reached yet.
	StepPending StepStatus = "pending"
	// StepActive indicates the step is the current step.
	StepActive StepStatus = "active"
	// StepComplete indicates the step has been completed.
	StepComplete StepStatus = "complete"
	// StepSkipped indicates the step was skipped.
	StepSkipped StepStatus = "skipped"
)

// Default labels for stepper navigation controls.
const (
	LabelPrevious = "Previous"
	LabelNext     = "Next"
	LabelSubmit   = "Submit"
)

// Step describes a single step in a multi-step wizard flow.
type Step struct {
	// Label is the user-visible name for this step.
	Label string
	// Href is the URL for this step.
	Href string
	// Status is the current state of this step (pending, active, complete, skipped).
	Status StepStatus
	// Icon is an optional icon name rendered alongside the step label.
	Icon Icon
}

// StepperConfig describes a multi-step wizard flow where the server knows the
// full step sequence, current position, and completion state. Templates consume
// this to render step indicators, progress bars, and navigation controls.
type StepperConfig struct {
	// Steps is the ordered list of steps in the wizard.
	Steps []Step
	// Current is the 0-based index of the active step.
	Current int
	// Prev is a navigation control pointing to the previous step. Nil if the
	// current step is the first step.
	Prev *Control
	// Next is a navigation control pointing to the next step. Nil if the
	// current step is the last step.
	Next *Control
	// Submit is a submit control shown only on the final step. Nil on all
	// other steps.
	Submit *Control
}

// NewStepper creates a StepperConfig from the given current index and steps.
// Steps before currentIndex are marked Complete, the step at currentIndex is
// marked Active, and steps after are marked Pending. Pre-set statuses (e.g.,
// StepSkipped) are preserved — only StepPending statuses are overwritten.
//
// Navigation controls are auto-generated: Prev points to the previous step's
// Href (nil on the first step), Next points to the next step's Href (nil on
// the last step), and Submit is generated only on the final step.
func NewStepper(currentIndex int, steps ...Step) StepperConfig {
	out := make([]Step, len(steps))
	copy(out, steps)

	for i := range out {
		switch {
		case i < currentIndex:
			if out[i].Status == "" || out[i].Status == StepPending {
				out[i].Status = StepComplete
			}
		case i == currentIndex:
			out[i].Status = StepActive
		default:
			if out[i].Status == "" {
				out[i].Status = StepPending
			}
		}
	}

	cfg := StepperConfig{
		Steps:   out,
		Current: currentIndex,
	}

	if currentIndex > 0 && currentIndex < len(out) {
		prev := RedirectLink(LabelPrevious, out[currentIndex-1].Href)
		cfg.Prev = &prev
	}

	if currentIndex < len(out)-1 {
		next := RedirectLink(LabelNext, out[currentIndex+1].Href)
		cfg.Next = &next
	}

	if currentIndex == len(out)-1 && len(out) > 0 {
		submit := RedirectLink(LabelSubmit, out[currentIndex].Href).WithVariant(VariantPrimary)
		cfg.Submit = &submit
	}

	return cfg
}
