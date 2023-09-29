package asset

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	CertCategory       = "certificate"
	PrivKeyCategory    = "privateKey"
	EncPrivKeyCategory = "encPrivateKey"
	CRLCategory        = "CRL"
)

// ErrUnexpectedCAAsset means fetched content of CA asset is not unexpected
// or not supported in cannect.
var ErrUnexpectedCAAsset = errors.New(
	"may not have expected content or not be in not supported format",
)

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
		return fmt.Errorf(
			`"-----BEGIN CERTIFICATE-----" pattern may be contained in %s: %w`,
			CertCategory, ErrUnexpectedCAAsset,
		)
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
		return fmt.Errorf(
			`"PRIVATE KEY-----" pattern may be contained in %s: %w`,
			PrivKeyCategory, ErrUnexpectedCAAsset,
		)
	}

	ok, err = regexp.Match("-----BEGIN ENCRYPTED", content)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(
			`"-----BEGIN ENCRYPTED" pattern may NOT be contained in %s: %w`,
			PrivKeyCategory, ErrUnexpectedCAAsset,
		)
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
		return fmt.Errorf(
			`"-----BEGIN ENCRYPTED PRIVATE KEY-----" pattern may be contained in %s: %w`,
			EncPrivKeyCategory, ErrUnexpectedCAAsset,
		)
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
		return fmt.Errorf(
			`"-----BEGIN X509 CRL-----" pattern may be contained in %s: %w`,
			CRLCategory, ErrUnexpectedCAAsset,
		)
	}

	return nil
}
