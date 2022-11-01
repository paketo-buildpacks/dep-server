package internal

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"
)

type Checksummer struct{}

func NewChecksummer() Checksummer {
	return Checksummer{}
}

func (c Checksummer) VerifyASC(asc, path string, pgpKeys ...string) error {
	if len(pgpKeys) == 0 {
		return errors.New("no pgp keys provided")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	for _, pgpKey := range pgpKeys {
		keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(pgpKey))
		if err != nil {			
			log.Printf("could not read armored key ring: %s", err.Error())
			continue
		}

		_, err = openpgp.CheckArmoredDetachedSignature(keyring, file, strings.NewReader(asc))
		if err != nil {
			log.Printf("failed to check signature: %s", err.Error())
			continue
		}
		log.Printf("found valid pgp key")
		return nil
	}

	return errors.New("no valid pgp keys provided")
}

func (c Checksummer) VerifyMD5(path, expectedMD5 string) error {
	actualMD5, err := c.getMD5(path)
	if err != nil {
		return fmt.Errorf("failed to get actual MD5: %w", err)
	}

	if actualMD5 != expectedMD5 {
		return fmt.Errorf("expected MD5 '%s' but got '%s'", expectedMD5, actualMD5)
	}

	return nil
}

func (c Checksummer) VerifySHA1(path, expectedSHA string) error {
	actualSHA, err := c.getSHA1(path)
	if err != nil {
		return fmt.Errorf("failed to get actual SHA256: %w", err)
	}

	if actualSHA != expectedSHA {
		return fmt.Errorf("expected SHA256 '%s' but got '%s'", expectedSHA, actualSHA)
	}

	return nil
}

func (c Checksummer) VerifySHA256(path, expectedSHA string) error {
	actualSHA, err := c.GetSHA256(path)
	if err != nil {
		return fmt.Errorf("failed to get actual SHA256: %w", err)
	}

	if actualSHA != expectedSHA {
		return fmt.Errorf("expected SHA256 '%s' but got '%s'", expectedSHA, actualSHA)
	}

	return nil
}

func (c Checksummer) VerifySHA512(path, expectedSHA string) error {
	actualSHA, err := c.GetSHA512(path)
	if err != nil {
		return fmt.Errorf("failed to get actual SHA256: %w", err)
	}

	if actualSHA != expectedSHA {
		return fmt.Errorf("expected SHA256 '%s' but got '%s'", expectedSHA, actualSHA)
	}

	return nil
}

func (c Checksummer) GetSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "nil", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "nil", fmt.Errorf("failed to calculate SHA256: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (c Checksummer) GetSHA512(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "nil", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha512.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "nil", fmt.Errorf("failed to calculate SHA256: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (c Checksummer) SplitPGPKeys(block string) []string {
	var keys []string
	var currentKey string
	inKey := false

	for _, line := range strings.Split(string(block), "\n") {
		if line == "-----BEGIN PGP PUBLIC KEY BLOCK-----" {
			currentKey = line + "\n"
			inKey = true
		} else if line == "-----END PGP PUBLIC KEY BLOCK-----" {
			currentKey = currentKey + line
			keys = append(keys, currentKey)
			inKey = false
		} else if inKey {
			currentKey = currentKey + line + "\n"
		}
	}

	return keys
}

func (c Checksummer) getMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "nil", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "nil", fmt.Errorf("failed to calculate MD5: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (c Checksummer) getSHA1(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "nil", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "nil", fmt.Errorf("failed to calculate SHA1: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
