package command

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	virtualgit "github.com/Programming-ark/singularity/git"
	gogit "github.com/go-git/go-git/v5"
)

func TestInitializeRepository(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), policyPath)
	policy := repositoryPolicy{
		RootBranch: "main",
		Domains:    map[string]domainPolicy{"users": {Paths: []string{"users/**"}}},
		Protected:  []string{"release"},
	}
	if err := initializeRepository(configPath, policy); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatal(err)
	}
	repo, err := gogit.PlainOpen(filepath.Join(filepath.Dir(configPath), "repo"))
	if err != nil {
		t.Fatal(err)
	}
	head, err := repo.Head()
	if err != nil {
		t.Fatal(err)
	}
	root, err := virtualgit.ManagedBranchRef("main")
	if err != nil || head.Name() != root {
		t.Fatalf("unexpected HEAD %s, want %s (error: %v)", head.Name(), root, err)
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		t.Fatal(err)
	}
	tree, err := commit.Tree()
	if err != nil {
		t.Fatal(err)
	}
	if commit.NumParents() != 0 || len(tree.Entries) != 0 || commit.Message != "Initialize Singularity governance" {
		t.Fatalf("unexpected initial commit: %s with %d tree entries", commit, len(tree.Entries))
	}
	for _, logical := range []string{"users", "release"} {
		name, err := virtualgit.ManagedBranchRef(logical)
		if err != nil {
			t.Fatal(err)
		}
		ref, err := repo.Reference(name, false)
		if err != nil {
			t.Fatal(err)
		}
		if ref.Hash() != head.Hash() {
			t.Fatalf("%s points to %s, want %s", name, ref.Hash(), head.Hash())
		}
	}
}

func TestInitializeRepositoryRollsBackInvalidBranch(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), policyPath)
	policy := repositoryPolicy{RootBranch: "main", Protected: []string{"invalid branch"}}
	if err := initializeRepository(configPath, policy); err == nil {
		t.Fatal("expected invalid branch to fail initialization")
	}
	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("policy should not exist after failure, got %v", err)
	}
	entries, err := os.ReadDir(filepath.Dir(configPath))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("unexpected files after failed initialization: %v", entries)
	}
}
