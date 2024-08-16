package pkg

import (
	"bytes"
	"cmp"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v2"
	"sort"
	"strings"
	"text/template"
)

/*
Test cases:
- value/structure is removed
- structure is changed (old structure needs to be removed)
*/

/*
	order:

- have a 'dont' migrate section
- then transform
- and copy across any other remaining.
*/
func Migrate(currentConfig map[string]interface{}, vTo *string, ms Migrations) (*string, error) {

	if len(currentConfig) == 0 {
		return nil, fmt.Errorf("no values to migrate")
	}

	migrations, err := ms.GetMigrations()
	if err != nil || len(migrations) == 0 {
		return nil, cmp.Or(err, fmt.Errorf("no migrations found"))
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ToVersion.LessThan(&migrations[j].ToVersion)
	})

	toVer, err := getVersions(vTo, migrations)
	if err != nil {
		return nil, err
	}

	var migratedConfig string

	for _, migration := range migrations {

		if migration.ToVersion.GreaterThan(toVer) {
			break
		}

		migrationData, err := ms.GetDataForMigration(&migration)

		if err != nil {
			return nil, fmt.Errorf("error reading migration: %v", err)
		}

		currentValues, err := apply(currentConfig, string(migrationData))
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(currentValues, &currentConfig)
		if err != nil {
			return nil, fmt.Errorf("error parsing migrated yaml values: %v", err) // TODO: Add migration info here
		}

		migratedConfig = string(currentValues)
	}

	return &migratedConfig, nil
}

func getVersions(vTo *string, migrations []Migration) (*version.Version, error) {
	var toVerPtr *version.Version
	var err error

	if vTo == nil {
		toVerPtr = &migrations[len(migrations)-1].ToVersion
	} else {
		toVerPtr, err = version.NewVersion(*vTo)
		if err != nil {
			return nil, fmt.Errorf("error parsing 'to' version: %v", err)
		}
	}

	return toVerPtr, nil
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
