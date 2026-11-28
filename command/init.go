package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

func Initialize(flags []string, out io.Writer) error {
	options := flag.NewFlagSet("singularity init", flag.ContinueOnError)
	options.SetOutput(out)
	advanced := options.Bool("advanced", false, "configure detailed repository governance")
	if err := options.Parse(flags); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if options.NArg() > 0 {
		return fmt.Errorf("unexpected argument for init: %q", options.Arg(0))
	}

	fmt.Fprintln(out, "Welcome to Singularity.")
	fmt.Fprintln(out, "This wizard collects the settings for your repository's governance policy.")

	wizard := initWizard{out: out}
	if err := wizard.configure(*advanced); err != nil {
		return err
	}

	title := "Proceed?"
	if *advanced {
		title = "Generate configuration?"
	}
	proceed := true
	review := answeredPrompt{
		label: title,
		field: huh.NewConfirm().Title(title).
			Description("Review your governance settings above. Managed physical branches are always enabled.").
			Value(&proceed),
	}
	if err := wizard.run(review); err != nil {
		return err
	}

	fmt.Fprint(out, wizard.history.answersBefore(nil))
	if proceed {
		fmt.Fprintln(out, "Governance settings collected.")
	}
	return nil
}

type initWizard struct {
	out     io.Writer
	history promptHistory
}

func (w *initWizard) configure(advanced bool) error {
	root := "main"
	rootPrompt := requiredInput("What is the repository root branch?", &root)
	if !advanced {
		var repositoryType string
		if err := w.run(rootPrompt, choice("What type of repository is this?", &repositoryType,
			"Monolith", "Monorepo", "Microservices", "Library")); err != nil {
			return err
		}
		if err := w.domains(false); err != nil {
			return err
		}
		consolidate := true
		var crossDomain string
		return w.run(
			confirmation("Require feature branches to merge through their domain branch?", &consolidate),
			choice("How should cross-domain changes be handled?", &crossDomain, "Create Handoff", "Reject Merge"),
		)
	}

	var additional []string
	if err := w.run(
		rootPrompt,
		answeredPrompt{
			label: "Additional protected branches?",
			field: huh.NewMultiSelect[string]().
				Title("Additional protected branches?").
				Description("Select any additional branches or continue with no selections.").
				Options(huh.NewOptions("release", "hotfix", "production")...).
				Value(&additional),
		},
	); err != nil {
		return err
	}
	if err := w.domains(true); err != nil {
		return err
	}
	if err := w.handoffs(); err != nil {
		return err
	}
	dependencies := false
	return w.run(confirmation("Enable dependency governance?", &dependencies))
}

func (w *initWizard) domains(advanced bool) error {
	names := make(map[string]bool)
	for {
		var name string
		namePrompt := answeredPrompt{
			label: "Domain name",
			field: huh.NewInput().Title("Domain name").Value(&name).Validate(func(value string) error {
				value = strings.TrimSpace(value)
				if err := requireValue(value); err != nil {
					return err
				}
				if names[value] {
					return fmt.Errorf("domain %q already exists", value)
				}
				return nil
			}),
		}
		if err := w.run(namePrompt); err != nil {
			return err
		}
		name = strings.TrimSpace(name)
		paths := name + "/**"
		pathPrompt := answeredPrompt{
			label: "Owned paths",
			field: huh.NewInput().Title("Owned paths").
				Description("Press Enter to accept the suggested path, or edit it. Separate multiple paths with commas.").
				Value(&paths).Validate(requireValue),
		}
		prompts := []answeredPrompt{pathPrompt}
		consolidate, directMerge := true, false
		if advanced {
			prompts = append(prompts,
				confirmation("Must child branches merge through this domain before reaching the repository root?", &consolidate),
				confirmation("Allow direct merges to the repository root branch?", &directMerge),
			)
		}
		if err := w.run(prompts...); err != nil {
			return err
		}
		names[name] = true

		var trust string
		var settings []answeredPrompt
		if advanced {
			settings = append(settings,
				answeredPrompt{
					label: "Which domains may submit handoffs into this domain without additional approval?",
					field: huh.NewInput().Title("Which domains may submit handoffs into this domain without additional approval?").
						Description("Separate domain names with commas. Leave blank for none.").Value(&trust),
				},
			)
		}
		more := false
		settings = append(settings, confirmation("Create another domain?", &more))
		if err := w.run(settings...); err != nil || !more {
			return err
		}
	}
}

func (w *initWizard) handoffs() error {
	var strategy string
	if err := w.run(choice("Cross-domain strategy?", &strategy, "Handoff", "Reject Merge")); err != nil {
		return err
	}
	if strategy != "Handoff" {
		return nil
	}
	approval, expire := true, true
	if err := w.run(
		confirmation("Require approval before handoff merge?", &approval),
		confirmation("Expire handoffs automatically?", &expire),
	); err != nil || !expire {
		return err
	}
	days := "30"
	return w.run(answeredPrompt{
		label: "Handoff expiration in days",
		field: huh.NewInput().Title("Handoff expiration in days?").Value(&days).Validate(func(value string) error {
			count, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || count <= 0 {
				return fmt.Errorf("enter a positive whole number")
			}
			return nil
		}),
	})
}

func (w *initWizard) run(prompts ...answeredPrompt) error {
	groups := make([]*huh.Group, 0, len(prompts))
	for _, prompt := range prompts {
		groups = append(groups, huh.NewGroup(prompt.field))
	}
	w.history = append(w.history, prompts...)
	return huh.NewForm(groups...).WithLayout(w.history).WithOutput(w.out).Run()
}

func requiredInput(title string, value *string) answeredPrompt {
	return answeredPrompt{label: title, field: huh.NewInput().Title(title).Value(value).Validate(requireValue)}
}

func requireValue(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("enter a value to continue")
	}
	return nil
}

func confirmation(title string, value *bool) answeredPrompt {
	return answeredPrompt{label: title, field: huh.NewConfirm().Title(title).Value(value)}
}

func choice(title string, value *string, choices ...string) answeredPrompt {
	return answeredPrompt{label: title, field: huh.NewSelect[string]().Title(title).
		Options(huh.NewOptions(choices...)...).Value(value)}
}

type answeredPrompt struct {
	label string
	field huh.Field
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
