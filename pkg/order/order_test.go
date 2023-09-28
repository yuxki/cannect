package order

import (
	"context"
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuxki/cannect/pkg/asset"
	catalogapi "github.com/yuxki/cannect/pkg/catalog"
	uriapi "github.com/yuxki/cannect/pkg/uri"
)

func testGenCatalogs(t *testing.T) []Catalog {
	t.Helper()

	rootURI, err := uriapi.NewFSURI("file://testdata/root-ca.crt")
	if err != nil {
		t.Fatal(err)
	}
	rootCAAsset := asset.NewCertiricate()
	rootCatalog := catalogapi.NewFSCatalog(rootURI, "", rootCAAsset)

	subURI, err := uriapi.NewFSURI("file://testdata/sub-ca.crt")
	if err != nil {
		t.Fatal(err)
	}
	subCAAsset := asset.NewCertiricate()
	subCatalog := catalogapi.NewFSCatalog(subURI, "", subCAAsset)

	serverURI, err := uriapi.NewFSURI("file://testdata/server.crt")
	if err != nil {
		t.Fatal(err)
	}
	serverCAAsset := asset.NewCertiricate()
	serverCatalog := catalogapi.NewFSCatalog(serverURI, "", serverCAAsset)

	return []Catalog{rootCatalog, subCatalog, serverCatalog}
}

func TestFSOrder_Order(t *testing.T) {
	t.Parallel()

	dir := "testdata"
	dstP := path.Join(dir, "TestFSOrder_Order.out")
	uri, err := uriapi.NewFSURI(fmt.Sprintf("file://%s", dstP))
	if err != nil {
		t.Fatal(err)
	}

	catalogs := testGenCatalogs(t)

	fsOrder := NewFSOrder(uri, catalogs)
	err = fsOrder.Order(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	result, err := os.ReadFile(uri.Path())
	if err != nil {
		t.Fatal(err)
	}

	want, err := os.ReadFile("testdata/chain.crt")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, want) {
		t.Fatal("")
	}
}

func TestEnvOrder_Order(t *testing.T) {
	t.Parallel()

	uri, err := uriapi.NewEnvURI(fmt.Sprintf("env://%s", "abc_efg"))
	if err != nil {
		t.Fatal(err)
	}

	catalogs := testGenCatalogs(t)
	outpath := "testdata/cannect.env.out"

	file, err := os.Create(outpath)
	if err != nil {
		t.Fatal(err)
	}

	envOrder := NewEnvOrder(uri, catalogs, file)
	err = envOrder.Order(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	file.Close()

	result, err := os.ReadFile(outpath)
	if err != nil {
		t.Fatal(err)
	}

	wantpath := "testdata/cannect.env"
	if runtime.GOOS == "windows" {
		wantpath = "testdata/cannect-windows.env"
	}

	want, err := os.ReadFile(wantpath)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(result, want); diff != "" {
		t.Fatal(diff)
	}
}
