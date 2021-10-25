//go:build integration
// +build integration

package prme_test

import (
	"fmt"
	"github.com/ivanfetch/prme"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestCreateFullPullRequestIntegration(t *testing.T) {
	t.Parallel()
	githubToken := os.Getenv("GH_TOKEN")
	PRURL, err := prme.CreateFullPullRequest("ivanfetch/ghapitest",
		prme.WithFullRepoBranch("main"),
		prme.WithToken(githubToken),
		prme.WithTitle("integration test"),
		prme.WithBody("integration test"),
		prme.WithBaseBranchName("integrationtest-review-branch"),
		prme.WithHeadBranchName("integrationtest-content-branch"),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("created pull request %s", PRURL)
	err = closePullRequest(PRURL, githubToken)
	if err != nil {
		t.Fatalf("while cleaning up pull request: %v", err)
	}
	err = deleteBranch("ivanfetch/ghapitest", "integrationtest-content-branch", githubToken)
	if err != nil {
		t.Fatalf("while cleaning up head branch: %v", err)
	}
	err = deleteBranch("ivanfetch/ghapitest", "integrationtest-review-branch", githubToken)
	if err != nil {
		t.Fatalf("while cleaning up base branch: %v", err)
	}
}

func deleteBranch(repo, branch, token string) error {
	r, err := prme.NewRepo(repo, token)
	if err != nil {
		return err
	}
	apiURI := fmt.Sprintf("/repos/%s/git/refs/heads/%s", r, branch)
	resp, err := r.Client.MakeAPIRequest(http.MethodDelete, apiURI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("HTTP %d for %s while deleting branch %q in repository %q", resp.StatusCode, apiURI, branch, r)
	}
	return nil
}

// A sample pull request URL is: https://github.com/ivanfetch/ghapitest/pull/7
func closePullRequest(URL, token string) error {
	URLComponents := strings.Split(URL, "/")
	repo := URLComponents[3] + "/" + URLComponents[4]
	PRNumber := URLComponents[6]
	apiURI := fmt.Sprintf("/repos/%s/pulls/%s", repo, PRNumber)
	r, err := prme.NewRepo(repo, token)
	if err != nil {
		return err
	}
	PRJSON := `{"state":"closed"}`
	resp, err := r.Client.MakeAPIRequestWithData(http.MethodPatch, apiURI, []byte(PRJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s while closing pull request %s in repository %q", resp.StatusCode, apiURI, PRNumber, r)
	}
	return nil
}
