package cannect

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
)

// Order is a struct that retrieves data from its own catalog and writes the
// contents at the destination specified by its own URI.
type Order interface {
	Order(context.Context) error
	setLogger(*log.Logger)
}

type OrderOption func(Order)

func WithOrderLogger(logger *log.Logger) func(Order) {
	return func(o Order) {
		o.setLogger(logger)
	}
}

// FSOrder implements the Order interface. It is responsible for
// placing a CAAsset object in a specific location within the local file system,
// identified by its unique URI path.
type FSOrder struct {
	uri      URI
	catalogs []Catalog
	logger   *log.Logger
}

func NewFSOrder(uri URI, catalogs []Catalog, opts ...OrderOption) *FSOrder {
	order := &FSOrder{
		uri:      uri,
		catalogs: catalogs,
	}

	for _, optF := range opts {
		optF(order)
	}

	return order
}

func (f *FSOrder) Order(ctx context.Context) error {
	if f.logger != nil {
		f.logger.Printf("Ordering: %s", f.uri.Text())
	}

	file, err := os.Create(f.uri.Path())
	if err != nil {
		return err
	}
	defer file.Close()

	for _, catalog := range f.catalogs {
		var buf []byte

		buf, err := catalog.Fetch(ctx)
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

func (f *FSOrder) setLogger(logger *log.Logger) {
	f.logger = logger
}

// EnvOrder implements the Order interface. This is responsible for writing values in
// the format of "export 'key'='value'" to its own file descriptors. It is specifically
// designed to write to environment variables by saving and executing the written file.
type EnvOrder struct {
	uri      URI
	file     *os.File
	catalogs []Catalog
	logger   *log.Logger
}

func NewEnvOrder(uri URI, catalogs []Catalog, file *os.File, opts ...OrderOption) *EnvOrder {
	order := &EnvOrder{
		uri:      uri,
		catalogs: catalogs,
		file:     file,
	}

	for _, optF := range opts {
		optF(order)
	}

	return order
}

func (e *EnvOrder) Order(ctx context.Context) error {
	if e.logger != nil {
		e.logger.Printf("Ordering: %s", e.uri.Text())
	}

	var buf []byte

	for _, catalog := range e.catalogs {
		b, err := catalog.Fetch(ctx)
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

func (e *EnvOrder) setLogger(logger *log.Logger) {
	e.logger = logger
}
