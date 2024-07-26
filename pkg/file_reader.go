package pkg

import "os"

type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	ReadDir(name string) ([]DirEntry, error)
}

type DirEntry interface {
	Name() string
	IsDir() bool
}

type RealFileSystem struct{}

func (r RealFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (r RealFileSystem) ReadDir(name string) ([]DirEntry, error) {
	dirEntries, err := os.ReadDir(name)
	if err != nil {
		return nil, err
	}
	var fsDirEntries []DirEntry
	for _, dirEntry := range dirEntries {
		fsDirEntries = append(fsDirEntries, dirEntry)
	}

	return fsDirEntries, nil
}
