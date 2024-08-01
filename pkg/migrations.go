package pkg

import (
	"fmt"
	"log"
	"os"
	"sort"
)

type Migrations interface {
	GetDataForMigration(migration *Migration) ([]byte, error)
	GetSortedMigrations() ([]Migration, error)
}

type FileSystemMigrations struct {
	migrationsPath string
	migrations     []Migration
}

func NewFileSystemMigrations(migrationsPath string) (*FileSystemMigrations, error) {

	migrations, err := getSortedMigrations(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations: %v", err)
	}

	return &FileSystemMigrations{
		migrationsPath: migrationsPath,
		migrations:     migrations,
	}, nil
}

func (migrations *FileSystemMigrations) GetDataForMigration(m *Migration) ([]byte, error) {
	fileName := fmt.Sprintf("%s-%s.yaml", m.From.String(), m.To.String())
	data, err := os.ReadFile(migrations.migrationsPath + fileName)
	if err != nil {
		return nil, fmt.Errorf("error reading migration: %v", err)
	}
	return data, nil
}

func getSortedMigrations(migrationsPath string) ([]Migration, error) {

	migrationFiles, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations directory: %v", err)
	}

	ms := make([]Migration, 0, len(migrationFiles))

	for _, file := range migrationFiles {
		migration, err := NewMigration(file.Name())
		if err != nil {
			log.Fatal(err)
		}

		ms = append(ms, *migration)
	}

	sort.Slice(ms, func(i, j int) bool {
		return ms[i].From.LessThan(&ms[j].From)
	})

	return ms, nil
}
