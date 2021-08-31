package prme

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type client struct {
	token, apiHost string
	httpClient     *http.Client
}

// clientOption specifies prme client options as functions.
type clientOption func(*client) error

// WithAPIHost sets the Github API hostname for an instance of the client.
func WithAPIHost(host string) clientOption {
	return func(c *client) error {
		c.apiHost = host
		return nil
	}
}

// WithHTTPClient sets a custom net/http.Client for an instance of the client.
func WithHTTPClient(hc *http.Client) clientOption {
	return func(c *client) error {
		c.httpClient = hc
		return nil
	}
}

func New(token string, options ...clientOption) (*client, error) {
	if token == "" {
		return nil, fmt.Errorf("the Github token can not be an empty string, please specify a personal access token")
	}

	c := &client{
		token:      token,
		apiHost:    "https://api.github.com",
		httpClient: &http.Client{Timeout: time.Second * 3},
	}

	for _, o := range options {
		err := o(c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c client) CommitExists(repo, ref string) (bool, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/git/commits/%s", c.apiHost, repo, ref)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP %d for %s while getting commit %q in repository %q", resp.StatusCode, apiURL, ref, repo)
	}
	var commitAPIResp struct{ Sha string }
	err = json.NewDecoder(resp.Body).Decode(&commitAPIResp)
	if err != nil {
		return false, err
	}
	if commitAPIResp.Sha != ref {
		return false, fmt.Errorf("incorrect commit sha %q returned while checking if commit %q exists", commitAPIResp.Sha, ref)
	}
	return true, nil
}

func (c client) CreateEmptyTreeCommit(repo string) (string, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/git/commits", c.apiHost, repo)
	// A commit to the Git builtin "empty tree."
	commitJSON := `{"message":"empty-tree commit","tree":"4b825dc642cb6eb9a060e54bf8d69288fbee4904"}`
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(commitJSON))
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("HTTP %d for %s while creating empty-tree commit for repository %q", resp.StatusCode, apiURL, repo)
	}
	var commitAPIResp struct{ Sha string }
	err = json.NewDecoder(resp.Body).Decode(&commitAPIResp)
	if err != nil {
		return "", err
	}
	return commitAPIResp.Sha, nil
}

func (c client) BranchExists(repo, branch string) (bool, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/branches/%s", c.apiHost, repo, branch)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected HTTP %d for %s while determining if branch %q exists in repository %q", resp.StatusCode, apiURL, branch, repo)
	}
	var branchAPIResp struct{ Name string }
	err = json.NewDecoder(resp.Body).Decode(&branchAPIResp)
	if err != nil {
		return false, err
	}
	if branchAPIResp.Name != branch {
		return false, fmt.Errorf("incorrect name %q returned while checking if branch %q exists", branchAPIResp.Name, branch)
	}
	return true, nil
}

func (c client) CreateBranch(repo, branch, commitSha string) error {
	apiURL := fmt.Sprintf("%s/repos/%s/git/refs", c.apiHost, repo)
	branchJSON := fmt.Sprintf(`{"ref":"refs/heads/%s","sha":"%s"}`, branch, commitSha)
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(branchJSON))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("HTTP %d for %s while creating branch %q using commit %q in repository %q", resp.StatusCode, apiURL, branch, commitSha, repo)
	}
	return nil
}

// MergeBranch merges headBranch into baseBranch in the given repository.
func (c client) MergeBranch(repo, baseBranch, headBranch string) error {
	apiURL := fmt.Sprintf("%s/repos/%s/merges", c.apiHost, repo)
	mergeJSON := fmt.Sprintf(`{"base":"%s","head":"%s"}`, baseBranch, headBranch)
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(mergeJSON))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("HTTP %d for %s while merging branch %q into %q in repository %q", resp.StatusCode, apiURL, headBranch, baseBranch, repo)
	}
	return nil
}

// CreatePullRequest creates a pull request using the specified title, body,
// and branches, returning the PR ID.
func (c client) CreatePullRequest(title, body, repo, baseBranch, headBranch string) (PRNumber int, err error) {
	apiURL := fmt.Sprintf("%s/repos/%s/pulls", c.apiHost, repo)
	PRJSON := fmt.Sprintf(`{"title":"%s","body":"%s","base":"%s","head":"%s"}`, title, body, baseBranch, headBranch)
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(PRJSON))
	if err != nil {
		return 0, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("HTTP %d for %s while creating pull request in repository %q, base branch %q, and head branch %q", resp.StatusCode, apiURL, repo, baseBranch, headBranch)
	}
	var PRAPIResp struct{ Number *int }
	err = json.NewDecoder(resp.Body).Decode(&PRAPIResp)
	if err != nil {
		return 0, err
	}
	if PRAPIResp.Number == nil {
		return 0, fmt.Errorf("Github API did not return a pull request ID")
	}
	return *PRAPIResp.Number, nil
}

func (c client) CreateFullPullRequest(repo, fullRepoBranch, PRTitle, PRBody, PRBaseBranch, PRHeadBranch string) (int, error) {
	ok, err := c.BranchExists(repo, fullRepoBranch)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("full repository branch %q does not exist in repository %q", fullRepoBranch, repo)
	}
	ok, err = c.BranchExists(repo, PRBaseBranch)
	if err != nil {
		return 0, err
	}
	if ok {
		return 0, fmt.Errorf("base branch %q already exists in repository %q", PRBaseBranch, repo)
	}
	ok, err = c.BranchExists(repo, PRHeadBranch)
	if err != nil {
		return 0, err
	}
	if ok {
		return 0, fmt.Errorf("head branch %q already exists in repository %q", PRHeadBranch, repo)
	}

	emptyCommit, err := c.CreateEmptyTreeCommit(repo)
	if err != nil {
		return 0, err
	}
	err = c.CreateBranch(repo, PRBaseBranch, emptyCommit)
	if err != nil {
		return 0, err
	}
	err = c.CreateBranch(repo, PRHeadBranch, emptyCommit)
	if err != nil {
		return 0, err
	}
	err = c.MergeBranch(repo, PRHeadBranch, fullRepoBranch)
	if err != nil {
		return 0, err
	}
	PRID, err := c.CreatePullRequest(PRTitle, PRBody, repo, PRBaseBranch, PRHeadBranch)
	if err != nil {
		return 0, err
	}
	return PRID, nil
}

func (c client) DeleteBranch(repo, branch string) error {
	apiURL := fmt.Sprintf("%s/repos/%s/git/refs/heads/%s", c.apiHost, repo, branch)
	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("HTTP %d for %s while deleting branch %q in repository %q", resp.StatusCode, apiURL, branch, repo)
	}
	return nil
}

func (c client) ClosePullRequest(repo string, ID int) error {
	apiURL := fmt.Sprintf("%s/repos/%s/pulls/%d", c.apiHost, repo, ID)
	PRJSON := `{"state":"closed"}`
	req, err := http.NewRequest("PATCH", apiURL, strings.NewReader(PRJSON))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s while closing pull request %d in repository %q", resp.StatusCode, apiURL, ID, repo)
	}
	return nil
}
