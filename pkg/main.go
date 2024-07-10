package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/dominikbraun/graph"
	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"regexp"
	"text/template"
)

type migration struct {
	from     version.Version
	to       version.Version
	filePath string
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
		return nil, err
	}

	toVersion, err := version.NewVersion(to)
	if err != nil {
		return nil, err
	}

	return &migration{
		filePath: filePath,
		from:     *fromVersion,
		to:       *toVersion,
	}, nil
}

func readValuesFile() string {
	valuesData, err := os.ReadFile("values.yaml")
	if err != nil {
		log.Fatal(err)
	}

	return string(valuesData[:])
}

func apply(values string, migration migration) string {

	var valuesData = map[string]interface{}{}
	err := yaml.Unmarshal([]byte(values), &valuesData)
	if err != nil {
		log.Fatal(err)
	}

	migrationData, err := os.ReadFile(migration.filePath)
	if err != nil {
		log.Fatal(err)
	}

	migrationTemplate, err := template.New("migration").Funcs(funcMap()).Parse(string(migrationData))
	if err != nil {
		log.Fatal(err)
	}

	var renderedMigration bytes.Buffer

	err = migrationTemplate.Execute(&renderedMigration, valuesData)
	if err != nil {
		log.Fatal(err)
	}

	return renderedMigration.String()
}

func main() {

	flagVersionFrom := flag.String("version-from", "", "Version from") // required
	flagVersionTo := flag.String("version-to", "", "Version to")       // optional, assume latest if not provided
	flagDryRun := flag.Bool("dry-run", false, "Dry run")               // optional
	flag.Parse()

	//TODO: backup values file?

	// Check if the version-from is provided
	if *flagVersionFrom == "" {
		log.Fatal("Version from is required")
	}

	//Load available migration file names
	migrationFiles, err := os.ReadDir("migrations")
	if err != nil {
		log.Fatal(err)
	}

	if len(migrationFiles) == 0 {
		log.Fatal("No migrations found")
	}

	migrations := make([]migration, len(migrationFiles))

	// This section will need to be improved if we want to handle continuing after a broken migration.
	for i, file := range migrationFiles {
		migration, err := newMigration(file.Name())
		if err != nil {
			log.Fatal(err)
		}

		migrations[i] = *migration
	}

	// Get the to/from versions
	versionFromPtr, err := version.NewVersion(*flagVersionFrom)
	if err != nil {
		log.Fatal(err) // Is this actually the right level?  Should we handle this and continue?
	}

	versionFrom := (*versionFromPtr).String()

	versionToPtr, err := getVersionTo(flagVersionTo, migrations)
	if err != nil {
		log.Fatal(err)
	}
	versionTo := (*versionToPtr).String()

	g := getVersionGraph(migrations)

	// Now see if we can go from versionFrom to versionTo
	verPath, err := graph.ShortestPath(g, versionFrom, versionTo)
	if err != nil {
		log.Fatal(err) // unable to find a path
	}

	// Map from fromVersion hash to migration
	migrationMap := make(map[string][]migration)
	for _, migration := range migrations {
		migrationMap[migration.from.String()] = append(migrationMap[migration.from.String()], migration) // is there a better way to do this?
	}

	//var appliedMigrations []*migration

	originalValues := readValuesFile()
	thisMigrationValues := originalValues

	// Now we have the path, we can walk it and generate the migration
	for i := 0; i < len(verPath)-1; i++ {
		ver := verPath[i]
		migrationsFrom := migrationMap[ver]
		for _, thisMigration := range migrationsFrom {
			if thisMigration.to.String() == verPath[i+1] {
				log.Printf("%s => %s\n", thisMigration.from.String(), thisMigration.to.String()) //Just for testing.
				//	appliedMigrations = append(appliedMigrations, thisMigration)

				thisMigrationValues = apply(thisMigrationValues, thisMigration)
			}
		}
	}

	//write the rendered migration to a file
	if *flagDryRun {
		log.Printf("\n%s\n", thisMigrationValues)
	} else {
		err = os.WriteFile("valuesOutput.yaml", []byte(thisMigrationValues), 0644)
	}
}

func getVersionTo(flagVersionTo *string, migrations []migration) (*version.Version, error) {
	if len(migrations) == 0 {
		return nil, fmt.Errorf("no migrations provided")
	}

	if *flagVersionTo == "" {
		var maxVer *version.Version
		for _, migration := range migrations {
			if maxVer == nil || migration.to.GreaterThan(maxVer) {
				maxVer = &migration.to
			}
		}
		return maxVer, nil
	} else {
		return version.NewVersion(*flagVersionTo)
	}
}

func getVersionGraph(migrations []migration) graph.Graph[string, string] {

	g := graph.New(graph.StringHash, graph.Directed(), graph.Acyclic())

	verWeight := func(from version.Version, to version.Version) int {
		return 1 //For now - we might want to weight the versions in the future.
	}

	// Add all the migration versions to the graph
	for _, migration := range migrations {
		errFr := g.AddVertex(migration.from.String())
		errTo := g.AddVertex(migration.to.String())
		if errFr != nil && errTo != nil {
			//Edge exists (ie dupe migration) - this shouldn't happen, they would need the same filename.
		} else {
			// Each edge is a migration
			err := g.AddEdge(migration.from.String(), migration.to.String(), graph.EdgeWeight(verWeight(migration.from, migration.to)))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return g
}

// Note: this is modified from the Helm project, so we line up with that.
func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	return f
}
