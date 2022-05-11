package internal

import "os"

type FileSystem struct{}

func NewFileSystem() FileSystem {
	return FileSystem{}
}

func (f FileSystem) WriteFile(filename, contents string) error {
	return os.WriteFile(filename, []byte(contents), 0644)
}
