package singularity_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Programming-ark/singularity/singularity"
	"go.yaml.in/yaml/v3"
)

func TestWritePolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), singularity.PolicyPath)
	policy := singularity.RepositoryPolicy{
		RootBranch:     "main",
		RepositoryType: "Monorepo",
		Protected:      []string{"release"},
		Domains: map[string]singularity.DomainPolicy{
			"users": {Paths: []string{"users/**", "shared/users/**"}},
		},
		Handoff: &singularity.HandoffPolicy{ApprovalRequired: true, ExpirationDays: 30},
	}
	policy.CrossDomain.Strategy = "handoff"
	policy.DependencyGovernance.Enabled = true
	if err := singularity.WritePolicy(path, policy); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded singularity.RepositoryPolicy
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
	if err := singularity.WritePolicy(path, policy); !errors.Is(err, os.ErrExist) {
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
