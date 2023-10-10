package catalog

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-github/v55/github"
	uriapi "github.com/yuxki/cannect/pkg/uri"
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
	uri     uriapi.FSURI
	alias   string
	checker AssetChecker
	logger  Logger
}

func NewFSCatalog(uri uriapi.FSURI, alias string, checker AssetChecker) *FSCatalog {
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
	uri     uriapi.GitHubURI
	alias   string
	checker AssetChecker
	logger  Logger
}

func NewGitHubCatalog(uri uriapi.GitHubURI, alias string, checker AssetChecker) *GitHubCatalog {
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

// S3Catalog is an implementation of the Catalog interface.
// It is responsible for fetching assets held by a Private CA from a AWS S3.
// It uses the AWS S3 GetObject API for this purpose.
type S3Catalog struct {
	uri     uriapi.S3URI
	alias   string
	checker AssetChecker
	logger  Logger
}

// The Fetch function utilizes the GetObjcet API in AWS S3. It
// requires the usage of an environment variable "AWS_ACCESS_KEY_ID" and
// "AWS_SECRET_ACCESS_KEY", "AWS_DEFAULT_REGION", to authorize the request.
// The function then returns the content of the file as a byte slice.
func (s *S3Catalog) Fetch(ctx context.Context) ([]byte, error) {
	if s.logger != nil {
		s.logger.Log(s.uri.Text())
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Bucket()),
		Key:    aws.String(s.uri.Key()),
	})
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()

	buf, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func NewS3Catalog(uri uriapi.S3URI, alias string, checker AssetChecker) *S3Catalog {
	ctlg := &S3Catalog{
		uri:     uri,
		alias:   alias,
		checker: checker,
	}

	return ctlg
}

func (s *S3Catalog) WithLogger(l Logger) *S3Catalog {
	s.logger = l
	return s
}
