package errors

import "fmt"

type NoSourceCodeError struct {
	Version string
}

func (n NoSourceCodeError) Error() string {
	return fmt.Sprintf("could not find source code for dependency version %s", n.Version)
}
