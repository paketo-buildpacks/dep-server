package internal

import "io/ioutil"

type FileSystem struct{}

func NewFileSystem() FileSystem {
	return FileSystem{}
}

func (f FileSystem) WriteFile(filename, contents string) error {
	return ioutil.WriteFile(filename, []byte(contents), 0644)
}
