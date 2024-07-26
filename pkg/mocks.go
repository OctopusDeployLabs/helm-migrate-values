package pkg

import "fmt"

type MockFileSystem struct {
	fileNameData   map[string]string
	fileErr        error
	dirNameEntries map[string][]MockDirEntry
	dirErr         error
}
type MockDirEntry struct {
	name  string
	isDir bool
}

func (d MockDirEntry) Name() string {
	return d.name
}
func (d MockDirEntry) IsDir() bool {
	return d.isDir
}

func (m MockFileSystem) ReadFile(name string) ([]byte, error) {
	if m.fileErr != nil {
		return nil, m.fileErr
	}
	data, ok := m.fileNameData[name]
	if !ok {
		return nil, fmt.Errorf("file '%s' not found", name)
	}
	return []byte(data), nil
}

func (m MockFileSystem) ReadDir(name string) ([]DirEntry, error) {
	if m.dirErr != nil {
		return nil, m.dirErr
	}
	entry, ok := m.dirNameEntries[name]
	if !ok {
		return nil, fmt.Errorf("directory '%s' not found", name)
	}
	var dirEntries []DirEntry
	for _, mockEntry := range entry {
		dirEntries = append(dirEntries, mockEntry)
	}
	return dirEntries, m.dirErr
}
