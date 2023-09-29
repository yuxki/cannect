package main

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	file, err := os.Open("testdata/test_catalog_order.json")
	if err != nil {
		t.Fatal(err)
	}

	want := CAnnectJSON{
		Catalogs: []CatalogJSON{
			{
				Alias:    "root-ca.crt",
				URI:      "file://testdata/root-ca.crt",
				Category: "certificate",
			},
			{
				Alias:    "sub-ca.crt",
				URI:      "file://testdata/sub-ca.crt",
				Category: "certificate",
			},
			{
				Alias:    "server.crt",
				URI:      "file://testdata/server.crt",
				Category: "certificate",
			},
		},
		Orders: []OrderJSON{
			{
				CatalogAliases: []string{
					"root-ca.crt",
				},
				URI: "file://testdata/test-root-ca.crt.crt",
			},
			{
				CatalogAliases: []string{
					"sub-ca.crt",
				},
				URI: "file://testdata/test-sub-ca.crt.crt",
			},
			{
				CatalogAliases: []string{
					"root-ca.crt",
					"sub-ca.crt",
					"server.crt",
				},
				URI: "file://testdata/test-server.crt.crt",
			},
		},
	}

	jsn, err := unmarshal(file)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(jsn, want); diff != "" {
		t.Error(diff)
	}
}

func TestUnmarshalBoth(t *testing.T) {
	t.Parallel()

	want := CAnnectJSON{
		Catalogs: []CatalogJSON{
			{
				Alias:    "root-ca.crt",
				URI:      "file://testdata/root-ca.crt",
				Category: "certificate",
			},
			{
				Alias:    "sub-ca.crt",
				URI:      "file://testdata/sub-ca.crt",
				Category: "certificate",
			},
			{
				Alias:    "server.crt",
				URI:      "file://testdata/server.crt",
				Category: "certificate",
			},
		},
		Orders: []OrderJSON{
			{
				CatalogAliases: []string{
					"root-ca.crt",
				},
				URI: "file://testdata/test-root-ca.crt.crt",
			},
			{
				CatalogAliases: []string{
					"sub-ca.crt",
				},
				URI: "file://testdata/test-sub-ca.crt.crt",
			},
			{
				CatalogAliases: []string{
					"root-ca.crt",
					"sub-ca.crt",
					"server.crt",
				},
				URI: "file://testdata/test-server.crt.crt",
			},
		},
	}

	catalogFile, err := os.Open("testdata/test_catalog.json")
	if err != nil {
		t.Fatal(err)
	}

	orderFile, err := os.Open("testdata/test_order.json")
	if err != nil {
		t.Fatal(err)
	}

	jsn, err := unmarshalBoth(catalogFile, orderFile)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(jsn, want); diff != "" {
		t.Error(diff)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	data := []struct {
		testdata string
		// input
		jsn CAnnectJSON
		// want
		err error
	}{
		{
			"OK",
			CAnnectJSON{
				Catalogs: []CatalogJSON{
					{
						Alias:    "root-ca.crt",
						URI:      "file://testdata/root-ca.crt",
						Category: "certificate",
					},
				},
				Orders: []OrderJSON{
					{
						CatalogAliases: []string{
							"root-ca.crt",
						},
						URI: "file://testdata/test-root-ca.crt.crt",
					},
				},
			},
			nil,
		},
		{
			"NG:Duplicated Order",
			CAnnectJSON{
				Catalogs: []CatalogJSON{
					{
						Alias:    "root-ca.crt",
						URI:      "file://testdata/root-ca.crt",
						Category: "certificate",
					},
				},
				Orders: []OrderJSON{
					{
						CatalogAliases: []string{
							"root-ca.crt",
						},
						URI: "file://testdata/test-root-ca.crt.crt",
					},
					{
						CatalogAliases: []string{
							"root-ca.crt",
						},
						URI: "file://testdata/test-root-ca.crt.crt",
					},
				},
			},
			errOrderURIDuplicated,
		},
		{
			"NG:Undefined Alias",
			CAnnectJSON{
				Catalogs: []CatalogJSON{
					{
						Alias:    "root-ca.crt",
						URI:      "file://testdata/root-ca.crt",
						Category: "certificate",
					},
				},
				Orders: []OrderJSON{
					{
						CatalogAliases: []string{
							"sub-ca.crt",
						},
						URI: "file://testdata/test-sub-ca.crt.crt",
					},
				},
			},
			errUndefinedAlias,
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.testdata, func(t *testing.T) {
			t.Parallel()

			err := validate(d.jsn)

			if d.err == nil {
				if err != nil {
					t.Fatalf("Expected no error but got: %s", err.Error())
				}
				return
			}

			if !errors.Is(err, d.err) {
				t.Fatalf("Expected %#v error but got: %#v", d.err, err)
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	jsn := CAnnectJSON{
		Catalogs: []CatalogJSON{
			{
				Alias:    "root-ca.crt",
				URI:      "file://testdata/root-ca.crt",
				Category: "certificate",
			},
			{
				Alias:    "sub-ca.crt",
				URI:      "file://testdata/sub-ca.crt",
				Category: "certificate",
			},
			{
				Alias:    "server.crt",
				URI:      "file://testdata/server.crt",
				Category: "certificate",
			},
		},
		Orders: []OrderJSON{
			{
				CatalogAliases: []string{
					"root-ca.crt",
				},
				URI: "file://testdata/test-root-ca.out",
			},
			{
				CatalogAliases: []string{
					"sub-ca.crt",
				},
				URI: "file://testdata/test-sub-ca.out",
			},
			{
				CatalogAliases: []string{
					"root-ca.crt",
					"sub-ca.crt",
					"server.crt",
				},
				URI: "file://testdata/test-chain.out",
			},
		},
	}

	cfg := runConfig{EnvOut: "./envout.env", ConLimit: 5}
	logger := log.New(os.Stdout, "", log.LstdFlags)
	err := run(context.TODO(), jsn, cfg, logger)
	if err != nil {
		t.Fatal(err)
	}

	rootWant, err := os.ReadFile("testdata/root-ca.crt")
	if err != nil {
		t.Fatal(err)
	}
	rootResult, err := os.ReadFile("testdata/test-root-ca.out")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(rootResult, rootWant); diff != "" {
		t.Error(diff)
	}

	subWant, err := os.ReadFile("testdata/sub-ca.crt")
	if err != nil {
		t.Fatal(err)
	}
	subResult, err := os.ReadFile("testdata/test-sub-ca.out")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(subResult, subWant); diff != "" {
		t.Error(diff)
	}

	chainWant, err := os.ReadFile("testdata/chain.crt")
	if err != nil {
		t.Fatal(err)
	}
	chainResult, err := os.ReadFile("testdata/test-chain.out")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(chainResult, chainWant); diff != "" {
		t.Error(diff)
	}
}
