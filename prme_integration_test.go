//+build integration

package prme_test

import (
	"os"
	"prme"
	"testing"
)

func TestCreateFullPullRequestIntegration(t *testing.T) {
	t.Parallel()

	repo := "ivanfetch/ghapitest"
	PRTitle := "review-integration"
	PRBody := "Review of full repository"
	PRBaseBranch := "fullpullrequest-orphan-integration"
	PRHeadBranch := "fullpullrequest-review-integration"

	g, err := prme.New(os.Getenv("GH_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}

	PRID, err := g.CreateFullPullRequest(repo, "main", PRTitle, PRBody, PRBaseBranch, PRHeadBranch)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("created pull request %d", PRID)

	err = g.ClosePullRequest(repo, PRID)
	if err != nil {
		t.Fatalf("while cleaning up pull request: %v", err)
	}
	err = g.DeleteBranch(repo, PRHeadBranch)
	if err != nil {
		t.Fatalf("while cleaning up head branch %q: %v", PRHeadBranch, err)
	}
	err = g.DeleteBranch(repo, PRBaseBranch)
	if err != nil {
		t.Fatalf("while cleaning up base branch %q: %v", PRBaseBranch, err)
	}
}
