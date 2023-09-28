package cannect

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/google/go-github/v55/github"
)

// FetchError is used to represent an error that occurs when fetching a
// data fails.
type FetchError struct {
	uri    string
	reason string
}

func (e FetchError) Error() string {
	return fmt.Sprintf("fetch failed at %s: %s", e.uri, e.reason)
}

// Catalog represents catalog of assets held by Private CA.
type Catalog interface {
	// Fetch retrieves data based on the information of its own URI.
	Fetch(context.Context) ([]byte, error)
}

type CatalogOption func(Catalog)

// FSCatalog is an implementation of the Catalog interface. It is responsible for
// fetching assets held by a Private CA from the local filesystem.
type FSCatalog struct {
	uri    URI
	alias  string
	asset  CAAsset
	logger Logger
}

func NewFSCatalog(uri URI, alias string, asset CAAsset) *FSCatalog {
	ctlg := &FSCatalog{
		uri:   uri,
		alias: alias,
		asset: asset,
	}

	return ctlg
}

// Use the os package to open the file with the provided file path
// and return the content of the file as a byte slice.
func (f *FSCatalog) Fetch(ctx context.Context) ([]byte, error) {
	if f.logger != nil {
		f.logger.Log(f.uri)
	}

	buf, err := os.ReadFile(f.uri.Path())
	if err != nil {
		return nil, err
	}

	err = f.asset.CheckContent(buf)
	if err != nil {
		return nil, err
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
	uri    GitHubURI
	alias  string
	asset  CAAsset
	logger Logger
}

func NewGitHubCatalog(uri GitHubURI, alias string, asset CAAsset) *GitHubCatalog {
	ctlg := &GitHubCatalog{
		uri:   uri,
		alias: alias,
		asset: asset,
	}

	return ctlg
}

// The Fetch function utilizes the Get repository content API in GitHub. It
// requires the usage of an environment variable called "GITHUB_TOKEN" to authorize the
// request. The function then returns the content of the file as a byte slice.
func (g *GitHubCatalog) Fetch(ctx context.Context) ([]byte, error) {
	if g.logger != nil {
		g.logger.Log(g.uri)
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

	err = g.asset.CheckContent(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (g *GitHubCatalog) WithLogger(l Logger) *GitHubCatalog {
	g.logger = l
	return g
}
