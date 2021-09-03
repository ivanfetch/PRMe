//+build integration

package prme_test

import (
	"os"
	"prme"
	"testing"
)

func TestCreateFullPullRequestIntegration(t *testing.T) {
	t.Parallel()

	PRTitle := "review-integration"
	PRBody := "Review of full repository"
	PRBaseBranch := "fullpullrequest-orphan-integration"
	PRHeadBranch := "fullpullrequest-review-integration"

	r, err := prme.NewRepo("ivanfetch/ghapitest", os.Getenv("GH_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}

	PRID, err := r.CreateFullPullRequest("main", PRTitle, PRBody, PRBaseBranch, PRHeadBranch)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("created pull request %d", PRID)

	err = r.ClosePullRequest(PRID)
	if err != nil {
		t.Fatalf("while cleaning up pull request: %v", err)
	}
	err = r.DeleteBranch(PRHeadBranch)
	if err != nil {
		t.Fatalf("while cleaning up head branch %q: %v", PRHeadBranch, err)
	}
	err = r.DeleteBranch(PRBaseBranch)
	if err != nil {
		t.Fatalf("while cleaning up base branch %q: %v", PRBaseBranch, err)
	}
}
