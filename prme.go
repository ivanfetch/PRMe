package prme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
func (c client) GetRepoID(repo string) (int64, error) {
	apiURL := fmt.Sprintf("%s/repos/%s", c.apiHost, repo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Github HTTP %d: %v", resp.StatusCode, string(data))
	}

	var apiResp struct{ Id int64 }
	err = json.Unmarshal(data, &apiResp)
	if err != nil {
		return 0, err
	}

	return apiResp.Id, nil
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
		return false, fmt.Errorf("HTTP %d while getting commit %q in repository %q", resp.StatusCode, ref, repo)
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
		return "", fmt.Errorf("HTTP %d while creating empty-tree commit for repository %q", resp.StatusCode, repo)
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
		return false, fmt.Errorf("unexpected HTTP %d determining if branch %q exists in repository %q", resp.StatusCode, branch, repo)
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
	var branchJSON = []byte(fmt.Sprintf(`{"ref":"refs/heads/%s","sha":"%s"}`, branch, commitSha))
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(branchJSON))
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
		return fmt.Errorf("HTTP %d while creating branch %q using commit %q in repository %q", resp.StatusCode, branch, commitSha, repo)
	}
	return nil
}

func (c client) CreateOrphanBranch(repo, branch string) error {
	commitSha, err := c.CreateEmptyTreeCommit(repo)
	if err != nil {
		return err
	}
	err = c.CreateBranch(repo, branch, commitSha)
	if err != nil {
		return err
	}
	return nil
}
