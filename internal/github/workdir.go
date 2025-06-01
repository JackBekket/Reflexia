package github

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/rs/zerolog/log"
)

func GithubWorkdir(
	workdir,
	repoURL,
	branch,
	githubUsername,
	githubToken string,
) (string, *git.Repository, string, func(), error) {
	u, err := url.ParseRequestURI(repoURL)
	if err != nil {
		return "", nil, "", nil, err
	}

	sPath := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	cdAfter := ""
	if len(sPath) < 2 {
		return "", nil, "", nil, errors.New("github repository url does not have at least two path elements")
	} else if len(sPath) > 2 {
		// set githubLink to a repository only
		u.Path = strings.Join(sPath[:2], "/")
		repoURL = u.String()

		if sPath[2] != "tree" {
			return "", nil, "", nil, errors.New("extended github repository url should have 'tree' route after repository name")
		}
		if len(sPath) < 4 {
			return "", nil, "", nil, errors.New("extended github repository url should have branch name after tree route")
		}
		// do not override branch set from the flag
		if branch == "" {
			branch = sPath[3]
		}
		cdAfter = strings.Join(sPath[4:], "/")
	}

	tempDirEl := []string{workdir, "temp"}
	tempDirEl = append(tempDirEl, sPath[:2]...)

	if branch == "" {
		rem := git.NewRemote(memory.NewStorage(), &gitConfig.RemoteConfig{
			Name: "origin",
			URLs: []string{repoURL},
		})
		refs, err := rem.List(&git.ListOptions{})
		if err != nil {
			return "", nil, "", nil, err
		}

		var defaultBranchRef *plumbing.Reference
		for _, ref := range refs {
			if ref.Name() == "HEAD" {
				defaultBranchRef = ref
				break
			}
		}

		if defaultBranchRef == nil {
			return "", nil, "", nil, fmt.Errorf("HEAD reference not found in ls-remote output")
		}

		if defaultBranchRef.Type() == plumbing.SymbolicReference {
			targetRefName := defaultBranchRef.Target()
			for _, ref := range refs {
				if ref.Name() == targetRefName {
					defaultBranchRef = ref
					break
				}
			}
		}

		branch = strings.Replace(
			defaultBranchRef.Name().String(),
			"refs/heads/",
			"", 1,
		)
	}
	autodocBranch := fmt.Sprintf(BranchFormat, branch)
	autodocBranchRefName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", autodocBranch))

	tempDirEl = append(tempDirEl, branch)
	tempDir := filepath.Join(tempDirEl...)

	workdir = tempDir
	var repo *git.Repository

	cancelFunc := func() {
		if err := os.RemoveAll(workdir); err != nil {
			log.Warn().Err(err).Msg("cancel func remove workdir")
		}
	}

	if _, err := os.Stat(workdir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(workdir, os.FileMode(0755)); err != nil {
				return "", nil, "", cancelFunc, err
			}

			cloneOptions := git.CloneOptions{
				URL:               repoURL,
				Depth:             1,
				SingleBranch:      true,
				RecurseSubmodules: git.NoRecurseSubmodules,
				ShallowSubmodules: true,
				ReferenceName:     plumbing.ReferenceName(branch),
			}
			if githubUsername != "" && githubToken != "" {
				cloneOptions.Auth = &http.BasicAuth{
					Username: githubUsername,
					Password: githubToken,
				}
			}

			if repo, err = git.PlainClone(workdir, false, &cloneOptions); err != nil {
				return "", nil, "", cancelFunc, err
			}

			wt, err := repo.Worktree()
			if err != nil {
				return "", nil, "", cancelFunc, fmt.Errorf("get worktree: %w", err)
			}

			_, refErr := repo.Reference(autodocBranchRefName, false)
			err = wt.Checkout(&git.CheckoutOptions{
				Branch: autodocBranchRefName,
				Create: refErr != nil,
			})
			if err != nil {
				return "", nil, "", cancelFunc, fmt.Errorf("checkout new branch '%s': %w", autodocBranch, err)
			}

		} else {
			return "", nil, "", cancelFunc, err
		}
	}

	if repo == nil {
		repo, err = git.PlainOpen(workdir)
		if err != nil {
			return "", nil, "", cancelFunc, err
		}

		wt, err := repo.Worktree()
		if err != nil {
			return "", nil, "", cancelFunc, fmt.Errorf("failed to get worktree: %w", err)
		}

		err = repo.Fetch(&git.FetchOptions{
			Depth: 1,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return "", nil, "", cancelFunc, fmt.Errorf("git fetch error: %w", err)
		}

		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Force:  true,
		})
		if err != nil {
			return "", nil, "", cancelFunc, fmt.Errorf("failed to checkout base branch '%s': %w", branch, err)
		}

		pullOptions := git.PullOptions{
			Depth:             1,
			SingleBranch:      true,
			RecurseSubmodules: git.NoRecurseSubmodules,
		}
		if githubUsername != "" && githubToken != "" {
			pullOptions.Auth = &http.BasicAuth{
				Username: githubUsername,
				Password: githubToken,
			}
		}
		err = wt.Pull(&pullOptions)
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return "", nil, "", cancelFunc, fmt.Errorf("failed to pull base branch '%s': %w", branch, err)
		}

		_, refErr := repo.Reference(autodocBranchRefName, false)
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: autodocBranchRefName,
			Create: refErr != nil,
			Force:  true,
		})
		if err != nil {
			return "", nil, "", cancelFunc, fmt.Errorf("checkout new branch '%s': %w", autodocBranch, err)
		}
	}

	workdir = filepath.Join(workdir, cdAfter)

	return workdir, repo, branch, cancelFunc, nil
}
