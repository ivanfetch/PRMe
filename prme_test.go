package prme_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"prme"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGitCommand(t *testing.T) {
	t.Parallel()
	_, err := prme.RunGitCommand(os.TempDir(), "version")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGitCommandReturnsError(t *testing.T) {
	t.Parallel()
	got, err := prme.RunGitCommand(os.TempDir(), "dummyCommand")
	if err == nil {
		t.Fatalf("expected error for command git dummyCommand, got %q", got)
	}
}

func TestRepoExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestRepoExists.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	r, err := prme.NewRepo("ivanfetch/ghapitest", "dummyToken",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := r.Exists()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("repository %s not found, using test data file %s", r, testFileName)
	}
}

func TestRepoNotExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestRepoNotExists.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		w.WriteHeader(http.StatusNotFound)
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	r, err := prme.NewRepo("ivanfetch/ghapitest", "dummyToken",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := r.Exists()
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("repository %s exists, using test data file %s", r, testFileName)
	}
}

func TestRepoExistsWithIncorrectJSONReturnsError(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestRepoExistsWithIncorrectJSONReturnsError.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		f, err := os.Open(testFileName)
		defer f.Close()
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	r, err := prme.NewRepo("ivanfetch/ghapitest", "dummyToken",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = r.Exists()
	if err == nil {
		t.Fatalf("error expected, looking for repository %q, using test data file %q", r, testFileName)
	}
}

func TestCommitExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCommitExists.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest/git/commits/87d2b8f97a27554711c1eb0d1bb0f8f623a2af25"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	r, err := prme.NewRepo("ivanfetch/ghapitest", "dummyToken",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	commitSha := "87d2b8f97a27554711c1eb0d1bb0f8f623a2af25"
	ok, err := r.CommitExists(commitSha)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatalf("commit %q not found in repository %s, using test data file %s", commitSha, r, testFileName)
	}
}

func TestCommitNotExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCommitNotExists.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest/git/commits/will-not-exist"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		w.WriteHeader(http.StatusNotFound)
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	r, err := prme.NewRepo("ivanfetch/ghapitest", "dummyToken",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	commitSha := "will-not-exist"
	ok, err := r.CommitExists(commitSha)
	if err != nil {
		t.Fatal(err)
	}

	if ok {
		t.Fatalf("commit %q found in repository %s, using test data file %s", commitSha, r, testFileName)
	}
}

func TestCommitExistsWithIncorrectJSONReturnsError(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCommitExistsWithIncorrectJSONReturnsError.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest/git/commits/will-not-match"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		f, err := os.Open(testFileName)
		defer f.Close()
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	r, err := prme.NewRepo("ivanfetch/ghapitest", "dummyToken",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	commitSha := "will-not-match"
	_, err = r.CommitExists(commitSha)
	if err == nil {
		t.Fatalf("error expected, looking for commit %q in repository %q, using test data file %q", commitSha, r, testFileName)
	}
}

func TestBranchExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestBranchExists.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest/branches/test"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	r, err := prme.NewRepo("ivanfetch/ghapitest", "dummyToken",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	branch := "test"
	ok, err := r.BranchExists(branch)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatalf("branch %q not found in repository %s, using test data file %s", branch, r, testFileName)
	}
}

func TestBranchNotExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestBranchNotExists.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest/branches/test"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		w.WriteHeader(http.StatusNotFound)
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	r, err := prme.NewRepo("ivanfetch/ghapitest", "dummyToken",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	branch := "test"
	ok, err := r.BranchExists(branch)
	if err != nil {
		t.Fatal(err)
	}

	if ok {
		t.Fatalf("branch %q found in repository %s, using test data file %s", branch, r, testFileName)
	}
}

