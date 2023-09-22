package cannect

import (
	"testing"
)

func TestCertiricate(t *testing.T) {
	t.Parallel()

	asset := NewCertiricate()
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

	asset := NewPrivateKey()
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

	asset := NewEncryptedPrivateKey()
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

	asset := NewCRL()
	err := asset.CheckContent([]byte("-----BEGIN X509 CRL-----"))
	if err != nil {
		t.Fatal(err)
	}

	err = asset.CheckContent([]byte("-----BEGIN CERTIFICATE-----"))
	if err == nil {
		t.Fatal("must cause verify error")
	}
}
