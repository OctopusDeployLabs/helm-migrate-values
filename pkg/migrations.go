package pkg

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"iter"
	"maps"
	"os"
	"strconv"
	"strings"
)

const filePrefix = "to-v"
const fileExtension = ".yaml"

type FileSystemMigrationSource struct {
	BaseDir        string
	VersionPathMap map[int]string
}

func NewFileSystemMigrationSource(dir string) (*FileSystemMigrationSource, error) {

	md, err := loadMigrationMetadata(dir)
	if err != nil {
		return nil, err
	}

	return &FileSystemMigrationSource{
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

	for _, file := range migrationFiles {
		if file.IsDir() || !strings.HasPrefix(file.Name(), filePrefix) || !strings.HasSuffix(file.Name(), fileExtension) {
			continue
		}
		ver, err := strconv.Atoi(strings.TrimRight(strings.TrimLeft(file.Name(), filePrefix), fileExtension))
		if err != nil {
			return nil, fmt.Errorf("error parsing version from '%s': %w", file.Name(), err)
		}

		versionPathMap[ver] = file.Name()
	}

	return versionPathMap, nil
}

type MigrationSource interface {
	GetTemplateFor(v int) (string, error)
	GetVersions() iter.Seq[int]
}

func (f *FileSystemMigrationSource) GetTemplateFor(v int) (string, error) {
	filePath := f.VersionPathMap[v]
	if filePath == "" {
		return "", fmt.Errorf("no migration found for version %d", v)
	}

	mTemplate, err := os.ReadFile(f.BaseDir + filePath)
	if err != nil {
		return "", fmt.Errorf("error reading migration template: %w", err)
	}

	return string(mTemplate), nil
}

func (f *FileSystemMigrationSource) GetVersions() iter.Seq[int] {
	return maps.Keys(f.VersionPathMap)
}

type MemoryMigrationSource struct {
	VersionDataMap map[int]string
}

func (m *MemoryMigrationSource) GetTemplateFor(v int) (string, error) {
	data, ok := m.VersionDataMap[v]
	if !ok {
		return "", fmt.Errorf("no migration found for version %d", v)
	}

	return data, nil
}

func (m *MemoryMigrationSource) GetVersions() iter.Seq[int] {
	return maps.Keys(m.VersionDataMap)
}

func (m *MemoryMigrationSource) AddMigrationData(v int, data map[string]interface{}) {
	if m.VersionDataMap == nil {
		m.VersionDataMap = make(map[int]string)
	}

	dataM, _ := yaml.Marshal(data)

	m.VersionDataMap[v] = string(dataM)
}
