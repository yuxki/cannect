package uri

import (
	"errors"
	"fmt"
	"regexp"
)

// ErrInvalidURI is an error that should be used when attempting to use an
// invalid URI.
var ErrInvalidURI = errors.New("invalid uri")

// FSURI represents a URI to a local file.
type FSURI struct {
	text   string
	scheme string
	path   string
}

func NewFSURI(uri string) (FSURI, error) {
	var fsURI FSURI

	reg := regexp.MustCompile("^(file):///?((?:[-_a-z0-9A-Z]+)(?:/[-_a-z0-9A-Z.]+)*)$")
	mt := reg.MatchString(uri)
	if !mt {
		return fsURI, fmt.Errorf(
			"could not match collect File System URI pattern with %s: %w", uri, ErrInvalidURI,
		)
	}

	submt := reg.FindAllStringSubmatch(uri, -1)
	fsURI.text = submt[0][0]
	fsURI.scheme = submt[0][1]
	fsURI.path = submt[0][2]

	return fsURI, nil
}

// Text returns the full URI as a string.
func (u FSURI) Text() string {
	return u.text
}

// Scheme returns the part of scheme in URI.
func (u FSURI) Scheme() string {
	return u.scheme
}

// Scheme returns the part of path in URI.
func (u FSURI) Path() string {
	return u.path
}

type EnvURI struct {
	text   string
	scheme string
	path   string
}

// FSURI represents a URI for an environment variable.
func NewEnvURI(uri string) (EnvURI, error) {
	var eURI EnvURI

	reg := regexp.MustCompile("^(env)://([_a-z0-9A-Z]+)$")
	mt := reg.MatchString(uri)
	if !mt {
		return eURI, fmt.Errorf(
			"could not match collect Env URI pattern with %s: %w", uri, ErrInvalidURI,
		)
	}

	submt := reg.FindAllStringSubmatch(uri, -1)
	eURI.text = submt[0][0]
	eURI.scheme = submt[0][1]
	eURI.path = submt[0][2]

	return eURI, nil
}

// Text returns the full URI as a string.
func (u EnvURI) Text() string {
	return u.text
}

// Scheme returns the part of scheme in URI.
func (u EnvURI) Scheme() string {
	return u.scheme
}

// Scheme returns the part of path in URI.
func (u EnvURI) Path() string {
	return u.path
}

type GitHubURI struct {
	text     string
	scheme   string
	path     string
	owner    string
	repo     string
	repopath string
	ref      string
}

// FSURI represents a URI for an GitHub Get Repository Contents API.
func NewGitHubURI(uri string) (GitHubURI, error) {
	var ghURI GitHubURI

	word := "[-_a-zA-Z0-9.]"
	reg := regexp.MustCompile(
		fmt.Sprintf(`^(github)://(/repos/(%s+)/(%s+)/contents/(%s+(?:/%s+)*)(?:\?ref=(%s+))?)$`,
			word, word, word, word, word,
		),
	)
	mt := reg.MatchString(uri)
	if !mt {
		return ghURI, fmt.Errorf(
			"could not match collect GitHub URI pattern with %s: %w", uri, ErrInvalidURI,
		)
	}

	submt := reg.FindAllStringSubmatch(uri, -1)
	ghURI.text = submt[0][0]
	ghURI.scheme = submt[0][1]
	ghURI.path = submt[0][2]
	ghURI.owner = submt[0][3]
	ghURI.repo = submt[0][4]
	ghURI.repopath = submt[0][5]
	ghURI.ref = submt[0][6]

	return ghURI, nil
}

func (u GitHubURI) Text() string {
	return u.text
}

func (u GitHubURI) Scheme() string {
	return u.scheme
}

func (u GitHubURI) Path() string {
	return u.path
}

func (u GitHubURI) Owner() string {
	return u.owner
}

func (u GitHubURI) Repo() string {
	return u.repo
}

func (u GitHubURI) RepoPath() string {
	return u.repopath
}

func (u GitHubURI) Ref() string {
	return u.ref
}
