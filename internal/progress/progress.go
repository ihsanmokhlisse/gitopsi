// Package progress provides live progress display for gitopsi operations.
package progress

import (
	"fmt"
	"time"

	"github.com/pterm/pterm"
)

// StepStatus represents the status of a step.
type StepStatus string

const (
	StatusPending  StepStatus = "pending"
	StatusRunning  StepStatus = "running"
	StatusSuccess  StepStatus = "success"
	StatusFailed   StepStatus = "failed"
	StatusSkipped  StepStatus = "skipped"
	StatusWarning  StepStatus = "warning"
)

// Step represents a single step in the progress.
type Step struct {
	Name      string
	Status    StepStatus
	Duration  time.Duration
	StartTime time.Time
	Message   string
	SubSteps  []*Step
}

// Section represents a group of steps.
type Section struct {
	Name    string
	Steps   []*Step
	spinner *pterm.SpinnerPrinter
}

// Progress manages the overall progress display.
type Progress struct {
	title       string
	projectName string
	sections    []*Section
	quiet       bool
	jsonOutput  bool
	currentStep *Step
}

// New creates a new Progress instance.
func New(title, projectName string) *Progress {
	return &Progress{
		title:       title,
		projectName: projectName,
		sections:    make([]*Section, 0),
	}
}

// SetQuiet sets quiet mode (minimal output).
func (p *Progress) SetQuiet(quiet bool) {
	p.quiet = quiet
}

// SetJSON sets JSON output mode.
func (p *Progress) SetJSON(json bool) {
	p.jsonOutput = json
}

// ShowHeader displays the header box.
func (p *Progress) ShowHeader() {
	if p.quiet || p.jsonOutput {
		return
	}

	pterm.DefaultBox.WithTitle(pterm.Bold.Sprint("🚀 gitopsi - GitOps Repository Generator")).
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(pterm.FgCyan)).
		Println(fmt.Sprintf("Project: %s", p.projectName))
	fmt.Println()
}

// StartSection begins a new section.
func (p *Progress) StartSection(name string) *Section {
	section := &Section{
		Name:  name,
		Steps: make([]*Step, 0),
	}
	p.sections = append(p.sections, section)

	if !p.quiet && !p.jsonOutput {
		pterm.DefaultSection.WithLevel(2).Println(name)
	}

	return section
}

// AddStep adds a step to a section and starts it.
func (s *Section) AddStep(name string) *Step {
	step := &Step{
		Name:      name,
		Status:    StatusRunning,
		StartTime: time.Now(),
	}
	s.Steps = append(s.Steps, step)
	return step
}

// StartStep starts a step with a spinner.
func (p *Progress) StartStep(section *Section, name string) *Step {
	step := section.AddStep(name)
	p.currentStep = step

	if !p.quiet && !p.jsonOutput {
		spinner, _ := pterm.DefaultSpinner.Start(name)
		section.spinner = spinner
	}

	return step
}

// UpdateStep updates the current step message.
func (p *Progress) UpdateStep(message string) {
	if p.currentStep != nil {
		p.currentStep.Message = message
	}

	if !p.quiet && !p.jsonOutput {
		for _, section := range p.sections {
			if section.spinner != nil && section.spinner.IsActive {
				section.spinner.UpdateText(message)
			}
		}
	}
}

// SuccessStep marks a step as successful.
func (p *Progress) SuccessStep(section *Section, step *Step) {
	step.Status = StatusSuccess
	step.Duration = time.Since(step.StartTime)

	if !p.quiet && !p.jsonOutput && section.spinner != nil {
		section.spinner.Success(fmt.Sprintf("%s [%s]", step.Name, formatDuration(step.Duration)))
	}
}

// FailStep marks a step as failed.
func (p *Progress) FailStep(section *Section, step *Step, err error) {
	step.Status = StatusFailed
	step.Duration = time.Since(step.StartTime)
	step.Message = err.Error()

	if !p.quiet && !p.jsonOutput && section.spinner != nil {
		section.spinner.Fail(fmt.Sprintf("%s - %s", step.Name, err.Error()))
	}
}

// WarningStep marks a step with a warning.
func (p *Progress) WarningStep(section *Section, step *Step, message string) {
	step.Status = StatusWarning
	step.Duration = time.Since(step.StartTime)
	step.Message = message

	if !p.quiet && !p.jsonOutput && section.spinner != nil {
		section.spinner.Warning(fmt.Sprintf("%s - %s", step.Name, message))
	}
}

// AddSubStep adds a sub-step to a step.
func (s *Step) AddSubStep(name string, status StepStatus) {
	subStep := &Step{
		Name:   name,
		Status: status,
	}
	s.SubSteps = append(s.SubSteps, subStep)
}

// ShowSubSteps displays sub-steps.
func (p *Progress) ShowSubSteps(step *Step) {
	if p.quiet || p.jsonOutput || len(step.SubSteps) == 0 {
		return
	}

	for _, sub := range step.SubSteps {
		icon := getStatusIcon(sub.Status)
		pterm.Printf("   %s %s\n", icon, sub.Name)
	}
}

// ShowValidation displays validation results.
func (p *Progress) ShowValidation(checks []ValidationCheck) {
	if p.quiet || p.jsonOutput {
		return
	}

	fmt.Println()
	pterm.DefaultSection.WithLevel(2).Println("Validation")

	for _, check := range checks {
		icon := getStatusIcon(StepStatus(check.Status))
		if check.Status == "passed" {
			pterm.Success.Printf("%s %s\n", icon, check.Name)
		} else if check.Status == "warning" {
			pterm.Warning.Printf("%s %s - %s\n", icon, check.Name, check.Message)
		} else {
			pterm.Error.Printf("%s %s - %s\n", icon, check.Name, check.Message)
		}
	}
}

// ShowError displays an error with suggestions.
func (p *Progress) ShowError(err error, suggestions []string) {
	if p.quiet || p.jsonOutput {
		return
	}

	fmt.Println()
	pterm.Error.Println(err.Error())

	if len(suggestions) > 0 {
		fmt.Println()
		pterm.Info.Println("💡 Suggestions:")
		for i, suggestion := range suggestions {
			pterm.Printf("   %d. %s\n", i+1, suggestion)
		}
	}
}

func getStatusIcon(status StepStatus) string {
	switch status {
	case StatusSuccess:
		return pterm.Green("✓")
	case StatusFailed:
		return pterm.Red("✗")
	case StatusWarning:
		return pterm.Yellow("⚠")
	case StatusRunning:
		return pterm.Cyan("●")
	case StatusSkipped:
		return pterm.Gray("○")
	default:
		if status == "passed" {
			return pterm.Green("✓")
		}
		return pterm.Gray("○")
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// ValidationCheck represents a validation check result.
type ValidationCheck struct {
	Name    string
	Check   string
	Status  string
	Message string
}

