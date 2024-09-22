package github

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	//"github.com/go-git/go-git/v5/plumbing/transport/http"     // for auth
)

// Clone repo
func Clone(url string, directory string) (*git.Repository, error) {
    fmt.Printf("git clone %s %s --recursive\n", url, directory)

    //authMethod, err := ssh.NewSSHAgentAuth("git")

    repo, err := git.PlainClone(directory, false, &git.CloneOptions{
        URL:               url,
        RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
        // Set the SSH client for authentication
        /*
        Auth: &http.BasicAuth{
			Username: "jackbekket", // yes, this can be anything except an empty string
			Password: "gh_token here",
		},
        */
    })

    return repo, err
}

 // Create a new branch and switch to it
 func CreateAndSwitch(r *git.Repository, branch string) error {
    w, err := r.Worktree()
    if err != nil {
        return err
    }

    branchRefName := plumbing.NewBranchReferenceName(branch)

    // Create and checkout to the new branch
    err = w.Checkout(&git.CheckoutOptions{
        Branch: branchRefName,
        Create: true,
        Force:  true,
    })
    if err != nil {
        return err
    } else {
        return nil
    }
}



// Push the current branch to the remote repository
func Push(r *git.Repository, branch string) error {
    /*
	remote, err := r.Remote("origin")
    if err != nil {
        return err
    }
	*/

    refSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)
    pushOptions := &git.PushOptions{
        RefSpecs: []config.RefSpec{config.RefSpec(refSpec)},
    }

    err := r.Push(pushOptions)
    if err != nil {
        return err
    } else {

    return nil
    }
}


func SwitchToMaster(r *git.Repository) error {
    w, err := r.Worktree()
    if err != nil {
        return err
    }

    masterBranch := "refs/heads/master"
    branchCoOpts := git.CheckoutOptions{
        Branch: plumbing.ReferenceName(masterBranch),
        Force:  true,
    }

    return w.Checkout(&branchCoOpts)
}
