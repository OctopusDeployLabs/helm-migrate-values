package pkg

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"log"
	"os"
	"regexp"
)

type Migrations interface {
	GetDataForMigration(migration *Migration) ([]byte, error)
	GetMigrations() ([]Migration, error)
}

type FileSystemMigrations struct {
	migrationsPath string
	migrations     []Migration
}

func NewFileSystemMigrations(migrationsPath string) (*FileSystemMigrations, error) {
	migrations, err := loadMigrations(migrationsPath)
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

func (migrations *FileSystemMigrations) GetMigrations() ([]Migration, error) {
	return migrations.migrations, nil
}

func loadMigrations(migrationsPath string) ([]Migration, error) {

	migrationFiles, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations directory: %v", err)
	}

	ms := make([]Migration, 0, len(migrationFiles))

	for _, file := range migrationFiles {
		migration, err := parseFilenameIntoMigration(file.Name())
		if err != nil {
			log.Fatal(err)
		}

		ms = append(ms, *migration)
	}

	return ms, nil
}

// From https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string (named groups version)
const semVerRegEx = "(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:-(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?"

func parseFilenameIntoMigration(filename string) (*Migration, error) {

	// Get version string from file name, eg '1.0.0-1.0.1.yaml'
	pattern := fmt.Sprintf(`(?P<fromVersion>%s)\-(?P<toVersion>%s)\.yaml`, semVerRegEx, semVerRegEx)
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(filename)
	names := re.SubexpNames()
	var from, to string
	for i, match := range matches {
		if names[i] == "fromVersion" {
			from = match
		} else if names[i] == "toVersion" {
			to = match
		}
	}

	fromVersion, err := version.NewVersion(from)
	if err != nil {
		return nil, fmt.Errorf("error parsing 'from' version '%s'': %v", from, err)
	}

	toVersion, err := version.NewVersion(to)
	if err != nil {
		return nil, fmt.Errorf("error parsing 'to' version '%s': %v", to, err)
	}

	if fromVersion.GreaterThanOrEqual(toVersion) {
		return nil, fmt.Errorf("migration 'from' versions must be less than their 'to' version")
	}

	return &Migration{
		From: *fromVersion,
		To:   *toVersion,
	}, nil
}
