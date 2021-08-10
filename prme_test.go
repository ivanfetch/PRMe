package prme_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"prme"
	"testing"
)

func TestConnectToGithubAPI(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/github-docs-repo.json"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
		gotRequestURL := r.RequestURI
		wantRequestURL := "/repos/github/docs"
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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

	gotID, err := g.GetRepoID("github/docs")
	if err != nil {
		t.Fatal(err)
	}

	var wantID int64 = 189621607 // ID for github/docs repo
	if gotID != wantID {
		t.Fatalf("got repository ID %d, want %d", gotID, wantID)
	}
}

func TestCommitExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCommitExists.json"
	repo := "ivanfetch/ghapitest"
	commitSha := "87d2b8f97a27554711c1eb0d1bb0f8f623a2af25"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
		gotRequestURL := r.RequestURI
		wantRequestURL := fmt.Sprintf("/repos/%s/git/commits/%s", repo, commitSha)
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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
	commitSha := "WillNotExist"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		gotRequestURL := r.RequestURI
		wantRequestURL := fmt.Sprintf("/repos/%s/git/commits/%s", repo, commitSha)
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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

func TestCommitExistsWithIncorrectJSON(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCommitExistsWithIncorrectJSON.json"
	repo := "ivanfetch/ghapitest"
	commitSha := "WillNotMatch"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(testFileName)
		defer f.Close()
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
		gotRequestURL := r.RequestURI
		wantRequestURL := fmt.Sprintf("/repos/%s/git/commits/%s", repo, commitSha)
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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
	if ok {
		t.Fatalf("commit %q found in repository %s, using test data file %s", commitSha, repo, testFileName)
	}

	if err != nil && err.Error() != `incorrect commit sha "DummySha" returned while checking if commit "WillNotMatch" exists` {
		t.Fatal(err)
	}

}

func TestBranchExists(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestBranchExists.json"
	repo := "ivanfetch/ghapitest"
	branch := "test"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
		gotRequestURL := r.RequestURI
		wantRequestURL := fmt.Sprintf("/repos/%s/branches/%s", repo, branch)
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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
		gotRequestURL := r.RequestURI
		wantRequestURL := fmt.Sprintf("/repos/%s/branches/%s", repo, branch)
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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

func TestBranchExistsWithIncorrectJSON(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestBranchExistsWithIncorrectJSON.json"
	repo := "ivanfetch/ghapitest"
	branch := "will-not-match"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(testFileName)
		defer f.Close()
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
		gotRequestURL := r.RequestURI
		wantRequestURL := fmt.Sprintf("/repos/%s/branches/%s", repo, branch)
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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
	if ok {
		t.Fatalf("branch %q found in repository %s, using test data file %s", branch, repo, testFileName)
	}

	if err != nil && err.Error() != fmt.Sprintf(`incorrect name "dummy-name" returned while checking if branch %q exists`, branch) {
		t.Fatal(err)
	}

}

func TestCreateEmptyTreeCommit(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCreateEmptyTreeCommit.json"
	repo := "ivanfetch/ghapitest"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		gotRequestURL := r.RequestURI
		wantRequestURL := fmt.Sprintf("/repos/%s/git/commits", repo)
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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
	if got != want {
		t.Fatalf("got %q, want %q, using test data file %s", got, want, testFileName)
	}
}

func TestCreateEmptyTreeCommitInNonexistentRepository(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCreateEmptyTreeCommitInNonexistentRepository.json"
	repo := "non-existent-repository"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		gotRequestURL := r.RequestURI
		wantRequestURL := fmt.Sprintf("/repos/%s/git/commits", repo)
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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
	if err != nil && err.Error() != fmt.Sprintf("HTTP 404 while creating empty-tree commit for repository %q", repo) {
		t.Fatal(err)
	}
}
func TestCreateBranch(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCreateBranch.json"
	repo := "ivanfetch/ghapitest"
	commitSha := "28ee640a8ce0c22adf3534c7f5971286bfd30642"
	branch := "test"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		gotRequestURL := r.RequestURI
		wantRequestURL := fmt.Sprintf("/repos/%s/git/refs", repo)
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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
func TestCreateBranchWithNonexistentCommit(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCreateBranchWithNonexistentCommit.json"
	repo := "ivanfetch/ghapitest"
	commitSha := "nonexistent-commit-sha"
	branch := "test"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatalf("error copying data from file %s to test HTTP server: %v", testFileName, err)
		}
		gotRequestURL := r.RequestURI
		wantRequestURL := fmt.Sprintf("/repos/%s/git/refs", repo)
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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
	if err == nil {
		t.Fatalf("created branch %q using commit sha %q in repository %s", branch, commitSha, repo)
	}
	if err != nil && err.Error() != fmt.Sprintf("HTTP 422 while creating branch %q using commit %q in repository %q", branch, commitSha, repo) {
		t.Fatal(err)
	}
}

func TestCreateBranchInNonexistentRepository(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/TestCreateBranchInNonexistentRepository.json"
	repo := "nonexistent-repository"
	commitSha := "dummy-sha"
	branch := "test"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		gotRequestURL := r.RequestURI
		wantRequestURL := fmt.Sprintf("/repos/%s/git/refs", repo)
		if wantRequestURL != gotRequestURL {
			t.Errorf("Want %q for Github URL, got %q", wantRequestURL, gotRequestURL)
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
	if err == nil {
		t.Fatalf("created branch %q using commit sha %q in repository %s", branch, commitSha, repo)
	}
	if err != nil && err.Error() != fmt.Sprintf("HTTP 404 while creating branch %q using commit %q in repository %q", branch, commitSha, repo) {
		t.Fatal(err)
	}
}