func TestBranchExistsWithIncorrectJSONReturnsError(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestBranchExistsWithIncorrectJSONReturnsError.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest/branches/will-not-match"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		f, err := os.Open(testFileName)
		defer f.Close()
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	r, err := prme.NewRepo("ivanfetch/ghapitest", "dummyToken",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	branch := "will-not-match"
	_, err = r.BranchExists(branch)
	if err == nil {
		t.Fatalf("error expected, looking for branch %q in repository %q, using test data file %q", branch, r, testFileName)
	}
}

func TestMergeBranch(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestMergeBranch.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest/merges"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		w.WriteHeader(http.StatusCreated)
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	r, err := prme.NewRepo("ivanfetch/ghapitest", "dummyToken",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	baseBranch := "review"
	headBranch := "main"
	err = r.MergeBranch(baseBranch, headBranch)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMergeBranchReturnsError(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		repo, baseBranch, headBranch, wantRequestURL string
		returnHTTPStatusCode                         int
	}{
		{
			repo:                 "ivanfetch/non-existent-repo",
			baseBranch:           "orphan",
			headBranch:           "review",
			wantRequestURL:       "/repos/ivanfetch/non-existent-repo/merges",
			returnHTTPStatusCode: http.StatusNotFound,
		},
		{
			repo:                 "ivanfetch/ghapitest",
			baseBranch:           "non-existent-base",
			headBranch:           "review",
			wantRequestURL:       "/repos/ivanfetch/ghapitest/merges",
			returnHTTPStatusCode: http.StatusUnprocessableEntity,
		},
		{
			repo:                 "ivanfetch/ghapitest",
			baseBranch:           "orphan",
			headBranch:           "non-existent-branch",
			wantRequestURL:       "/repos/ivanfetch/ghapitest/merges",
			returnHTTPStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotRequestURL := r.RequestURI
			if tc.wantRequestURL != gotRequestURL {
				t.Errorf("Want %q for Github URL, got %q using test case %v", tc.wantRequestURL, gotRequestURL, tc)
			}
			w.WriteHeader(tc.returnHTTPStatusCode)
		}))
		defer ts.Close()

		r, err := prme.NewRepo(tc.repo, "dummyToken",
			prme.WithHTTPClient(ts.Client()),
			prme.WithAPIHost(ts.URL),
		)
		if err != nil {
			t.Fatal(err)
		}
		err = r.MergeBranch(tc.baseBranch, tc.headBranch)
		if err == nil {
			t.Errorf("expected error, using repository %q, base branch %q, and head branch %q", r, tc.baseBranch, tc.headBranch)
		}
	}
}

