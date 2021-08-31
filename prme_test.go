package prme_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"prme"
	"testing"
)

func TestCommitExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCommitExists.json"
	repo := "ivanfetch/ghapitest"
	commitSha := "87d2b8f97a27554711c1eb0d1bb0f8f623a2af25"

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

	g, err := prme.New("dummy token",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := g.CommitExists(repo, commitSha)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatalf("commit %q not found in repository %s, using test data file %s", commitSha, repo, testFileName)
	}
}

func TestCommitNotExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCommitNotExists.json"
	repo := "ivanfetch/ghapitest"
	commitSha := "will-not-exist"

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

	g, err := prme.New("dummy token",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := g.CommitExists(repo, commitSha)
	if err != nil {
		t.Fatal(err)
	}

	if ok {
		t.Fatalf("commit %q found in repository %s, using test data file %s", commitSha, repo, testFileName)
	}
}

func TestCommitExistsWithIncorrectJSONReturnsError(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCommitExistsWithIncorrectJSONReturnsError.json"
	repo := "ivanfetch/ghapitest"
	commitSha := "will-not-match"

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

	g, err := prme.New("dummy token",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = g.CommitExists(repo, commitSha)
	if err == nil {
		t.Fatalf("error expected, looking for commit %q in repository %q, using test data file %q", commitSha, repo, testFileName)
	}
}

func TestBranchExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestBranchExists.json"
	repo := "ivanfetch/ghapitest"
	branch := "test"

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

	g, err := prme.New("dummy token",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := g.BranchExists(repo, branch)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatalf("branch %q not found in repository %s, using test data file %s", branch, repo, testFileName)
	}
}

func TestBranchNotExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestBranchNotExists.json"
	repo := "ivanfetch/ghapitest"
	branch := "test"

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

	g, err := prme.New("dummy token",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := g.BranchExists(repo, branch)
	if err != nil {
		t.Fatal(err)
	}

	if ok {
		t.Fatalf("branch %q found in repository %s, using test data file %s", branch, repo, testFileName)
	}
}

func TestBranchExistsWithIncorrectJSONReturnsError(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestBranchExistsWithIncorrectJSONReturnsError.json"
	repo := "ivanfetch/ghapitest"
	branch := "will-not-match"

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

	g, err := prme.New("dummy token",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = g.BranchExists(repo, branch)
	if err == nil {
		t.Fatalf("error expected, looking for branch %q in repository %q, using test data file %q", branch, repo, testFileName)
	}

}

func TestCreateEmptyTreeCommit(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCreateEmptyTreeCommit.json"
	repo := "ivanfetch/ghapitest"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest/git/commits"
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

	g, err := prme.New("dummy token",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := g.CreateEmptyTreeCommit(repo)
	if err != nil {
		t.Fatal(err)
	}

	want := "828e2e095e8dde51386b842b736afa59f6277152"
	if want != got {
		t.Fatalf("want %q, got %q, using test data file %s", want, got, testFileName)
	}
}

func TestCreateEmptyTreeCommitInNonexistentRepositoryReturnsError(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCreateEmptyTreeCommitInNonexistentRepositoryReturnsError.json"
	repo := "nonexistent-repository"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/nonexistent-repository/git/commits"
		gotRequestURL := r.RequestURI
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
		}
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		w.WriteHeader(http.StatusNotFound)
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
	}))
	defer ts.Close()

	g, err := prme.New("dummy token",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := g.CreateEmptyTreeCommit(repo)
	if err == nil {
		t.Fatalf("created commit %q in repository %q", got, repo)
	}
}

