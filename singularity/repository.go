package singularity

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	virtualgit "github.com/Programming-ark/singularity/git"
)

// InitializeRepository writes the policy and creates its managed Git repository.
func InitializeRepository(configPath string, policy RepositoryPolicy) error {
	dir := filepath.Dir(configPath)
	repoPath := filepath.Join(dir, "repo")
	if _, err := os.Lstat(repoPath); err == nil {
		return fmt.Errorf("virtual repository path already exists at %s", repoPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check virtual repository path: %w", err)
	}
	if err := WritePolicy(configPath, policy); err != nil {
		return err
	}
	if err := initializeVirtualRepo(dir, repoPath, policy); err != nil {
		return errors.Join(err, os.Remove(configPath))
	}
	return nil
}

func initializeVirtualRepo(dir, repoPath string, policy RepositoryPolicy) error {
	staging, err := os.MkdirTemp(dir, "repo-init-")
	if err != nil {
		return fmt.Errorf("stage virtual repository: %w", err)
	}
	defer os.RemoveAll(staging)

	repo, commit, err := virtualgit.InitVirtualRepo(staging, strings.TrimSpace(policy.RootBranch))
	if err != nil {
		return err
	}
	branches := append(slices.Sorted(maps.Keys(policy.Domains)), policy.Protected...)
	for _, name := range branches {
		if err := virtualgit.CreateManagedBranch(repo, name, commit); err != nil {
			return err
		}
	}
	if err := os.Rename(staging, repoPath); err != nil {
		return fmt.Errorf("publish virtual repository: %w", err)
	}
	return nil
}
