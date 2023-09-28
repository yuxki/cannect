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
	"github.com/yuxki/cannect/pkg/catalog"
	"github.com/yuxki/cannect/pkg/uri"
)

func testGenCatalogs(t *testing.T) []Catalog {
	t.Helper()

	rootURI, err := uri.NewFSURI("file://testdata/root-ca.crt")
	if err != nil {
		t.Fatal(err)
	}
	rootCAAsset := asset.NewCertiricate()
	rootCatalog := catalog.NewFSCatalog(rootURI, "", rootCAAsset)

	subURI, err := uri.NewFSURI("file://testdata/sub-ca.crt")
	if err != nil {
		t.Fatal(err)
	}
	subCAAsset := asset.NewCertiricate()
	subCatalog := catalog.NewFSCatalog(subURI, "", subCAAsset)

	serverURI, err := uri.NewFSURI("file://testdata/server.crt")
	if err != nil {
		t.Fatal(err)
	}
	serverCAAsset := asset.NewCertiricate()
	serverCatalog := catalog.NewFSCatalog(serverURI, "", serverCAAsset)

	return []Catalog{rootCatalog, subCatalog, serverCatalog}
}

func TestFSOrder_Order(t *testing.T) {
	t.Parallel()

	dir := "testdata"
	dstP := path.Join(dir, "TestFSOrder_Order.out")
	fsURI, err := uri.NewFSURI(fmt.Sprintf("file://%s", dstP))
	if err != nil {
		t.Fatal(err)
	}

	catalogs := testGenCatalogs(t)

	fsOrder := NewFSOrder(fsURI, catalogs)
	err = fsOrder.Order(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	result, err := os.ReadFile(fsURI.Path())
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

	uri, err := uri.NewEnvURI(fmt.Sprintf("env://%s", "abc_efg"))
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
