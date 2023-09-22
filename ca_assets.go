package cannect

import (
	"fmt"
	"regexp"
)

const (
	CertCategory       = "certificate"
	PrivKeyCategory    = "privateKey"
	EncPrivKeyCategory = "encPrivateKey"
	CRLCategory        = "CRL"
)

type InvalidCAAssetError struct {
	category string
	reason   string
}

func (e InvalidCAAssetError) Error() string {
	return fmt.Sprintf("may not be %s or not supported format: %s", e.category, e.reason)
}

// CAAsset represents the different types of assets within a private CA.
// It is responsible for keeping track of the assets and indicating which
// assets are associated with structures that have the CAAsset as a member.
type CAAsset interface {
	// Verify that the content in the asset as expected.
	CheckContent([]byte) error
}

type Certiricate struct{}

func NewCertiricate() Certiricate {
	return Certiricate{}
}

func (c Certiricate) CheckContent(content []byte) error {
	ok, err := regexp.Match("-----BEGIN CERTIFICATE-----", content)
	if err != nil {
		return err
	}

	if !ok {
		return InvalidCAAssetError{
			category: CertCategory,
			reason:   `not contain "-----BEGIN CERTIFICATE-----" pattern`,
		}
	}

	return nil
}

type PrivateKey struct{}

func NewPrivateKey() PrivateKey {
	return PrivateKey{}
}

func (p PrivateKey) CheckContent(content []byte) error {
	ok, err := regexp.Match("PRIVATE KEY-----", content)
	if err != nil {
		return err
	}

	if !ok {
		return InvalidCAAssetError{
			category: PrivKeyCategory,
			reason:   `not contain "PRIVATE KEY-----" pattern`,
		}
	}

	ok, err = regexp.Match("-----BEGIN ENCRYPTED", content)
	if err != nil {
		return err
	}
	if ok {
		return InvalidCAAssetError{
			category: PrivKeyCategory,
			reason:   `contain "-----BEGIN ENCRYPTED" pattern`,
		}
	}

	return nil
}

type EncryptedPrivateKey struct{}

func NewEncryptedPrivateKey() EncryptedPrivateKey {
	return EncryptedPrivateKey{}
}

func (e EncryptedPrivateKey) CheckContent(content []byte) error {
	ok, err := regexp.Match("-----BEGIN ENCRYPTED PRIVATE KEY-----", content)
	if err != nil {
		return err
	}

	if !ok {
		return InvalidCAAssetError{
			category: EncPrivKeyCategory,
			reason:   `not contain "-----BEGIN ENCRYPTED PRIVATE KEY-----" pattern`,
		}
	}

	return nil
}

type CRL struct{}

func NewCRL() CRL {
	return CRL{}
}

func (c CRL) CheckContent(content []byte) error {
	ok, err := regexp.Match("-----BEGIN X509 CRL-----", content)
	if err != nil {
		return err
	}

	if !ok {
		return InvalidCAAssetError{
			category: EncPrivKeyCategory,
			reason:   `not contain "-----BEGIN X509 CRL-----" pattern`,
		}
	}

	return nil
}
