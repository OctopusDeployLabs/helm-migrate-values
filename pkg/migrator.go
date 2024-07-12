package pkg

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"sort"
	"text/template"
)

type FileReader interface {
	ReadFile(name string) ([]byte, error)
	ReadDir()
}

func Migrate(values string, vFrom string, vTo *string, migrationsPath string, fileReader FileReader) (*string, error) {

	fromVer, toVer, err := getVersions(vFrom, vTo)
	if err != nil {
		return nil, err
	}

	migrations, err := getMigrations(migrationsPath)
	if err != nil {
		return nil, err
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].from.LessThan(&migrations[j].from)
	})

	if toVer == nil {
		toVer = &migrations[len(migrations)-1].to
	}

	if err = EnsurePathExists(migrations, fromVer, toVer); err != nil {
		return nil, err
	}

	var currentValuesData = []byte(values)

	for _, migration := range migrations {

		if migration.to.GreaterThan(toVer) {
			break
		}

		if migration.from.GreaterThanOrEqual(fromVer) {
			file := migrationsPath + migration.fileName
			migrationData, err := fileReader.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("error reading migration file: %v", err)
			}

			currentValuesData, err = apply(currentValuesData, string(migrationData))
			if err != nil {
				return nil, err
			}
		}
	}

	values = string(currentValuesData)

	return &values, nil

}

func getMigrations(migrationsPath string) ([]Migration, error) {

	migrationFiles, err := os.ReadDir(migrationsPath)
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

func apply(values []byte, migration string) ([]byte, error) {

	var valuesData = map[string]interface{}{}
	err := yaml.Unmarshal(values, &valuesData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling yaml values: %v", err)
	}

	migrationTemplate, err := template.New("migration").Funcs(extraFuncs()).Parse(migration)
	if err != nil {
		return nil, fmt.Errorf("error parsing migration template: %v", err)
	}

	var renderedMigration bytes.Buffer

	err = migrationTemplate.Execute(&renderedMigration, valuesData)
	if err != nil {
		return nil, fmt.Errorf("error executing migration template: %v", err)
	}

	return renderedMigration.Bytes(), nil
}

// Modified from https://github.com/helm/helm/blob/2feac15cc3252c97c997be2ced1ab8afe314b429/pkg/engine/funcs.go#L43
func extraFuncs() template.FuncMap {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	return f
}
