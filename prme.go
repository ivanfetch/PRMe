package prme

import (
	"bytes"
	"encoding/json"
	"errors"
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

func NewClient(token string, options ...clientOption) (*client, error) {
	if token == "" {
		return nil, errors.New("the Github token can not be an empty string, please specify a personal access token")
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

func (c *client) MakeAPIRequest(method, URI string) (*http.Response, error) {
	if strings.HasPrefix(URI, "/") == false {
		URI = "/" + URI
	}
	URL := c.apiHost + URI
	req, err := http.NewRequest(method, URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) MakeAPIRequestWithData(method, URI string, body []byte) (*http.Response, error) {
	if strings.HasPrefix(URI, "/") == false {
		URI = "/" + URI
	}
	URL := c.apiHost + URI
	req, err := http.NewRequest(method, URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type repo struct {
	client       *client
	ownerAndName string
}

func (r repo) String() string {
	return r.ownerAndName
}

func NewRepo(ownerAndName, token string, clientOptions ...clientOption) (*repo, error) {
	if ownerAndName == "" {
		return nil, errors.New("the repository can not be empty")
	}
	client, err := NewClient(token, clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("while constructing client for repository: %w", err)
	}
	return &repo{
		client:       client,
		ownerAndName: ownerAndName,
	}, nil
}

func (r repo) CommitExists(ref string) (bool, error) {
	apiURI := fmt.Sprintf("/repos/%s/git/commits/%s", r, ref)
	resp, err := r.client.MakeAPIRequest(http.MethodGet, apiURI)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP %d for %s while getting commit %q in repository %q", resp.StatusCode, apiURI, ref, r)
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

func (r repo) CreateEmptyTreeCommit() (string, error) {
	apiURI := fmt.Sprintf("/repos/%s/git/commits", r)
	// A commit to the Git builtin "empty tree."
	commitJSON := `{"message":"empty-tree commit","tree":"4b825dc642cb6eb9a060e54bf8d69288fbee4904"}`
	resp, err := r.client.MakeAPIRequestWithData(http.MethodPost, apiURI, []byte(commitJSON))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("HTTP %d for %s while creating empty-tree commit for repository %q", resp.StatusCode, apiURI, r)
	}
	var commitAPIResp struct{ Sha string }
	err = json.NewDecoder(resp.Body).Decode(&commitAPIResp)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return commitAPIResp.Sha, nil
}

func (r repo) BranchExists(branch string) (bool, error) {
	apiURI := fmt.Sprintf("/repos/%s/branches/%s", r, branch)
	resp, err := r.client.MakeAPIRequest(http.MethodGet, apiURI)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected HTTP %d for %s while determining if branch %q exists in repository %q", resp.StatusCode, apiURI, branch, r)
	}
	var branchAPIResp struct{ Name string }
	err = json.NewDecoder(resp.Body).Decode(&branchAPIResp)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if branchAPIResp.Name != branch {
		return false, fmt.Errorf("incorrect name %q returned while checking if branch %q exists", branchAPIResp.Name, branch)
	}
	return true, nil
}

func (r repo) CreateBranch(branch, commitSha string) error {
	apiURI := fmt.Sprintf("/repos/%s/git/refs", r)
	branchJSON := fmt.Sprintf(`{"ref":"refs/heads/%s","sha":"%s"}`, branch, commitSha)
	resp, err := r.client.MakeAPIRequestWithData(http.MethodPost, apiURI, []byte(branchJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("HTTP %d for %s while creating branch %q using commit %q in repository %q", resp.StatusCode, apiURI, branch, commitSha, r)
	}
	return nil
}

// MergeBranch merges headBranch into baseBranch in the given repository.
func (r repo) MergeBranch(baseBranch, headBranch string) error {
	apiURI := fmt.Sprintf("/repos/%s/merges", r)
	mergeJSON := fmt.Sprintf(`{"base":"%s","head":"%s"}`, baseBranch, headBranch)
	resp, err := r.client.MakeAPIRequestWithData(http.MethodPost, apiURI, []byte(mergeJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("HTTP %d for %s while merging branch %q into %q in repository %q", resp.StatusCode, apiURI, headBranch, baseBranch, r)
	}
	return nil
}

// CreatePullRequest creates a pull request using the specified title, body,
// and branches, returning the PR ID.
func (r repo) CreatePullRequest(title, body, baseBranch, headBranch string) (PRNumber int, err error) {
	apiURI := fmt.Sprintf("/repos/%s/pulls", r)
	PRJSON := fmt.Sprintf(`{"title":"%s","body":"%s","base":"%s","head":"%s"}`, title, body, baseBranch, headBranch)
	resp, err := r.client.MakeAPIRequestWithData(http.MethodPost, apiURI, []byte(PRJSON))
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("HTTP %d for %s while creating pull request in repository %q, base branch %q, and head branch %q", resp.StatusCode, apiURI, r, baseBranch, headBranch)
	}
	var PRAPIResp struct{ Number *int }
	err = json.NewDecoder(resp.Body).Decode(&PRAPIResp)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if PRAPIResp.Number == nil {
		return 0, errors.New("the Github API did not return a pull request number")
	}
	return *PRAPIResp.Number, nil
}

func (r repo) CreateFullPullRequest(fullRepoBranch, PRTitle, PRBody, PRBaseBranch, PRHeadBranch string) (int, error) {
	ok, err := r.BranchExists(fullRepoBranch)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("full repository branch %q does not exist in repository %q", fullRepoBranch, r)
	}
	ok, err = r.BranchExists(PRBaseBranch)
	if err != nil {
		return 0, err
	}
	if ok {
		return 0, fmt.Errorf("base branch %q already exists in repository %q", PRBaseBranch, r)
	}
	ok, err = r.BranchExists(PRHeadBranch)
	if err != nil {
		return 0, err
	}
	if ok {
		return 0, fmt.Errorf("head branch %q already exists in repository %q", PRHeadBranch, r)
	}

	emptyCommit, err := r.CreateEmptyTreeCommit()
	if err != nil {
		return 0, err
	}
	err = r.CreateBranch(PRBaseBranch, emptyCommit)
	if err != nil {
		return 0, err
	}
	err = r.CreateBranch(PRHeadBranch, emptyCommit)
	if err != nil {
		return 0, err
	}
	err = r.MergeBranch(PRHeadBranch, fullRepoBranch)
	if err != nil {
		return 0, err
	}
	PRNum, err := r.CreatePullRequest(PRTitle, PRBody, PRBaseBranch, PRHeadBranch)
	if err != nil {
		return 0, err
	}
	return PRNum, nil
}

func (r repo) DeleteBranch(branch string) error {
	apiURI := fmt.Sprintf("/repos/%s/git/refs/heads/%s", r, branch)
	resp, err := r.client.MakeAPIRequest(http.MethodDelete, apiURI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("HTTP %d for %s while deleting branch %q in repository %q", resp.StatusCode, apiURI, branch, r)
	}
	return nil
}

func (r repo) ClosePullRequest(number int) error {
	apiURI := fmt.Sprintf("/repos/%s/pulls/%d", r, number)
	PRJSON := `{"state":"closed"}`
	resp, err := r.client.MakeAPIRequestWithData(http.MethodPatch, apiURI, []byte(PRJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s while closing pull request %d in repository %q", resp.StatusCode, apiURI, number, r)
	}
	return nil
}