func TestCreatePullRequest(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCreatePullRequest.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest/pulls"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		w.WriteHeader(http.StatusCreated)
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	r, err := prme.NewRepo("ivanfetch/ghapitest", "dummyToken",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := r.CreatePullRequest("test1",
		"A full review of this repository",
		"orphan",
		"review")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://github.com/ivanfetch/ghapitest/pull/7"
	if want != got {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestCreatePullRequestReturnsError(t *testing.T) {
	testCases := []struct {
		repo, wantRequestURL, title, body, baseBranch, headBranch string
		returnHTTPStatusCode                                      int
	}{
		{
			repo:                 "ivanfetch/non-existent-repo",
			baseBranch:           "orphan",
			headBranch:           "review",
			wantRequestURL:       "/repos/ivanfetch/non-existent-repo/pulls",
			returnHTTPStatusCode: http.StatusNotFound,
		},
		{
			repo:                 "ivanfetch/ghapitest",
			baseBranch:           "non-existent-base",
			headBranch:           "review",
			wantRequestURL:       "/repos/ivanfetch/ghapitest/pulls",
			returnHTTPStatusCode: http.StatusUnprocessableEntity,
		},
		{
			repo:                 "ivanfetch/ghapitest",
			baseBranch:           "orphan",
			headBranch:           "non-existent-branch",
			wantRequestURL:       "/repos/ivanfetch/ghapitest/pulls",
			returnHTTPStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotRequestURL := r.RequestURI
			if tc.wantRequestURL != gotRequestURL {
				t.Errorf("Want %q for Github URL, got %q using test case %v", tc.wantRequestURL, gotRequestURL, tc)
			}
			w.WriteHeader(tc.returnHTTPStatusCode)
		}))
		defer ts.Close()

		r, err := prme.NewRepo(tc.repo, "dummyToken",
			prme.WithHTTPClient(ts.Client()),
			prme.WithAPIHost(ts.URL),
		)
		if err != nil {
			t.Fatal(err)
		}
		_, err = r.CreatePullRequest(tc.title, tc.body, tc.baseBranch, tc.headBranch)
		if err == nil {
			t.Fatalf("error expected using repository %q, base branch %q, and head branch %q", r, tc.baseBranch, tc.headBranch)
		}
	}
}

func TestNewFullPullRequestCreatorFromArgs(t *testing.T) {
	testCases := []struct {
		description  string
		args         []string
		setEnv, want prme.FullPullRequestCreator
	}{
		{
			description: "no arguments which will use default values",
			args:        []string{"dummyRepo"},
			// Avoid environment in the calling OS breaking the test.
			setEnv: prme.FullPullRequestCreator{
				Token:          "",
				FullRepoBranch: "",
				Title:          "",
				Body:           "",
				BaseBranch:     "",
				HeadBranch:     "",
			},
			want: prme.FullPullRequestCreator{
				Repo:           "dummyRepo",
				FullRepoBranch: "main",
				Title:          "Full Review",
				Body:           "A full review of the entire repository. When this PR is complete, be sure to manually merge its base branch into the main branch for this repository.",
				BaseBranch:     "prme-full-review",
				HeadBranch:     "prme-full-content",
			},
		},
		{
			description: "set environment variables",
			args:        []string{"dummyRepo"},
			setEnv: prme.FullPullRequestCreator{
				Token:          "dummyTokenSetByEnvVar",
				FullRepoBranch: "master",
				Title:          "complete review",
				Body:           "A full review.",
				BaseBranch:     "orphan",
				HeadBranch:     "review",
			},
			want: prme.FullPullRequestCreator{
				Token:          "dummyTokenSetByEnvVar",
				Repo:           "dummyRepo",
				FullRepoBranch: "master",
				Title:          "complete review",
				Body:           "A full review.",
				BaseBranch:     "orphan",
				HeadBranch:     "review",
			},
		},
		{
			description: "specify flags",
			args:        []string{"-title", "my review", "-body", "another review!", "-fbranch", "prod", "-bbranch", "base", "-hbranch", "myreview", "myrepo"},
			want: prme.FullPullRequestCreator{
				Repo:           "myrepo",
				FullRepoBranch: "prod",
				Title:          "my review",
				Body:           "another review!",
				BaseBranch:     "base",
				HeadBranch:     "myreview",
			},
		},
	}
	// Use of t.Setenv() below, prohibits t.Parallel()
	for _, tc := range testCases {
		t.Setenv("GH_TOKEN", tc.setEnv.Token)
		t.Setenv("PRME_TITLE", tc.setEnv.Title)
		t.Setenv("PRME_BODY", tc.setEnv.Body)
		t.Setenv("PRME_FBRANCH", tc.setEnv.FullRepoBranch)
		t.Setenv("PRME_BBRANCH", tc.setEnv.BaseBranch)
		t.Setenv("PRME_HBRANCH", tc.setEnv.HeadBranch)

		got, err := prme.NewFullPullRequestCreatorFromArgs(tc.args, ioutil.Discard, ioutil.Discard)
		if err != nil {
			t.Fatalf("for test-case %s, %v", tc.description, err)
		}
		t.Logf("test %q got FullPullRequestCreator: %+v", tc.description, got)

		cmpOptions := cmp.AllowUnexported(*got)
		if !cmp.Equal(tc.want, *got, cmpOptions) {
			t.Fatalf("got incorrect full pull request options for test %s\ndiff reflects want vs. got: %s", tc.description, cmp.Diff(tc.want, *got, cmpOptions))
		}
	}
}
