package pkg

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v2"
	"log"
	"sort"
	"strings"
	"text/template"
)

func Migrate(currentConfig map[string]interface{}, vFrom string, vTo *string, migrationsPath string, fileSystem FileSystem) (*string, error) {

	if len(currentConfig) == 0 {
		return nil, fmt.Errorf("no values to migrate")
	}

	fromVer, toVer, err := getVersions(vFrom, vTo)
	if err != nil {
		return nil, err
	}

	migrations, err := getMigrations(migrationsPath, fileSystem)
	if err != nil {
		return nil, err
	}
	if len(migrations) == 0 {
		return nil, fmt.Errorf("no migrations found in %s", migrationsPath)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].from.LessThan(&migrations[j].from)
	})

	if toVer == nil {
		toVer = &migrations[len(migrations)-1].to
	}

	if err = EnsureMigrationPathExists(migrations, fromVer, toVer); err != nil {
		return nil, err
	}

	var migratedConfig string

	for _, migration := range migrations {

		if migration.to.GreaterThan(toVer) {
			break
		}

		if migration.from.GreaterThanOrEqual(fromVer) {

			migrationData, err := fileSystem.ReadFile(migrationsPath + migration.fileName)

			if err != nil {
				return nil, fmt.Errorf("error reading migration file: %v", err)
			}

			currentValues, err := apply(currentConfig, string(migrationData))
			if err != nil {
				return nil, fmt.Errorf("error applying migration: %v", err)
			}

			err = yaml.Unmarshal(currentValues, &currentConfig)
			if err != nil {
				return nil, fmt.Errorf("error parsing migrated yaml values: %v", err) // TODO: Add migration info here
			}

			migratedConfig = string(currentValues)
		}
	}

	return &migratedConfig, nil
}

func getMigrations(migrationsPath string, fileSystem FileSystem) ([]Migration, error) {

	migrationFiles, err := fileSystem.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations directory: %v", err)
	}

	migrations := make([]Migration, 0, len(migrationFiles))

	for _, file := range migrationFiles {
		migration, err := NewMigration(file.Name())
		if err != nil {
			log.Fatal(err)
		}

		migrations = append(migrations, *migration)
	}

	return migrations, nil
}

func getVersions(vFrom string, vTo *string) (*version.Version, *version.Version, error) {
	fromVerPtr, err := version.NewVersion(vFrom)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing 'from' version: %v", err)
	}

	var toVerPtr *version.Version

	if vTo != nil {
		toVerPtr, err = version.NewVersion(*vTo)
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing 'to' version: %v", err)
		}
	}
	return fromVerPtr, toVerPtr, nil
}

func apply(valuesData map[string]interface{}, migration string) ([]byte, error) {
	migrationTemplate, err := template.New("migration").Funcs(extraFuncs()).Parse(migration)
	if err != nil {
		return nil, fmt.Errorf("error parsing migration template: %v", err)
	}

	var renderedMigrationBuf bytes.Buffer

	err = migrationTemplate.Execute(&renderedMigrationBuf, valuesData)
	if err != nil {
		return nil, fmt.Errorf("error executing migration template: %v", err)
	}

	return renderedMigrationBuf.Bytes(), nil
}

// Modified from https://github.com/helm/helm/blob/2feac15cc3252c97c997be2ced1ab8afe314b429/pkg/engine/funcs.go#L43
func extraFuncs() template.FuncMap {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	f["quoteEach"] = quoteEach

	return f
}

func quoteEach(s []interface{}) (string, error) {
	var quoted []string
	for _, v := range s {
		str, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("quoteEach: expected all elements to be strings, got %T", v)
		}
		quoted = append(quoted, fmt.Sprintf("%q", str))
	}
	return strings.Join(quoted, ", "), nil
}
