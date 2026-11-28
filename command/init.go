package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/huh"
)

func Initialize(flags []string, out io.Writer) error {
	fmt.Fprintln(out, "Welcome to Singularity.")
	fmt.Fprintln(out, "This wizard collects the settings for your repository's governance policy.")

	answers := struct {
		ProtectedBranch    string
		AdditionalBranches []string
		RequirePR          bool
		Ownership          string
		DomainPaths        string
		TeamPaths          string
		UnmanagedFiles     string
		Permissions        string
		BranchMappings     string
		Enforcement        string
	}{
		ProtectedBranch: "main",
		Ownership:       "domain",
		UnmanagedFiles:  "reject",
		Permissions:     "pattern",
		Enforcement:     "block",
	}

	required := func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("enter a value to continue")
		}
		return nil
	}

	var history promptHistory
	prompt := func(label string, field huh.Field, hidden ...func() bool) *huh.Group {
		group := huh.NewGroup(field)
		entry := answeredPrompt{label: label, field: field}
		if len(hidden) > 0 {
			entry.hidden = hidden[0]
			group.WithHideFunc(entry.hidden)
		}
		history = append(history, entry)
		return group
	}

	form := huh.NewForm(
		prompt("Primary protected branch",
			huh.NewInput().
				Title("What is your primary protected branch?").
				Value(&answers.ProtectedBranch).
				Validate(required),
		),
		prompt("Additional protected branches",
			huh.NewMultiSelect[string]().
				Title("Protect additional branches?").
				Description("Select any branches to protect or continue with no selections.").
				Options(
					huh.NewOption("release", "release"),
					huh.NewOption("hotfix", "hotfix"),
					huh.NewOption("production", "production"),
				).
				Value(&answers.AdditionalBranches),
		),
		prompt("Repository ownership",
			huh.NewSelect[string]().
				Title("How should repository ownership be organized?").
				Options(
					huh.NewOption("Domain-based", "domain"),
					huh.NewOption("Team-based", "team"),
					huh.NewOption("Open repository", "open"),
				).
				Value(&answers.Ownership),
		),
		prompt("Domain paths",
			huh.NewText().
				Title("Which domains own which paths?").
				Description("Enter one domain per line, for example: billing = billing/**").
				Value(&answers.DomainPaths).
				Validate(required),
			func() bool { return answers.Ownership != "domain" },
		),
		prompt("Team paths",
			huh.NewText().
				Title("Which teams own which paths?").
				Description("Enter one team per line, for example: platform = infra/**").
				Value(&answers.TeamPaths).
				Validate(required),
			func() bool { return answers.Ownership != "team" },
		),
		prompt("Unmanaged files",
			huh.NewSelect[string]().
				Title("What should happen when a file has no owner?").
				Options(
					huh.NewOption("Reject merge", "reject"),
					huh.NewOption("Allow merge", "allow"),
				).
				Value(&answers.UnmanagedFiles),
			func() bool { return answers.Ownership == "open" },
		),
		prompt("Branch permissions",
			huh.NewSelect[string]().
				Title("How should branches receive permissions?").
				Options(
					huh.NewOption("By branch pattern", "pattern"),
					huh.NewOption("Allow any branch", "any"),
				).
				Value(&answers.Permissions),
			func() bool { return answers.Ownership == "open" },
		),
		prompt("Branch mappings",
			huh.NewText().
				Title("Which owners can each branch pattern modify?").
				Description("Enter one mapping per line, for example: feature/billing-* = billing").
				Value(&answers.BranchMappings).
				Validate(required),
			func() bool { return answers.Ownership == "open" || answers.Permissions != "pattern" },
		),
		prompt("Enforcement",
			huh.NewSelect[string]().
				Title("How should Singularity enforce policies?").
				Options(
					huh.NewOption("Block merges", "block"),
					huh.NewOption("Warn only", "warn"),
				).
				Value(&answers.Enforcement),
		),
		prompt("",
			huh.NewNote().
				Title("Review your policy settings").
				DescriptionFunc(func() string {
					branches := append([]string{answers.ProtectedBranch}, answers.AdditionalBranches...)
					summary := fmt.Sprintf("Protected branches: %s\nRequire pull requests: %t\nOwnership: %s",
						strings.Join(branches, ", "), answers.RequirePR, answers.Ownership)
					switch answers.Ownership {
					case "domain":
						summary += "\nDomain paths:\n" + answers.DomainPaths
					case "team":
						summary += "\nTeam paths:\n" + answers.TeamPaths
					}
					if answers.Ownership != "open" {
						summary += fmt.Sprintf("\nUnmanaged files: %s\nBranch Permissions: %s", answers.UnmanagedFiles, answers.Permissions)
						if answers.Permissions == "pattern" {
							summary += "\nBranch mappings:\n" + answers.BranchMappings
						}
					}
					return summary + "\nEnforcement: " + answers.Enforcement
				}, &answers).
				Next(true),
		),
		prompt("Generate singularity.yaml",
			huh.NewConfirm().
				Title("Generate singularity.yaml?").
				Description("This wizard currently only collects answers."),
		),
	).WithLayout(history).WithOutput(out)
	if err := form.Run(); err != nil {
		return err
	}
	fmt.Fprint(out, history.answersBefore(nil))
	return nil
}

type answeredPrompt struct {
	label  string
	field  huh.Field
	hidden func() bool
}

type promptHistory []answeredPrompt

func (history promptHistory) View(form *huh.Form) string {
	return history.answersBefore(form.GetFocusedField()) + huh.LayoutDefault.View(form)
}

func (history promptHistory) GroupWidth(form *huh.Form, group *huh.Group, width int) int {
	return huh.LayoutDefault.GroupWidth(form, group, width)
}

func (history promptHistory) answersBefore(active huh.Field) string {
	var transcript strings.Builder
	for _, prompt := range history {
		if prompt.field == active {
			break
		}
		if prompt.label == "" || (prompt.hidden != nil && prompt.hidden()) {
			continue
		}
		var value string
		switch answer := prompt.field.GetValue().(type) {
		case bool:
			value = "No"
			if answer {
				value = "Yes"
			}
		case []string:
			value = strings.Join(answer, ", ")
			if value == "" {
				value = "None"
			}
		default:
			value = fmt.Sprint(answer)
		}
		value = strings.ReplaceAll(strings.TrimSpace(value), "\n", "\n  ")
		fmt.Fprintf(&transcript, "✓ %s: %s\n", prompt.label, value)
	}
	if transcript.Len() > 0 && active != nil {
		transcript.WriteByte('\n')
	}
	return transcript.String()
}
