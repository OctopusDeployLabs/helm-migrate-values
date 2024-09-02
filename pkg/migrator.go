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

func MigrateFromPath(currentConfig map[string]interface{}, vTo *int, migrationsDir string) (map[string]interface{}, error) {

	if len(currentConfig) == 0 {
		return currentConfig, nil
	}

	ms, err := NewFileSystemMigrationSource(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("error creating migration source: %w", err)
	}

	return Migrate(currentConfig, vTo, ms)
}

func Migrate(currentConfig map[string]interface{}, vTo *int, ms MigrationSource) (map[string]interface{}, error) {

	versions := slices.Sorted(ms.GetVersions())

	if len(versions) == 0 {
		return currentConfig, nil
	}

	migratedConfig := currentConfig

	for _, version := range versions {
		if vTo != nil && version > *vTo {
			break
		}

		mTemplate, err := ms.GetTemplateFor(version)
		if err != nil {
			return nil, fmt.Errorf("error retrieving migration template: %w", err)
		}

		migratedConfig, err = apply(migratedConfig, mTemplate)
		if err != nil {
			return nil, fmt.Errorf("error applying migration: %w", err)
		}
	}

	return migratedConfig, nil
}

func apply(valuesData map[string]interface{}, mTemplate string) (map[string]interface{}, error) {
	parsedTemplate, err := template.New("migration").Funcs(extraFuncs()).Parse(mTemplate)
	if err != nil {
		return nil, fmt.Errorf("error parsing migration template: %w", err)
	}

	var renderedMigrationBuf bytes.Buffer
	err = parsedTemplate.Execute(&renderedMigrationBuf, valuesData)
	if err != nil {
		return nil, fmt.Errorf("error executing migration template: %w", err)
	}

	var migratedConfig map[string]interface{}
	err = yaml.Unmarshal(renderedMigrationBuf.Bytes(), &migratedConfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing migrated yaml values %w", err)
	}

	return migratedConfig, nil
}

// Modified from https://github.com/helm/helm/blob/2feac15cc3252c97c997be2ced1ab8afe314b429/pkg/engine/funcs.go#L43
func extraFuncs() template.FuncMap {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	f["quoteEach"] = quoteEach
	f["toYaml"] = toYaml

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

func toYaml(t any) string {
	result, _ := yaml.Marshal(t)
	return string(result)
}
