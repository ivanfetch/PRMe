package prme_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"prme"
	"testing"
)

/* This test is (likely) replaced by the test below it.
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
*/

func TestConnectToGithubAPI(t *testing.T) {
	t.Parallel()

	testFileName := "testdata/github-docs-repo.json"

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
