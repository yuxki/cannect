package uri

import (
	"errors"
	"testing"
)

type uriCommonTestData struct {
	testcase string
	// input
	uri string
	// want
	scheme string
	path   string
	// err
	err error
}

func testCommonTestData(t *testing.T, data uriCommonTestData, text, scheme, path string, err error) {
	t.Helper()

	if data.err == nil {
		if err != nil {
			t.Fatalf("Error must not be occurred in this test case but got: %s", err.Error())
		}
	} else {
		if err == nil {
			t.Fatalf("expected error returned but got: %s", err.Error())
		}
		if !errors.Is(err, data.err) {
			t.Fatalf(`expected error is "%#v" but got: %#v`, data.err, err)
		}
		return
	}

	if text != data.uri {
		t.Errorf("Expected uri text is %s but got: %s", data.uri, text)
	}

	if scheme != data.scheme {
		t.Errorf("Expected scheme is %s but got: %s", data.scheme, scheme)
	}

	if path != data.path {
		t.Errorf("Expected path is %s but got: %s", data.path, path)
	}
}

func Test_NewFSURI(t *testing.T) {
	t.Parallel()

	data := []uriCommonTestData{
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
			ErrInvalidURI,
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.testcase, func(t *testing.T) {
			t.Parallel()

			uri, err := NewFSURI(d.uri)
			testCommonTestData(t, d, uri.Text(), uri.Scheme(), uri.Path(), err)
		})
	}
}

func Test_NewEnvURI(t *testing.T) {
	t.Parallel()

	data := []uriCommonTestData{
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
			ErrInvalidURI,
		},
		{
			"NG:path:invalid",
			"env://abc_defg_hij/aaa",
			"env",
			"abc_defg_hij",
			ErrInvalidURI,
		},
		{
			"NG:scheme:undefined",
			"ng://ng",
			"",
			"",
			ErrInvalidURI,
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.testcase, func(t *testing.T) {
			t.Parallel()

			uri, err := NewEnvURI(d.uri)
			testCommonTestData(t, d, uri.Text(), uri.Scheme(), uri.Path(), err)
		})
	}
}

func Test_NewGitHubURI(t *testing.T) {
	t.Parallel()

	data := []struct {
		uriCommonTestData
		// want
		owenr    string
		repo     string
		repopath string
		ref      string
	}{
		{
			uriCommonTestData: uriCommonTestData{
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
			uriCommonTestData: uriCommonTestData{
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
			testCommonTestData(t, d.uriCommonTestData, uri.Text(), uri.Scheme(), uri.Path(), err)

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

func Test_NewS3URI(t *testing.T) {
	t.Parallel()

	data := []struct {
		uriCommonTestData
		// want
		bucket string
		key    string
		err    error
	}{
		{
			uriCommonTestData: uriCommonTestData{
				"OK:without logical hierarchy",
				"s3://fooBucket/barKey",
				"s3",
				"fooBucket/barKey",
				nil,
			},
			bucket: "fooBucket",
			key:    "barKey",
			err:    nil,
		},
		{
			uriCommonTestData: uriCommonTestData{
				"OK:with logical hierarchy",
				"s3://fooBucket/fooKey/barKey/bazKey",
				"s3",
				"fooBucket/fooKey/barKey/bazKey",
				nil,
			},
			bucket: "fooBucket",
			key:    "fooKey/barKey/bazKey",
			err:    nil,
		},
		{
			uriCommonTestData: uriCommonTestData{
				"NG:invalid URI",
				"ng://fooBucket/fooKey/barKey/bazKey",
				"ng",
				"fooBucket/fooKey/barKey/bazKey",
				nil,
			},
			bucket: "fooBucket",
			key:    "fooKey/barKey/bazKey",
			err:    ErrInvalidURI,
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.testcase, func(t *testing.T) {
			t.Parallel()

			uri, err := NewS3URI(d.uri)
			if err != nil {
				if !errors.Is(d.err, errors.Unwrap(err)) {
					t.Fatalf("Expected error is %#v but got: %#v", d.err, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected error is nil  but got: %#v", err)
			}

			testCommonTestData(t, d.uriCommonTestData, uri.Text(), uri.Scheme(), uri.Path(), err)

			if uri.Bucket() != d.bucket {
				t.Errorf("Expected bucket is %s but got: %s", d.bucket, uri.Bucket())
			}
			if uri.Key() != d.key {
				t.Errorf("Expected key is %s but got: %s", d.key, uri.Key())
			}
		})
	}
}
