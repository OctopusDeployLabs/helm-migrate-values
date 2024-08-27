package pkg

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v2"
	"slices"
	"strings"
	"text/template"
)

func MigrateFromFiles(currentConfig map[string]interface{}, vTo *int, migrationsDir string) (*string, error) {

	// I don't think this is correct - we might need to set new values? e.g. like we did with targets in v2
	/*
		if len(currentConfig) == 0 {
			return nil, fmt.Errorf("no values to migrate")
		} */

	ms, err := NewFileSystemMigrationSource(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("error creating migration source: %v", err)
	}

	return Migrate(currentConfig, vTo, ms)
}

func Migrate(currentConfig map[string]interface{}, vTo *int, ms MigrationSource) (*string, error) {

	versions := slices.Sorted(ms.GetVersions())

	if len(versions) == 0 {
		return nil, fmt.Errorf("no migrations found")
	}

	var migratedConfig string

	for _, version := range versions {
		if vTo != nil && version > *vTo {
			break
		}

		mTemplate, err := ms.GetTemplateFor(version)
		if err != nil {
			return nil, fmt.Errorf("error reading migration template: %v", err)
		}

		currentValues, err := apply(currentConfig, mTemplate)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(currentValues, &currentConfig)
		if err != nil {
			return nil, fmt.Errorf("error parsing migrated yaml values in v%d: %v", version, err)
		}
		migratedConfig = string(currentValues)
	}

	return &migratedConfig, nil
}

func apply(valuesData map[string]interface{}, mTemplate string) ([]byte, error) {
	parsedTemplate, err := template.New("migration").Funcs(extraFuncs()).Parse(mTemplate)
	if err != nil {
		return nil, fmt.Errorf("error parsing migration template: %v", err)
	}

	var renderedMigrationBuf bytes.Buffer

	err = parsedTemplate.Execute(&renderedMigrationBuf, valuesData)
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

/*

	migrations, err := ms.GetMigrations()
	if err != nil || len(migrations) == 0 {
		return nil, cmp.Or(err, fmt.Errorf("no migrations found"))
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ToVersion.LessThan(&migrations[j].ToVersion)
	})                */
