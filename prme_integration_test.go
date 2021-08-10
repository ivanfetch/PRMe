//+build integration

package prme_test

import (
	"os"
	"prme"
	"testing"
)

func TestConnectToGithubAPIIntegration(t *testing.T) {
	t.Parallel()

	g, err := prme.New(os.Getenv("GH_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}

	got, err := g.GetRepoID("github/docs")
	if err != nil {
		t.Fatal(err)
	}

	var want int64 = 189621607 // ID for github/docs repo
	if got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}

func TestCreateOrphanBranchIntegration(t *testing.T) {
	t.Parallel()

	repo := "ivanfetch/ghapitest"
	branch := "test2"

	g, err := prme.New(os.Getenv("GH_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}

	ok, err := g.BranchExists(repo, branch)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("Branch %q already exists on repository %q", branch, repo)
	}

	err = g.CreateOrphanBranch(repo, branch)
	if err != nil {
		t.Fatal(err)
	}

	ok, err = g.BranchExists(repo, branch)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("Branch %q does not actually exists on repository %q", branch, repo)
	}
}

func TestCreateEmptyTreeCommitIntegration(t *testing.T) {
	t.Parallel()

	repo := "ivanfetch/ghapitest"

	g, err := prme.New(os.Getenv("GH_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}

	got, err := g.CreateEmptyTreeCommit(repo)
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Fatalf("empty sha returned creating an empty-tree commit in repository %q", repo)
	}

	ok, err := g.CommitExists(repo, got)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("empty-tree commit %q does not actually exist in repository %q", got, repo)
	}
}
