package cannect

import (
	"testing"
)

func testDummyURI(t *testing.T) FSURI {
	t.Helper()

	uri, err := NewFSURI("file://a-bc/d_efg/hi222j.test")
	if err != nil {
		t.Fatal(err)
	}

	return uri
}

func TestCertiricate(t *testing.T) {
	t.Parallel()

	uri := testDummyURI(t)
	asset := NewCertiricate(uri)
	err := asset.CheckContent([]byte("-----BEGIN CERTIFICATE-----"))
	if err != nil {
		t.Fatal(err)
	}

	err = asset.CheckContent([]byte("-----BEGIN X509 CRL-----"))
	if err == nil {
		t.Fatal("must cause verify error")
	}
}

func TestPrivateKey(t *testing.T) {
	t.Parallel()

	uri := testDummyURI(t)
	asset := NewPrivateKey(uri)
	err := asset.CheckContent([]byte("-----BEGIN EC PRIVATE KEY-----"))
	if err != nil {
		t.Fatal(err)
	}

	err = asset.CheckContent([]byte("-----BEGIN ENCRYPTED PRIVATE KEY-----"))
	if err == nil {
		t.Fatal("must cause verify error")
	}
}

func TestEncryptedPrivateKey(t *testing.T) {
	t.Parallel()

	uri := testDummyURI(t)
	asset := NewEncryptedPrivateKey(uri)
	err := asset.CheckContent([]byte("-----BEGIN ENCRYPTED PRIVATE KEY-----"))
	if err != nil {
		t.Fatal(err)
	}

	err = asset.CheckContent([]byte("-----BEGIN EC PRIVATE KEY-----"))
	if err == nil {
		t.Fatal("must cause verify error")
	}
}

func TestCRL(t *testing.T) {
	t.Parallel()

	uri := testDummyURI(t)
	asset := NewCRL(uri)
	err := asset.CheckContent([]byte("-----BEGIN X509 CRL-----"))
	if err != nil {
		t.Fatal(err)
	}

	err = asset.CheckContent([]byte("-----BEGIN CERTIFICATE-----"))
	if err == nil {
		t.Fatal("must cause verify error")
	}
}
