package pkg

import "os"

type FileSystem interface {
	ReadFile(name string) ([]byte, error)
}

type RealFileSystem struct{}

func (r RealFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}
