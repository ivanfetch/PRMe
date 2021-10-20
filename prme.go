package prme

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var Version, GitCommit string // Populated by build process

type Client struct {
	token, apiHost string
	httpClient     *http.Client
}

// clientOption specifies prme client options as functions.
type clientOption func(*Client) error

// WithAPIHost sets the Github API hostname for an instance of the client.
func WithAPIHost(host string) clientOption {
	return func(c *Client) error {
		c.apiHost = host
		return nil
	}
}

// WithHTTPClient sets a custom net/http.Client for an instance of the client.
func WithHTTPClient(hc *http.Client) clientOption {
	return func(c *Client) error {
		c.httpClient = hc
		return nil
	}
}

func NewClient(token string, options ...clientOption) (*Client, error) {
	if token == "" {
		return nil, errors.New("the Github token cannot be empty, please specify a personal access token")
	}

	c := &Client{
		token:      token,
		apiHost:    "https://api.github.com",
		httpClient: &http.Client{Timeout: time.Second * 10},
	}

	for _, o := range options {
		err := o(c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Client) MakeAPIRequest(method, URI string) (*http.Response, error) {
	if !strings.HasPrefix(URI, "/") {
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

func (c *Client) MakeAPIRequestWithData(method, URI string, body []byte) (*http.Response, error) {
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

func RunGitCommand(workingDir string, arg string, extraArgs ...string) (string, error) {
	args := append([]string{arg}, extraArgs...)
	cmd := exec.Command("git", args...)
	cmd.Dir = workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command %q returned error %w and output: %s", cmd, err, output)
	}
	return strings.TrimSuffix(string(output), "\n"), nil
}

type repo struct {
	Client       *Client
	ownerAndName string
}

func (r repo) String() string {
	return r.ownerAndName
}

func NewRepo(ownerAndName, token string, clientOptions ...clientOption) (*repo, error) {
	if ownerAndName == "" {
		return nil, errors.New("the repository cannot be empty, please specify a repository of the form OwnerName/RepositoryName")
	}
	if !strings.Contains(ownerAndName, "/") {
		return nil, errors.New("the repository must be of the form OwnerName/RepositoryName")
	}
	c, err := NewClient(token, clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("while constructing client for repository: %w", err)
	}
	return &repo{
		Client:       c,
		ownerAndName: ownerAndName,
	}, nil
}

func (r repo) Exists() (bool, error) {
	apiURI := fmt.Sprintf("/repos/%s", r)
	resp, err := r.Client.MakeAPIRequest(http.MethodGet, apiURI)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP %d for %s while getting repository %q", resp.StatusCode, apiURI, r)
	}
	var repoAPIResp struct {
		FullName string `json:"full_name"`
	}
	err = json.NewDecoder(resp.Body).Decode(&repoAPIResp)
	if err != nil {
		return false, err
	}
	if strings.ToLower(repoAPIResp.FullName) != strings.ToLower(r.String()) {
		return false, fmt.Errorf("incorrect repository name %q returned while checking if repository %q exists", repoAPIResp.FullName, r)
	}
	return true, nil
}

func (r repo) CommitExists(ref string) (bool, error) {
	apiURI := fmt.Sprintf("/repos/%s/git/commits/%s", r, ref)
	resp, err := r.Client.MakeAPIRequest(http.MethodGet, apiURI)
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

func (r repo) CreateOrphanBranches(branchNames ...string) error {
	if len(branchNames) == 0 {
		return errors.New("please supply at least one branch name")
	}
	for i, branchName := range branchNames {
		if branchName == "" {
			return fmt.Errorf("branchName[%d] cannot be empty", i)
		}
	}
	repoURL := fmt.Sprintf("ssh://git@github.com/%s", r)
	tempDir, err := os.MkdirTemp("", "pr-me-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	tempDirWithRepo := tempDir + "/" + r.String()
	_, err = RunGitCommand(tempDir, "clone", repoURL, r.String())
	if err != nil {
		return err
	}
	commitSha, err := RunGitCommand(tempDirWithRepo, "commit-tree", "4b825dc642cb6eb9a060e54bf8d69288fbee4904", "-m", "empty-tree commit")
	if err != nil {
		return err
	}
	if commitSha == "" {
		return errors.New("empty commit sha returned after creating empty-tree commit")
	}
	for _, branchName := range branchNames {
		_, err = RunGitCommand(tempDirWithRepo, "branch", branchName, commitSha)
		if err != nil {
			return err
		}
	}
	gitPushArgs := append([]string{"origin"}, branchNames...)
	_, err = RunGitCommand(tempDirWithRepo, "push", gitPushArgs...)
	if err != nil {
		return err
	}
	return nil
}

func (r repo) BranchExists(branch string) (bool, error) {
	apiURI := fmt.Sprintf("/repos/%s/branches/%s", r, branch)
	resp, err := r.Client.MakeAPIRequest(http.MethodGet, apiURI)
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

// MergeBranch merges headBranch into baseBranch in the given repository.
func (r repo) MergeBranch(baseBranch, headBranch string) error {
	apiURI := fmt.Sprintf("/repos/%s/merges", r)
	mergeJSON := fmt.Sprintf(`{"base":"%s","head":"%s"}`, baseBranch, headBranch)
	resp, err := r.Client.MakeAPIRequestWithData(http.MethodPost, apiURI, []byte(mergeJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("HTTP %d for %s while merging branch %q into %q in repository %q", resp.StatusCode, apiURI, headBranch, baseBranch, r)
	}
	return nil
}

// CreatePullRequest creates a pull request using the specified properties.
// returning the PR URL.
func (r repo) CreatePullRequest(title, body, baseBranch, headBranch string) (PRURL string, err error) {
	apiURI := fmt.Sprintf("/repos/%s/pulls", r)
	PRJSON := fmt.Sprintf(`{"title":"%s","body":"%s","base":"%s","head":"%s"}`, title, body, baseBranch, headBranch)
	resp, err := r.Client.MakeAPIRequestWithData(http.MethodPost, apiURI, []byte(PRJSON))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("HTTP %d for %s while creating pull request in repository %q, base branch %q, and head branch %q", resp.StatusCode, apiURI, r, baseBranch, headBranch)
	}
	var PRAPIResp struct {
		HTMLURL *string `json:"html_url"`
	}
	err = json.NewDecoder(resp.Body).Decode(&PRAPIResp)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if PRAPIResp.HTMLURL == nil {
		return "", errors.New("the Github API did not return a pull request HTML URL")
	}
	return *PRAPIResp.HTMLURL, nil
}

type FullPullRequestCreator struct {
	Token, Repo, FullRepoBranch, Title, Body, BaseBranch, HeadBranch string
}

type fullPullRequestCreatorOption func(*FullPullRequestCreator) error

func WithToken(token string) fullPullRequestCreatorOption {
	return func(f *FullPullRequestCreator) error {
		if token == "" {
			return errors.New("token cannot be empty, please specify a Github personal access token")
		}
		f.Token = token
		return nil
	}
}

func WithFullRepoBranch(branch string) fullPullRequestCreatorOption {
	return func(f *FullPullRequestCreator) error {
		if branch == "" {
			return errors.New("the full repo branch cannot be empty")
		}
		f.FullRepoBranch = branch
		return nil
	}
}

func WithTitle(title string) fullPullRequestCreatorOption {
	return func(f *FullPullRequestCreator) error {
		if title == "" {
			return errors.New("the title cannot be empty")
		}
		f.Title = title
		return nil
	}
}

func WithBody(body string) fullPullRequestCreatorOption {
	return func(f *FullPullRequestCreator) error {
		if body == "" {
			return errors.New("the body cannot be empty")
		}
		f.Body = body
		return nil
	}
}

func WithBaseBranchName(branch string) fullPullRequestCreatorOption {
	return func(f *FullPullRequestCreator) error {
		if branch == "" {
			return errors.New("the base branch name cannot be empty")
		}
		f.BaseBranch = branch
		return nil
	}
}

func WithHeadBranchName(branch string) fullPullRequestCreatorOption {
	return func(f *FullPullRequestCreator) error {
		if branch == "" {
			return errors.New("the head branch name cannot be empty")
		}
		f.HeadBranch = branch
		return nil
	}
}

func NewFullPullRequestCreator(repo string, options ...fullPullRequestCreatorOption) (*FullPullRequestCreator, error) {
	if repo == "" {
		return nil, errors.New("repo cannot be empty")
	}
	f := &FullPullRequestCreator{
		Repo:           repo,
		Token:          "",
		Title:          "Full Review",
		Body:           "A full review of the entire repository. When this PR is complete, be sure to manually merge its base branch into the main branch for this repository.",
		BaseBranch:     "prme-full-review",
		HeadBranch:     "prme-full-content",
		FullRepoBranch: "main",
	}
	for _, option := range options {
		err := option(f)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

func (f FullPullRequestCreator) Create() (string, error) {
	if f.FullRepoBranch == "" {
		return "", errors.New("the full repo branch cannot be empty")
	}
	if f.BaseBranch == "" {
		return "", errors.New("the base branch cannot be empty")
	}
	if f.HeadBranch == "" {
		return "", errors.New("the head branch cannot be empty")
	}
	if f.Title == "" {
		return "", errors.New("the title cannot be empty")
	}
	if f.Body == "" {
		return "", errors.New("the body cannot be empty")
	}
	r, err := NewRepo(f.Repo, f.Token)
	if err != nil {
		return "", err
	}
	ok, err := r.Exists()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("repository %q does not exist or the access token does not provide access", r)
	}
	ok, err = r.BranchExists(f.FullRepoBranch)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("full repository branch %q does not exist in repository %q", f.FullRepoBranch, r)
	}
	ok, err = r.BranchExists(f.BaseBranch)
	if err != nil {
		return "", err
	}
	if ok {
		return "", fmt.Errorf("base branch %q already exists in repository %q", f.BaseBranch, r)
	}
	ok, err = r.BranchExists(f.HeadBranch)
	if err != nil {
		return "", err
	}
	if ok {
		return "", fmt.Errorf("head branch %q already exists in repository %q", f.HeadBranch, r)
	}

	err = r.CreateOrphanBranches(f.BaseBranch, f.HeadBranch)
	if err != nil {
		return "", err
	}
	err = r.MergeBranch(f.HeadBranch, f.FullRepoBranch)
	if err != nil {
		return "", err
	}
	PRURL, err := r.CreatePullRequest(f.Title, f.Body, f.BaseBranch, f.HeadBranch)
	if err != nil {
		return "", err
	}
	return PRURL, nil
}

func flagOrEnvValue(f *flag.Flag) {
	envVarName := "PRME_" + strings.ToUpper(f.Name)
	envVarValue := os.Getenv(envVarName)
	if envVarValue != "" && f.Value.String() == f.DefValue {
		_ = f.Value.Set(envVarValue)
	}
}

func NewFullPullRequestCreatorFromArgs(args []string, output, errOutput io.Writer) (*FullPullRequestCreator, error) {
	fs := flag.NewFlagSet("prme", flag.ExitOnError)
	fs.SetOutput(errOutput)
	fs.Usage = func() {
		fmt.Fprintf(errOutput, `This program creates a pull request that reviews all content of a Github repository.

The GH_TOKEN environment variable must be set to a Github personal access token. To create a token, see https://github.com/settings/tokens

Usage: %s [flags] <repository>
The <repository> should be of the form OwnerName/RepositoryName

For example:
export GH_TOKEN='ghp_.....'
%s ivanfetch/pr-me

Available command-line flags:
`,
			fs.Name(), fs.Name())
		fs.PrintDefaults()
		fmt.Fprintf(errOutput, `
The following environment variables override defaults. Command-line flags will override everything.

		<Environment Variable>	<Current Value>
PRME_FBRANCH	%q
PRME_TITLE	%q
PRME_BODY	%q
PRME_BBRANCH	%q
PRME_HBRANCH	%q
`,
			os.Getenv("PRME_FBRANCH"), os.Getenv("PRME_TITLE"), os.Getenv("PRME_BODY"), os.Getenv("PRME_BBRANCH"), os.Getenv("PRME_HBRANCH"))
	}

	defaultValues, err := NewFullPullRequestCreator("dummyRepo")
	if err != nil {
		return nil, fmt.Errorf("while getting default values: %w", err)
	}

	CLIVersion := fs.Bool("version", false, "Display the version and git commit.")
	CLIFullRepoBranch := fs.String("fbranch", defaultValues.FullRepoBranch, "The name of the existing branch, such as main or master, containing all repository content. This is also set via the PRME_FBRANCH environment variable.")
	CLITitle := fs.String("title", defaultValues.Title, "The title of the pull request. This is also set via the PRME_TITLE environment variable.")
	CLIBody := fs.String("body", defaultValues.Body, "The body; first comment of the pull request. This is also set via the PRME_TITLE environment variable.")
	CLIBaseBranch := fs.String("bbranch", defaultValues.BaseBranch, "The name of the base orphan branch to create for the pull request.This is also set via the PRME_BBRANCH environment variable.")
	CLIHeadBranch := fs.String("hbranch", defaultValues.HeadBranch, "The name of the head review branch to create for the pull request, where review fixes should be pushed. This is also set via the PRME_HBRANCH environment variable.")
	err = fs.Parse(args)
	if err != nil {
		return nil, err
	}
	fs.VisitAll(flagOrEnvValue)
	if *CLIVersion {
		return nil, fmt.Errorf("%s version %s, git commit %s\n", fs.Name(), Version, GitCommit)
	}
	if fs.NArg() == 0 {
		return nil, fmt.Errorf(
			`Set the GH_TOKEN environment variable to a Github personal access token, then run this program with a repository name for which you would like a pull request that reviews all files.
For example: %s IvanFetch/myproject

Run %s -h for additional help.`,
			fs.Name(), fs.Name())
	}
	if fs.NArg() > 1 {
		return nil, fmt.Errorf("Please only specify one repository name, and make sure any command-line flags come first. RUn %s -h for additional help.", fs.Name())
	}
	repoName := strings.TrimPrefix(fs.Args()[0], "github.com/")
	f, err := NewFullPullRequestCreator(repoName)
	if err != nil {
		return nil, err
	}
	f.Token = os.Getenv("GH_TOKEN")
	if f.Token == "" {
		return nil, errors.New("Please set the GH_TOKEN environment variable to a Github personal access token. Tokens can be managed at https://github.com/settings/tokens")
	}
	f.FullRepoBranch = *CLIFullRepoBranch
	f.Title = *CLITitle
	f.Body = *CLIBody
	f.BaseBranch = *CLIBaseBranch
	f.HeadBranch = *CLIHeadBranch
	return f, nil
}

func CreateFullPullRequest(repo string, options ...fullPullRequestCreatorOption) (string, error) {
	f, err := NewFullPullRequestCreator(repo, options...)
	if err != nil {
		return "", err
	}
	PRURL, err := f.Create()
	if err != nil {
		return "", err
	}
	return PRURL, nil
}

func CreateFullPullRequestFromArgs(args []string, output, errOutput io.Writer) (string, error) {
	FPR, err := NewFullPullRequestCreatorFromArgs(args, output, errOutput)
	if err != nil {
		return "", err
	}
	PRURL, err := FPR.Create()
	if err != nil {
		return "", err
	}
	return PRURL, nil
}

func RunCLI() {
	PRURL, err := CreateFullPullRequestFromArgs(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("A full pull request has been created at %s\n", PRURL)
}
