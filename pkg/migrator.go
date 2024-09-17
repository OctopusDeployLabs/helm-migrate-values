package pkg

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v2"
	"helm-migrate-values/internal"
	"os"
	"slices"
	"strings"
	"text/template"
)

func MigrateFromPath(currentConfig map[string]interface{}, vFrom int, vTo *int, migrationsDir string, log internal.Logger) (map[string]interface{}, error) {

	if len(currentConfig) == 0 {
		fmt.Println("no existing values to migrate")
		return nil, nil
	}

	if !directoryExists(migrationsDir) {
		fmt.Println("no migrations found")
		return nil, nil
	}

	log.Debug("migrating values from path: %s", migrationsDir)

	mp, err := NewFileSystemMigrationProvider(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("error creating migration provider: %w", err)
	}

	return Migrate(currentConfig, vFrom, vTo, mp, log)
}

func Migrate(currentConfig map[string]interface{}, vFrom int, vTo *int, mp MigrationProvider, log internal.Logger) (map[string]interface{}, error) {

	log.Debug("migrating values")
	versions := slices.Sorted(mp.GetVersions())

	if len(versions) == 0 {
		log.Warning("no migrations found")
		return nil, nil
	}

	migratedConfig := make(map[string]interface{})
	for key, value := range currentConfig {
		migratedConfig[key] = value
	}

	for _, version := range versions {
		if version > vFrom {
			if vTo != nil && version > *vTo {
				break
			}

			log.Debug("loading migration template for version: %d", version)
			mTemplate, err := mp.GetTemplateFor(version)
			if err != nil {
				return nil, fmt.Errorf("error retrieving migration template: %w", err)
			}

			log.Debug("applying migration template for version: %d", version)
			migratedConfig, err = apply(migratedConfig, mTemplate)
			if err != nil {
				return nil, fmt.Errorf("error applying migration: %w", err)
			}
		}
	}

	return migratedConfig, nil
}

func apply(valuesData map[string]interface{}, mTemplate string) (map[string]interface{}, error) {
	parsedTemplate, err := template.New("migration").Option("missingkey=zero").Funcs(extraFuncs()).Parse(mTemplate)
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

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
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
