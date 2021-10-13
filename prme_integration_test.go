//go:build integration
// +build integration

package prme_test

import (
	"fmt"
	"net/http"
	"os"
	"prme"
	"strings"
	"testing"
)

func TestCreateFullPullRequestIntegration(t *testing.T) {
	/* Perhaps use this instead of the +build line
	   if os.Getenv("PRME_INTEGRATION_TESTS") == "" {
	   		t.Skip("set the PRME_INTEGRATION_TESTS environment variable to run integration tests")
	   	}
	*/

	t.Parallel()

	PRURL, err := prme.CreateFullPullRequest("ivanfetch/ghapitest",
		prme.WithFullRepoBranch("main"),
		prme.WithToken(os.Getenv("GH_TOKEN")),
		prme.WithTitle("integration test"),
		prme.WithBody("integration test"),
		prme.WithBaseBranchName("integrationtest-review-branch"),
		prme.WithHeadBranchName("integrationtest-content-branch"),
	)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("created pull request %s", PRURL)

	err = closePullRequest(PRURL)
	if err != nil {
		t.Fatalf("while cleaning up pull request: %v", err)
	}
	err = deleteBranch(os.Getenv("GH_TOKEN"), "ivanfetch/ghapiteest", "integrationtest-content-branch")
	if err != nil {
		t.Fatalf("while cleaning up head branch: %v", err)
	}
	err = deleteBranch(os.Getenv("GH_TOKEN"), "ivanfetch/ghapitest", "integrationtest-review-branch")
	if err != nil {
		t.Fatalf("while cleaning up base branch: %v", err)
	}
}

func deleteBranch(token, repo, branch string) error {
	r, err := prme.NewRepo("ivanfetch/ghapitest", os.Getenv("GH_TOKEN"))
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

func closePullRequest(URL string) error {
	// A sample pull request URL is: https://github.com/ivanfetch/ghapitest/pull/7
	URLComponents := strings.Split(URL, "/")
	repo := URLComponents[3] + "/" + URLComponents[4]
	PRNumber := URLComponents[6]
	apiURI := fmt.Sprintf("/repos/%s/pulls/%s", repo, PRNumber)
	r, err := prme.NewRepo(repo, os.Getenv("GH_TOKEN"))
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
