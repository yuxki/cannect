package order

import (
	"context"
	"fmt"
	"os"
	"runtime"

	uriapi "github.com/yuxki/cannect/pkg/uri"
)

type Logger interface {
	// Log about provided URI.
	Log(uriText string)
}

// Catalog represents catalog of assets held by Private CA.
type Catalog interface {
	// Fetch retrieves data based on the information of its own URI.
	Fetch(context.Context) ([]byte, error)
}

// FSOrder implements the Order interface. It is responsible for
// placing a CAAsset object in a specific location within the local file system,
// identified by its unique URI path.
type FSOrder struct {
	uri      uriapi.FSURI
	catalogs []Catalog
	l        Logger
}

func NewFSOrder(uri uriapi.FSURI, catalogs []Catalog) *FSOrder {
	order := &FSOrder{
		uri:      uri,
		catalogs: catalogs,
	}

	return order
}

func (f *FSOrder) Order(ctx context.Context) error {
	if f.l != nil {
		f.l.Log(f.uri.Text())
	}

	file, err := os.Create(f.uri.Path())
	if err != nil {
		return err
	}
	defer file.Close()

	for idx := range f.catalogs {
		var buf []byte

		buf, err := f.catalogs[idx].Fetch(ctx)
		if err != nil {
			return err
		}

		_, err = file.Write(buf)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FSOrder) WithLogger(l Logger) *FSOrder {
	f.l = l
	return f
}

// EnvOrder implements the Order interface. This is responsible for writing values in
// the format of "export 'key'='value'" to its own file descriptors. It is specifically
// designed to write to environment variables by saving and executing the written file.
type EnvOrder struct {
	uri      uriapi.EnvURI
	file     *os.File
	catalogs []Catalog
	l        Logger
}

func NewEnvOrder(uri uriapi.EnvURI, catalogs []Catalog, file *os.File) *EnvOrder {
	order := &EnvOrder{
		uri:      uri,
		catalogs: catalogs,
		file:     file,
	}

	return order
}

func (e *EnvOrder) Order(ctx context.Context) error {
	if e.l != nil {
		e.l.Log(e.uri.Text())
	}

	var buf []byte

	for idx := range e.catalogs {
		b, err := e.catalogs[idx].Fetch(ctx)
		if err != nil {
			return err
		}

		buf = append(buf, b...)
	}

	nl := "\n"
	if runtime.GOOS == "windows" {
		nl = "\r\n"
	}

	_, err := e.file.WriteString(fmt.Sprintf("export '%s'='%s'%s", e.uri.Path(), string(buf), nl))
	if err != nil {
		return err
	}

	return nil
}

func (e *EnvOrder) WithLogger(l Logger) *EnvOrder {
	e.l = l
	return e
}
