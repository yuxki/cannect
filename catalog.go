package cannect

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
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
	// GetAlias returns an alias used to associate itself with the Order interface.
	GetAlias() string
	// GetAsset retrieves and returns an instance of the CAAsset.
	GetCAAsset() CAAsset
	setLogger(*log.Logger)
}

type CatalogOption func(Catalog)

func WithCatalogLogger(logger *log.Logger) func(Catalog) {
	return func(c Catalog) {
		c.setLogger(logger)
	}
}

// FSCatalog is an implementation of the Catalog interface. It is responsible for
// fetching assets held by a Private CA from the local filesystem.
type FSCatalog struct {
	uri    URI
	alias  string
	asset  CAAsset
	logger *log.Logger
}

func NewFSCatalog(uri URI, alias string, asset CAAsset, opts ...CatalogOption) *FSCatalog {
	ctlg := &FSCatalog{
		uri:   uri,
		alias: alias,
		asset: asset,
	}

	for _, optF := range opts {
		optF(ctlg)
	}

	return ctlg
}

// Use the os package to open the file with the provided file path
// and return the content of the file as a byte slice.
func (f *FSCatalog) Fetch(ctx context.Context) ([]byte, error) {
	if f.logger != nil {
		f.logger.Printf("Catalog Fetching: %s", f.uri.Text())
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

func (f *FSCatalog) GetAlias() string {
	return f.alias
}

func (f *FSCatalog) GetCAAsset() CAAsset {
	return f.asset
}

func (f *FSCatalog) setLogger(logger *log.Logger) {
	f.logger = logger
}

// GitHubCatalog is an implementation of the Catalog interface.
// It is responsible for fetching assets held by a Private CA from a GitHub repository.
// It uses the GitHub Get Repository Content API for this purpose.
type GitHubCatalog struct {
	uri    GitHubURI
	alias  string
	asset  CAAsset
	logger *log.Logger
}

func NewGitHubCatalog(uri GitHubURI, alias string, asset CAAsset, opts ...CatalogOption) *GitHubCatalog {
	ctlg := &GitHubCatalog{
		uri:   uri,
		alias: alias,
		asset: asset,
	}

	for _, optF := range opts {
		optF(ctlg)
	}

	return ctlg
}

// The Fetch function utilizes the Get repository content API in GitHub. It
// requires the usage of an environment variable called "GITHUB_TOKEN" to authorize the
// request. The function then returns the content of the file as a byte slice.
func (g *GitHubCatalog) Fetch(ctx context.Context) ([]byte, error) {
	if g.logger != nil {
		g.logger.Printf("Catalog Fetching: %s", g.uri.Text())
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

func (g *GitHubCatalog) GetAlias() string {
	return g.alias
}

func (g *GitHubCatalog) GetCAAsset() CAAsset {
	return g.asset
}

func (g *GitHubCatalog) setLogger(logger *log.Logger) {
	g.logger = logger
}