func TestCreateBranch(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCreateBranch.json"
	repo := "ivanfetch/ghapitest"
	commitSha := "28ee640a8ce0c22adf3534c7f5971286bfd30642"
	branch := "test"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantRequestURL := "/repos/ivanfetch/ghapitest/git/refs"
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

	g, err := prme.New("dummy token",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = g.CreateBranch(repo, branch, commitSha)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateBranchReturnsError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		repo, branch, commitSha, wantRequestURL string
		returnHTTPStatusCode                    int
	}{
		{
			repo:                 "ivanfetch/non-existent-repo",
			branch:               "dummy-branch-name",
			commitSha:            "dummy-commit-sha",
			wantRequestURL:       "/repos/ivanfetch/non-existent-repo/git/refs",
			returnHTTPStatusCode: http.StatusNotFound,
		},
		{
			repo:                 "ivanfetch/ghapitest",
			branch:               "dummy-branch-name",
			commitSha:            "nonexistent-commit-sha",
			wantRequestURL:       "/repos/ivanfetch/ghapitest/git/refs",
			returnHTTPStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotRequestURL := r.RequestURI
			if tc.wantRequestURL != gotRequestURL {
				t.Errorf("Want %q for Github URL, got %q using repository %q, branch %q, and commit sha %q", tc.wantRequestURL, gotRequestURL, tc.repo, tc.branch, tc.commitSha)
			}
			w.WriteHeader(tc.returnHTTPStatusCode)
		}))
		defer ts.Close()

		g, err := prme.New("dummy token",
			prme.WithHTTPClient(ts.Client()),
			prme.WithAPIHost(ts.URL),
		)
		if err != nil {
			t.Fatal(err)
		}

		err = g.CreateBranch(tc.repo, tc.branch, tc.commitSha)
		if err == nil {
			t.Errorf("expected error, using repository %q, branch %q, and commit sha %q", tc.repo, tc.branch, tc.commitSha)
		}
	}
}

func TestMergeBranch(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestMergeBranch.json"
	repo := "ivanfetch/ghapitest"
	baseBranch := "review"
	headBranch := "main"

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

	g, err := prme.New("dummy token",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = g.MergeBranch(repo, baseBranch, headBranch)
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
				t.Errorf("Want %q for Github URL, got %q using repository %q, base branch %q, and head branch %q", tc.wantRequestURL, gotRequestURL, tc.repo, tc.baseBranch, tc.headBranch)
			}
			w.WriteHeader(tc.returnHTTPStatusCode)
		}))
		defer ts.Close()

		g, err := prme.New("dummy token",
			prme.WithHTTPClient(ts.Client()),
			prme.WithAPIHost(ts.URL),
		)
		if err != nil {
			t.Fatal(err)
		}
		err = g.MergeBranch(tc.repo, tc.baseBranch, tc.headBranch)
		if err == nil {
			t.Errorf("expected error, using repository %q, base branch %q, and head branch %q", tc.repo, tc.baseBranch, tc.headBranch)
		}
	}
}

func TestCreatePullRequest(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCreatePullRequest.json"
	title := "test1"
	body := "A full review of this repository"
	repo := "ivanfetch/ghapitest"
	baseBranch := "orphan"
	headBranch := "review"

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

	g, err := prme.New("dummy token",
		prme.WithHTTPClient(ts.Client()),
		prme.WithAPIHost(ts.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := g.CreatePullRequest(title, body, repo, baseBranch, headBranch)
	if err != nil {
		t.Fatal(err)
	}
	want := 7
	if want != got {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestCreatePullRequestReturnsError(t *testing.T) {
	testCases := []struct {
		repo, baseBranch, headBranch, wantRequestURL string
		returnHTTPStatusCode                         int
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
				t.Errorf("Want %q for Github URL, got %q using repository %q, base branch %q, and head branch %q", tc.wantRequestURL, gotRequestURL, tc.repo, tc.baseBranch, tc.headBranch)
			}
			w.WriteHeader(tc.returnHTTPStatusCode)
		}))
		defer ts.Close()

		g, err := prme.New("dummy token",
			prme.WithHTTPClient(ts.Client()),
			prme.WithAPIHost(ts.URL),
		)
		if err != nil {
			t.Fatal(err)
		}

		_, err = g.CreatePullRequest("title not used", "body not used", tc.repo, tc.baseBranch, tc.headBranch)
		if err == nil {
			t.Fatalf("error expected using repository %q, base branch %q, and head branch %q", tc.repo, tc.baseBranch, tc.headBranch)
		}
	}
}
