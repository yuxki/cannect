package catalog

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/google/go-github/v55/github"
	"github.com/yuxki/cannect/pkg/uri"
)

type Logger interface {
	// Log about provided URI.
	Log(uriText string)
}

type AssetChecker interface {
	// Verify that the content in the asset as expected.
	CheckContent([]byte) error
}

// FetchError is used to represent an error that occurs when fetching a
// data fails.
type FetchError struct {
	uri    string
	reason string
}

func (e FetchError) Error() string {
	return fmt.Sprintf("fetch failed at %s: %s", e.uri, e.reason)
}

// FSCatalog is an implementation of the Catalog interface. It is responsible for
// fetching assets held by a Private CA from the local filesystem.
type FSCatalog struct {
	uri     uri.FSURI
	alias   string
	checker AssetChecker
	logger  Logger
}

func NewFSCatalog(uri uri.FSURI, alias string, checker AssetChecker) *FSCatalog {
	ctlg := &FSCatalog{
		uri:     uri,
		alias:   alias,
		checker: checker,
	}

	return ctlg
}

// Use the os package to open the file with the provided file path
// and return the content of the file as a byte slice.
func (f *FSCatalog) Fetch(ctx context.Context) ([]byte, error) {
	if f.logger != nil {
		f.logger.Log(f.uri.Text())
	}

	buf, err := os.ReadFile(f.uri.Path())
	if err != nil {
		return nil, err
	}

	err = f.checker.CheckContent(buf)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", f.uri.Path(), err)
	}

	return buf, nil
}

func (f *FSCatalog) WithLogger(l Logger) *FSCatalog {
	f.logger = l
	return f
}

// GitHubCatalog is an implementation of the Catalog interface.
// It is responsible for fetching assets held by a Private CA from a GitHub repository.
// It uses the GitHub Get Repository Content API for this purpose.
type GitHubCatalog struct {
	uri     uri.GitHubURI
	alias   string
	checker AssetChecker
	logger  Logger
}

func NewGitHubCatalog(uri uri.GitHubURI, alias string, checker AssetChecker) *GitHubCatalog {
	ctlg := &GitHubCatalog{
		uri:     uri,
		alias:   alias,
		checker: checker,
	}

	return ctlg
}

// The Fetch function utilizes the Get repository content API in GitHub. It
// requires the usage of an environment variable called "GITHUB_TOKEN" to authorize the
// request. The function then returns the content of the file as a byte slice.
func (g *GitHubCatalog) Fetch(ctx context.Context) ([]byte, error) {
	if g.logger != nil {
		g.logger.Log(g.uri.Text())
	}

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))
	content, _, _, err := client.Repositories.GetContents(ctx,
		g.uri.Owner(),
		g.uri.Repo(),
		g.uri.RepoPath(),
		&github.RepositoryContentGetOptions{
			Ref: g.uri.Ref(),
		},
	)
	if err != nil {
		return nil, err
	}

	if *content.Type != "file" {
		return nil, FetchError{uri: g.uri.Text(), reason: "Only support file type."}
	}

	buf, err := base64.URLEncoding.DecodeString(*content.Content)
	if err != nil {
		return nil, err
	}

	err = g.checker.CheckContent(buf)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", g.uri.Path(), err)
	}

	return buf, nil
}

func (g *GitHubCatalog) WithLogger(l Logger) *GitHubCatalog {
	g.logger = l
	return g
}
