package command

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestWritePolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), policyPath)
	policy := repositoryPolicy{
		RootBranch:     "main",
		RepositoryType: "Monorepo",
		Protected:      []string{"release"},
		Domains: map[string]domainPolicy{
			"users": {Paths: splitList(" users/**, shared/users/**, , users/** ")},
		},
		Handoff: &handoffPolicy{ApprovalRequired: true, ExpirationDays: 30},
	}
	policy.CrossDomain.Strategy = "handoff"
	policy.DependencyGovernance.Enabled = true
	if err := writePolicy(path, policy); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded repositoryPolicy
	if err := yaml.Unmarshal(original, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.RootBranch != "main" || decoded.RepositoryType != "monorepo" ||
		!slices.Equal(decoded.Protected, []string{"main", "release"}) ||
		!slices.Equal(decoded.Domains["users"].Paths, []string{"users/**", "shared/users/**"}) ||
		decoded.CrossDomain.Strategy != "handoff" || decoded.Handoff == nil ||
		!decoded.Handoff.ApprovalRequired || decoded.Handoff.ExpirationDays != 30 ||
		!decoded.DependencyGovernance.Enabled {
		t.Fatalf("unexpected policy:\n%s", original)
	}

	policy.RootBranch = "different"
	if err := writePolicy(path, policy); !errors.Is(err, os.ErrExist) {
		t.Fatalf("second write returned %v, want os.ErrExist", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, original) {
		t.Fatal("second write modified the existing policy")
	}
}
