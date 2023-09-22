package cannect

import (
	"context"
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func testGenCatalogs(t *testing.T) []Catalog {
	t.Helper()

	rootURI, err := NewFSURI("file://testdata/root-ca.crt")
	if err != nil {
		t.Fatal(err)
	}
	rootCAAsset := NewCertiricate(rootURI)
	rootCatalog := NewFSCatalog(rootURI, "", rootCAAsset)

	subURI, err := NewFSURI("file://testdata/sub-ca.crt")
	if err != nil {
		t.Fatal(err)
	}
	subCAAsset := NewCertiricate(subURI)
	subCatalog := NewFSCatalog(subURI, "", subCAAsset)

	serverURI, err := NewFSURI("file://testdata/server.crt")
	if err != nil {
		t.Fatal(err)
	}
	serverCAAsset := NewCertiricate(serverURI)
	serverCatalog := NewFSCatalog(serverURI, "", serverCAAsset)

	return []Catalog{rootCatalog, subCatalog, serverCatalog}
}

func TestFSOrder_Order(t *testing.T) {
	t.Parallel()

	dir := "testdata"
	dstP := path.Join(dir, "TestFSOrder_Order.out")
	uri, err := NewFSURI(fmt.Sprintf("file://%s", dstP))
	if err != nil {
		t.Fatal(err)
	}

	catalogs := testGenCatalogs(t)

	fsOrder := NewFSOrder(uri, catalogs)
	err = fsOrder.Order(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	result, err := os.ReadFile(uri.path)
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

	uri, err := NewEnvURI(fmt.Sprintf("env://%s", "abc_efg"))
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
