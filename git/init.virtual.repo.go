package git

import (
	"fmt"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// InitVirtualRepo creates the hidden Git repository and commits its empty tree.
func InitVirtualRepo(path, root string) (*gogit.Repository, plumbing.Hash, error) {
	ref, err := ManagedBranchRef(root)
	if err != nil {
		return nil, plumbing.ZeroHash, err
	}
	repo, err := gogit.PlainInitWithOptions(path, &gogit.PlainInitOptions{
		InitOptions: gogit.InitOptions{DefaultBranch: ref},
	})
	if err != nil {
		return nil, plumbing.ZeroHash, fmt.Errorf("initialize virtual repository: %w", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, plumbing.ZeroHash, fmt.Errorf("open virtual worktree: %w", err)
	}
	identity := &object.Signature{Name: "Singularity", Email: "singularity@localhost", When: time.Now().UTC()}
	commit, err := worktree.Commit("Initialize Singularity governance", &gogit.CommitOptions{
		AllowEmptyCommits: true,
		Author:            identity,
		Committer:         identity,
	})
	if err != nil {
		return nil, plumbing.ZeroHash, fmt.Errorf("create initial Singularity commit: %w", err)
	}
	return repo, commit, nil
}
