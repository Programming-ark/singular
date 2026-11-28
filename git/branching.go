package git

import (
	"crypto/sha256"
	"errors"
	"fmt"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// ManagedBranchRef derives the physical ref from a logical branch name.
func ManagedBranchRef(logicalName string) (plumbing.ReferenceName, error) {
	logicalRef := plumbing.NewBranchReferenceName(logicalName)
	if logicalName == "" {
		return "", fmt.Errorf("branch name is required")
	}
	if err := logicalRef.Validate(); err != nil {
		return "", fmt.Errorf("invalid logical branch %q: %w", logicalName, err)
	}
	digest := sha256.Sum256([]byte(logicalName))
	return plumbing.NewBranchReferenceName(fmt.Sprintf("sg-%s-%x", logicalName, digest[:5])), nil
}

// CreateManagedBranch creates a physical branch at the supplied commit.
func CreateManagedBranch(repo *gogit.Repository, logicalName string, commit plumbing.Hash) error {
	ref, err := ManagedBranchRef(logicalName)
	if err != nil {
		return err
	}
	if _, err := repo.CommitObject(commit); err != nil {
		return fmt.Errorf("resolve branch point %s: %w", commit, err)
	}
	existing, err := repo.Reference(ref, false)
	switch {
	case err == nil:
		if existing.Hash() == commit {
			return nil
		}
		return fmt.Errorf("managed branch for %q already exists", logicalName)
	case !errors.Is(err, plumbing.ErrReferenceNotFound):
		return err
	}
	if err := repo.Storer.SetReference(plumbing.NewHashReference(ref, commit)); err != nil {
		return fmt.Errorf("create managed branch for %q: %w", logicalName, err)
	}
	return nil
}
