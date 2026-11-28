package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

const policyPath = ".singularity/policy.yaml"

type repositoryPolicy struct {
	RootBranch     string                  `yaml:"root_branch"`
	RepositoryType string                  `yaml:"repository_type,omitempty"`
	Protected      []string                `yaml:"protected"`
	Domains        map[string]domainPolicy `yaml:"domains"`
	CrossDomain    struct {
		Strategy string `yaml:"strategy"`
	} `yaml:"cross_domain"`
	Handoff              *handoffPolicy `yaml:"handoff,omitempty"`
	DependencyGovernance struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"dependency_governance"`
}

type domainPolicy struct {
	Paths                 []string `yaml:"paths"`
	ConsolidationRequired bool     `yaml:"consolidation_required"`
	DirectMerge           bool     `yaml:"direct_merge"`
	TrustedDomains        []string `yaml:"trusted_domains,omitempty"`
}

type handoffPolicy struct {
	ApprovalRequired bool `yaml:"approval_required"`
	ExpirationDays   int  `yaml:"expiration_days,omitempty"`
}

func splitList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" && !slices.Contains(items, item) {
			items = append(items, item)
		}
	}
	return items
}

func writePolicy(path string, policy repositoryPolicy) error {
	policy.RootBranch = strings.TrimSpace(policy.RootBranch)
	policy.RepositoryType = strings.ToLower(policy.RepositoryType)
	policy.Protected = append([]string{policy.RootBranch}, policy.Protected...)
	data, err := yaml.Marshal(policy)
	if err != nil {
		return fmt.Errorf("encode policy: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create policy directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("policy already exists at %s: %w", path, err)
	}
	if err != nil {
		return fmt.Errorf("create policy: %w", err)
	}
	_, writeErr := file.Write(data)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return fmt.Errorf("write policy: %w", errors.Join(err, os.Remove(path)))
	}
	return nil
}
