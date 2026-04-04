package linkwell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStepper_MiddleStep(t *testing.T) {
	stepper := NewStepper(1,
		Step{Label: "Account", Href: "/onboard/account"},
		Step{Label: "Profile", Href: "/onboard/profile"},
		Step{Label: "Preferences", Href: "/onboard/prefs"},
		Step{Label: "Review", Href: "/onboard/review"},
	)

	require.Len(t, stepper.Steps, 4)
	require.Equal(t, 1, stepper.Current)

	// Steps before current are Complete.
	require.Equal(t, StepComplete, stepper.Steps[0].Status)
	// Current step is Active.
	require.Equal(t, StepActive, stepper.Steps[1].Status)
	// Steps after current are Pending.
	require.Equal(t, StepPending, stepper.Steps[2].Status)
	require.Equal(t, StepPending, stepper.Steps[3].Status)

	// Prev points to step 0.
	require.NotNil(t, stepper.Prev)
	require.Equal(t, "/onboard/account", stepper.Prev.Href)
	require.Equal(t, LabelPrevious, stepper.Prev.Label)
	require.Equal(t, ControlKindLink, stepper.Prev.Kind)

	// Next points to step 2.
	require.NotNil(t, stepper.Next)
	require.Equal(t, "/onboard/prefs", stepper.Next.Href)
	require.Equal(t, LabelNext, stepper.Next.Label)
	require.Equal(t, ControlKindLink, stepper.Next.Kind)

	// Submit is nil (not on last step).
	require.Nil(t, stepper.Submit)
}

func TestNewStepper_FirstStep(t *testing.T) {
	stepper := NewStepper(0,
		Step{Label: "Account", Href: "/onboard/account"},
		Step{Label: "Profile", Href: "/onboard/profile"},
		Step{Label: "Preferences", Href: "/onboard/prefs"},
	)

	require.Equal(t, StepActive, stepper.Steps[0].Status)
	require.Equal(t, StepPending, stepper.Steps[1].Status)
	require.Equal(t, StepPending, stepper.Steps[2].Status)

	// No Prev on first step.
	require.Nil(t, stepper.Prev)

	// Next points to step 1.
	require.NotNil(t, stepper.Next)
	require.Equal(t, "/onboard/profile", stepper.Next.Href)

	// Submit is nil (not on last step).
	require.Nil(t, stepper.Submit)
}

func TestNewStepper_LastStep(t *testing.T) {
	stepper := NewStepper(2,
		Step{Label: "Account", Href: "/onboard/account"},
		Step{Label: "Profile", Href: "/onboard/profile"},
		Step{Label: "Review", Href: "/onboard/review"},
	)

	require.Equal(t, StepComplete, stepper.Steps[0].Status)
	require.Equal(t, StepComplete, stepper.Steps[1].Status)
	require.Equal(t, StepActive, stepper.Steps[2].Status)

	// Prev points to step 1.
	require.NotNil(t, stepper.Prev)
	require.Equal(t, "/onboard/profile", stepper.Prev.Href)

	// No Next on last step.
	require.Nil(t, stepper.Next)

	// Submit is present on last step.
	require.NotNil(t, stepper.Submit)
	require.Equal(t, LabelSubmit, stepper.Submit.Label)
	require.Equal(t, "/onboard/review", stepper.Submit.Href)
	require.Equal(t, VariantPrimary, stepper.Submit.Variant)
}

func TestNewStepper_SingleStep(t *testing.T) {
	stepper := NewStepper(0,
		Step{Label: "Only Step", Href: "/only"},
	)

	require.Len(t, stepper.Steps, 1)
	require.Equal(t, StepActive, stepper.Steps[0].Status)

	// Single step: no Prev, no Next.
	require.Nil(t, stepper.Prev)
	require.Nil(t, stepper.Next)

	// Single step is also the last step, so Submit is present.
	require.NotNil(t, stepper.Submit)
	require.Equal(t, LabelSubmit, stepper.Submit.Label)
}

func TestNewStepper_PreservesSkippedStatus(t *testing.T) {
	stepper := NewStepper(2,
		Step{Label: "Account", Href: "/account"},
		Step{Label: "Optional", Href: "/optional", Status: StepSkipped},
		Step{Label: "Review", Href: "/review"},
	)

	// Skipped status is preserved even though the step is before current.
	require.Equal(t, StepSkipped, stepper.Steps[1].Status)
	// Regular step before current is marked Complete.
	require.Equal(t, StepComplete, stepper.Steps[0].Status)
	require.Equal(t, StepActive, stepper.Steps[2].Status)
}

func TestNewStepper_PreservesIcons(t *testing.T) {
	stepper := NewStepper(0,
		Step{Label: "Account", Href: "/account", Icon: IconHome},
		Step{Label: "Profile", Href: "/profile", Icon: IconCheck},
	)

	require.Equal(t, IconHome, stepper.Steps[0].Icon)
	require.Equal(t, IconCheck, stepper.Steps[1].Icon)
}

func TestNewStepper_PreservesLabelsAndHrefs(t *testing.T) {
	stepper := NewStepper(1,
		Step{Label: "Step A", Href: "/a"},
		Step{Label: "Step B", Href: "/b"},
		Step{Label: "Step C", Href: "/c"},
	)

	require.Equal(t, "Step A", stepper.Steps[0].Label)
	require.Equal(t, "/a", stepper.Steps[0].Href)
	require.Equal(t, "Step B", stepper.Steps[1].Label)
	require.Equal(t, "/b", stepper.Steps[1].Href)
	require.Equal(t, "Step C", stepper.Steps[2].Label)
	require.Equal(t, "/c", stepper.Steps[2].Href)
}

func TestStepStatusConstants(t *testing.T) {
	require.Equal(t, StepStatus("pending"), StepPending)
	require.Equal(t, StepStatus("active"), StepActive)
	require.Equal(t, StepStatus("complete"), StepComplete)
	require.Equal(t, StepStatus("skipped"), StepSkipped)
}
