package cannect

import (
	"testing"
)

type URIInterfaceTestData struct {
	testcase string
	// input
	uri string
	// want
	scheme string
	path   string
	// err
	err error
}

func testURIInterfaceObj(t *testing.T, data URIInterfaceTestData, uri URI, err error) {
	t.Helper()

	if data.err == nil {
		if err != nil {
			t.Fatalf("Error must not be occurred in this test case but got: %s", err.Error())
		}
	} else {
		if err == nil {
			t.Fatalf("expected error returned but got: %s", err.Error())
		}
		if err.Error() != data.err.Error() {
			t.Fatalf(`expected error is "%s" but got: %s`, data.err.Error(), err.Error())
		}
		return
	}

	if uri.Text() != data.uri {
		t.Errorf("Expected uri text is %s but got: %s", data.uri, uri.Text())
	}

	if uri.Scheme() != data.scheme {
		t.Errorf("Expected scheme is %s but got: %s", data.scheme, uri.Scheme())
	}

	if uri.Path() != data.path {
		t.Errorf("Expected path is %s but got: %s", data.path, uri.Path())
	}
}

func Test_NewFSURI(t *testing.T) {
	t.Parallel()

	data := []URIInterfaceTestData{
		{
			"OK:scheme:file",
			"file://a-bc/d_efg/hi222j.test",
			"file",
			"a-bc/d_efg/hi222j.test",
			nil,
		},
		{
			"NG:scheme:undefined",
			"ng://ng",
			"",
			"",
			InvalidURIError{uri: "ng://ng"},
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.testcase, func(t *testing.T) {
			t.Parallel()

			uri, err := NewFSURI(d.uri)
			testURIInterfaceObj(t, d, uri, err)
		})
	}
}

func Test_NewEnvURI(t *testing.T) {
	t.Parallel()

	data := []URIInterfaceTestData{
		{
			"OK:scheme:env",
			"env://abc_defg_hij",
			"env",
			"abc_defg_hij",
			nil,
		},
		{
			"NG:path:invalid",
			"env://abc_defg_hij-aaa",
			"env",
			"abc_defg_hij",
			InvalidURIError{uri: "env://abc_defg_hij-aaa"},
		},
		{
			"NG:path:invalid",
			"env://abc_defg_hij/aaa",
			"env",
			"abc_defg_hij",
			InvalidURIError{uri: "env://abc_defg_hij/aaa"},
		},
		{
			"NG:scheme:undefined",
			"ng://ng",
			"",
			"",
			InvalidURIError{uri: "ng://ng"},
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.testcase, func(t *testing.T) {
			t.Parallel()

			uri, err := NewEnvURI(d.uri)
			testURIInterfaceObj(t, d, uri, err)
		})
	}
}

func Test_NewGitHubURI(t *testing.T) {
	t.Parallel()

	data := []struct {
		URIInterfaceTestData
		// want
		owenr    string
		repo     string
		repopath string
		ref      string
	}{
		{
			URIInterfaceTestData: URIInterfaceTestData{
				"OK:scheme:github without ref",
				"github:///repos/yuxki/cannect/contents/cmd/cannect/cannect.go",
				"github",
				"/repos/yuxki/cannect/contents/cmd/cannect/cannect.go",
				nil,
			},
			owenr:    "yuxki",
			repo:     "cannect",
			repopath: "cmd/cannect/cannect.go",
			ref:      "",
		},
		{
			URIInterfaceTestData: URIInterfaceTestData{
				"OK:scheme:github with ref",
				"github:///repos/yuxki/cannect/contents/cmd/cannect/cannect.go?ref=v0.1.0",
				"github",
				"/repos/yuxki/cannect/contents/cmd/cannect/cannect.go?ref=v0.1.0",
				nil,
			},
			owenr:    "yuxki",
			repo:     "cannect",
			repopath: "cmd/cannect/cannect.go",
			ref:      "v0.1.0",
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.testcase, func(t *testing.T) {
			t.Parallel()

			uri, err := NewGitHubURI(d.uri)
			testURIInterfaceObj(t, d.URIInterfaceTestData, uri, err)

			if uri.Owner() != d.owenr {
				t.Errorf("Expected owner is %s but got: %s", d.owenr, uri.Owner())
			}
			if uri.Repo() != d.repo {
				t.Errorf("Expected repo is %s but got: %s", d.repo, uri.Repo())
			}
			if uri.RepoPath() != d.repopath {
				t.Errorf("Expected owner is %s but got: %s", d.repopath, uri.RepoPath())
			}
			if uri.Ref() != d.ref {
				t.Errorf("Expected owner is %s but got: %s", d.ref, uri.Ref())
			}
		})
	}
}
