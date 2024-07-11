package pkg

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"regexp"
	"sort"
	"text/template"
)

type migration struct {
	from     version.Version
	to       version.Version
	filePath string
}

func Migrate(values string, vFrom string, vTo *string, migrationsPath string) (*string, error) {

	fromVerPtr, toVerPtr, err := getVersions(vFrom, vTo)
	if err != nil {
		return nil, err
	}

	migrationFiles, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations directory: %v", err)
	}

	migrations, err := getMigrations(migrationFiles, fromVerPtr, toVerPtr)
	if err != nil {
		return nil, err
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].from.LessThan(&migrations[j].from)
	})

	if err = ensureContiguous(migrations); err != nil {
		return nil, err
	}

	if toVerPtr == nil {
		toVerPtr = &migrations[len(migrations)-1].to
	}

	var valuesData = []byte(values)

	for _, migration := range migrations {

		if migration.to.GreaterThan(toVerPtr) {
			break
		}

		if migration.from.GreaterThanOrEqual(fromVerPtr) {
			file := migrationsPath + migration.filePath
			migrationData, err := os.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("error reading migration file: %v", err)
			}

			valuesData, err = apply(valuesData, string(migrationData))
			if err != nil {
				return nil, err
			}
		}
	}
	values = string(valuesData)

	return &values, nil

}

func getMigrations(migrationFiles []os.DirEntry, fromVerPtr *version.Version, toVerPtr *version.Version) ([]migration, error) {
	migrations := make([]migration, 0, len(migrationFiles))

	fromVerHasMigration := false
	toVerHasMigration := false

	for _, file := range migrationFiles {
		migration, err := newMigration(file.Name())
		if err != nil {
			log.Fatal(err)
		}

		fromVerHasMigration = fromVerHasMigration || migration.from.Equal(fromVerPtr)
		toVerHasMigration = toVerHasMigration || (toVerPtr != nil && migration.to.Equal(toVerPtr))

		migrations = append(migrations, *migration)
	}

	if !fromVerHasMigration || !toVerHasMigration && toVerPtr != nil {
		return nil, fmt.Errorf("no migration found between provided versions")
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

func newMigration(fileName string) (*migration, error) {
	from, to, filePath := "", "", "migrations/"+fileName

	// Get version string from file name, eg migration-v1.0.0-v1.0.1.yaml
	pattern := `migration-v([0-9]+\.[0-9]+\.[0-9]+(?:-[\w\.-]+)?)\-v([0-9]+\.[0-9]+\.[0-9]+(?:-[\w\.-]+)?)\.yaml` //obvs not my work, need to check this/work out a better way.
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(fileName)
	from = matches[1]
	to = matches[2]

	fromVersion, err := version.NewVersion(from)
	if err != nil {
		return nil, fmt.Errorf("error parsing 'from' version: %v", err)
	}

	toVersion, err := version.NewVersion(to)
	if err != nil {
		return nil, fmt.Errorf("error parsing 'to' version: %v", err)
	}

	return &migration{
		filePath: filePath,
		from:     *fromVersion,
		to:       *toVersion,
	}, nil
}

func ensureContiguous(migrations []migration) error {

	for i, current := range migrations[:len(migrations)-1] {
		next := migrations[i+1]

		if !current.to.Equal(&next.from) {
			return fmt.Errorf("migrations path is broken")
		}
	}

	return nil
}

func apply(values []byte, migration string) ([]byte, error) {

	var valuesData = map[string]interface{}{}
	err := yaml.Unmarshal(values, &valuesData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling yaml values: %v", err)
	}

	migrationTemplate, err := template.New("migration").Funcs(funcMap()).Parse(migration)
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
func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	return f
}
