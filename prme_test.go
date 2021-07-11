package prme_test

import (
	"os"
	"prme"
	"testing"
)

func TestConnectToGithubAPI(t *testing.T) {
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
