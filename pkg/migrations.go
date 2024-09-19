package pkg

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"iter"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type FileSystemMigrationProvider struct {
	BaseDir        string
	VersionPathMap map[int]string
}

func NewFileSystemMigrationProvider(dir string) (*FileSystemMigrationProvider, error) {
	md, err := loadMigrationMetadata(dir)
	if err != nil {
		return nil, err
	}

	return &FileSystemMigrationProvider{
		BaseDir:        dir,
		VersionPathMap: md,
	}, nil
}

type FileSystemMigrationMeta struct {
	ToVersion int
	Path      string
}

type Migration struct {
	ToVersion int
	Data      map[string]interface{}
}

func loadMigrationMetadata(dir string) (map[int]string, error) {
	migrationFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations directory: %w", err)
	}

	versionPathMap := make(map[int]string)

	filePattern := `^to-v(\d+)\.(yml|yaml)$`
	re := regexp.MustCompile(filePattern)

	for _, file := range migrationFiles {
		matches := re.FindStringSubmatch(file.Name())
		if file.IsDir() || len(matches) < 2 {
			continue
		}

		ver, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("error parsing version from '%s': %w", file.Name(), err)
		}

		versionPathMap[ver] = file.Name()
	}

	return versionPathMap, nil
}

type MigrationProvider interface {
	GetTemplateFor(v int) (string, error)
	GetVersions() iter.Seq[int]
}

func (f *FileSystemMigrationProvider) GetTemplateFor(v int) (string, error) {
	fPath := f.VersionPathMap[v]
	if fPath == "" {
		return "", fmt.Errorf("no migration found for version %d", v)
	}

	fullPath := filepath.Join(f.BaseDir, fPath)
	mTemplate, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("error reading migration template: %w", err)
	}

	return string(mTemplate), nil
}

func (f *FileSystemMigrationProvider) GetVersions() iter.Seq[int] {
	return maps.Keys(f.VersionPathMap)
}

type MemoryMigrationProvider struct {
	VersionDataMap map[int]string
}

func (m *MemoryMigrationProvider) GetTemplateFor(v int) (string, error) {
	data, ok := m.VersionDataMap[v]
	if !ok {
		return "", fmt.Errorf("no migration found for version %d", v)
	}

	return data, nil
}

func (m *MemoryMigrationProvider) GetVersions() iter.Seq[int] {
	return maps.Keys(m.VersionDataMap)
}

func (m *MemoryMigrationProvider) AddMigrationData(v int, data map[string]interface{}) {
	if m.VersionDataMap == nil {
		m.VersionDataMap = make(map[int]string)
	}

	dataM, _ := yaml.Marshal(data)

	m.VersionDataMap[v] = string(dataM)
}
