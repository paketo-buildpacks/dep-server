package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ChecksummerPHP
type ChecksummerPHP interface {
	GetMD5(path string) (string, error)
}

type Checksummer struct {
}

func NewPHPChecksummer() Checksummer {
	return Checksummer{}
}

func (c Checksummer) GetMD5(path string) (string, error) {
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
