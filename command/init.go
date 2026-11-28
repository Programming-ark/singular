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
	if !proceed {
		return nil
	}
	if err := writePolicy(policyPath, wizard.policy); err != nil {
		return err
	}
	fmt.Fprintf(out, "Created %s\n", policyPath)
	return nil
}

type initWizard struct {
	out     io.Writer
	history promptHistory
	policy  repositoryPolicy
}

func (w *initWizard) configure(advanced bool) error {
	policy := &w.policy
	policy.RootBranch = "main"
	policy.Domains = make(map[string]domainPolicy)
	rootPrompt := requiredInput("What is the repository root branch?", &policy.RootBranch)
	if !advanced {
		if err := w.run(rootPrompt, choice("What type of repository is this?", &policy.RepositoryType,
			"Monolith", "Monorepo", "Microservices", "Library")); err != nil {
			return err
		}
		if err := w.domains(false); err != nil {
			return err
		}
		consolidate := true
		if err := w.run(
			confirmation("Require feature branches to merge through their domain branch?", &consolidate),
			crossDomainChoice("How should cross-domain changes be handled?", "Create Handoff", &policy.CrossDomain.Strategy),
		); err != nil {
			return err
		}
		for name, domain := range policy.Domains {
			domain.ConsolidationRequired = consolidate
			policy.Domains[name] = domain
		}
		if policy.CrossDomain.Strategy == "handoff" {
			policy.Handoff = &handoffPolicy{ApprovalRequired: true}
		}
		return nil
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
	for _, branch := range additional {
		if branch != strings.TrimSpace(policy.RootBranch) {
			policy.Protected = append(policy.Protected, branch)
		}
	}
	if err := w.domains(true); err != nil {
		return err
	}
	if err := w.handoffs(); err != nil {
		return err
	}
	return w.run(confirmation("Enable dependency governance?", &policy.DependencyGovernance.Enabled))
}

func (w *initWizard) domains(advanced bool) error {
	for {
		var name string
		namePrompt := answeredPrompt{
			label: "Domain name",
			field: huh.NewInput().Title("Domain name").Value(&name).Validate(func(value string) error {
				value = strings.TrimSpace(value)
				if err := requireValue(value); err != nil {
					return err
				}
				if _, exists := w.policy.Domains[value]; exists {
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
				Value(&paths).Validate(func(value string) error {
				if len(splitList(value)) == 0 {
					return fmt.Errorf("enter at least one owned path")
				}
				return nil
			}),
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
		if err := w.run(settings...); err != nil {
			return err
		}
		w.policy.Domains[name] = domainPolicy{
			Paths:                 splitList(paths),
			ConsolidationRequired: consolidate,
			DirectMerge:           directMerge,
			TrustedDomains:        splitList(trust),
		}
		if !more {
			return nil
		}
	}
}

func (w *initWizard) handoffs() error {
	strategy := &w.policy.CrossDomain.Strategy
	if err := w.run(crossDomainChoice("Cross-domain strategy?", "Handoff", strategy)); err != nil {
		return err
	}
	if *strategy != "handoff" {
		return nil
	}
	handoff := &handoffPolicy{ApprovalRequired: true}
	w.policy.Handoff = handoff
	expire := true
	if err := w.run(
		confirmation("Require approval before handoff merge?", &handoff.ApprovalRequired),
		confirmation("Expire handoffs automatically?", &expire),
	); err != nil || !expire {
		return err
	}
	days := "30"
	if err := w.run(answeredPrompt{
		label: "Handoff expiration in days",
		field: huh.NewInput().Title("Handoff expiration in days?").Value(&days).Validate(func(value string) error {
			count, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || count <= 0 {
				return fmt.Errorf("enter a positive whole number")
			}
			return nil
		}),
	}); err != nil {
		return err
	}
	handoff.ExpirationDays, _ = strconv.Atoi(strings.TrimSpace(days))
	return nil
}

func crossDomainChoice(title, handoffLabel string, value *string) answeredPrompt {
	return answeredPrompt{label: title, field: huh.NewSelect[string]().Title(title).
		Options(huh.NewOption(handoffLabel, "handoff"), huh.NewOption("Reject Merge", "reject_merge")).Value(value)}
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
