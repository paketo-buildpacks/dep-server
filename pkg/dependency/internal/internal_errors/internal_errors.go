package internal_errors

import "fmt"

type AssetNotFound struct {
	AssetName string
}

func (a AssetNotFound) Error() string {
	return fmt.Sprintf("could not find asset %s", a.AssetName)
}
