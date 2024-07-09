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
	from     *version.Version
	to       *version.Version
	fileName string
	fromHash string
	toHash   string
}

func newMigration(migrationFile os.DirEntry) (*migration, error) {

	//Get version string from file name, eg migration-v1.0.0-v1.0.1.yaml
	fileName := "migrations/" + migrationFile.Name()
	from, to := "", ""

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
		fileName: fileName,
		from:     fromVersion,
		to:       toVersion,
		fromHash: verHash(fromVersion),
		toHash:   verHash(toVersion),
	}, nil
}

func verHash(version *version.Version) string {
	return version.String()
}

func readValuesFile() string {
	valuesData, err := os.ReadFile("values.yaml")
	if err != nil {
		log.Fatal(err)
	}

	return string(valuesData[:])
}

func apply(values string, migration *migration, dryRunFlagPtr *bool) string {

	var valuesData = map[string]interface{}{}
	err := yaml.Unmarshal([]byte(values), &valuesData)
	if err != nil {
		log.Fatal(err)
	}

	migrationData, err := os.ReadFile(migration.fileName)
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

	versionFromPtr := flag.String("version-from", "", "Version from") // required
	versionToPtr := flag.String("version-to", "", "Version to")       // optional, assume latest if not provided
	dryRunFlagPtr := flag.Bool("dry-run", false, "Dry run")
	flag.Parse()

	//TODO: backup values file?

	// Check if the version-from is provided
	if *versionFromPtr == "" {
		log.Fatal("Version from is required")
	}

	//Load available migrations
	migrationFiles, err := os.ReadDir("migrations")
	if err != nil {
		log.Fatal(err)
	}

	if len(migrationFiles) == 0 {
		log.Fatal("No migrations found")
	}

	versionFrom, err := version.NewVersion(*versionFromPtr)
	if err != nil {
		log.Fatal(err) // Is this actually the right level?  Should we handle this and continue?
	}

	var versionTo *version.Version

	if *versionFromPtr != "" {
		versionTo, err = version.NewVersion(*versionToPtr)
		if err != nil {
			log.Fatal(err)
		}
	}

	migrations := make([]*migration, len(migrationFiles))

	// This section will need to be improved if we want to handle continuing after a broken migration.
	for i, file := range migrationFiles {
		migration, err := newMigration(file)
		if err != nil {
			log.Fatal(err)
		}
		migrations[i] = migration
	}

	verHash := func(version *version.Version) string {
		return version.String()
	}

	g := graph.New(verHash, graph.Directed(), graph.Acyclic())

	var maxVer *version.Version

	verWeight := func(from *version.Version, to *version.Version) int {
		return 1 //For now - we might want to weight the migrations in the future.
	}

	// Add all the migration versions to the graph
	for _, migration := range migrations {
		if maxVer == nil || migration.to.GreaterThan(maxVer) {
			maxVer = migration.to
		}

		errFr := g.AddVertex(migration.from)
		errTo := g.AddVertex(migration.to)
		if errFr != nil && errTo != nil {
			//Edge exists (ie dupe migration) - this shouldn't happen, they would need the same filename.
		} else {
			err := g.AddEdge(verHash(migration.from), verHash(migration.to), graph.EdgeWeight(verWeight(migration.from, migration.to)))
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if versionTo == nil {
		versionTo = maxVer
	}

	// Now see if we can go from versionFrom to versionTo
	path, err := graph.ShortestPath(g, verHash(versionFrom), verHash(versionTo))
	if err != nil {
		log.Fatal(err) // unable to find a path
	}

	// Map from fromVersion hash to migration
	migrationMap := make(map[string][]*migration)
	for _, migration := range migrations {
		//Add the migration to the map
		migrationMap[migration.fromHash] = append(migrationMap[migration.fromHash], migration) // is there a better way to do this?
	}

	// var toApply []*migration

	renderedMigration := readValuesFile()

	// Now we have the path, we can walk it and generate the migration
	for i := 0; i < len(path)-1; i++ {
		v := path[i]
		migrationsFrom := migrationMap[v]
		for _, thisMigration := range migrationsFrom {
			if thisMigration.toHash == path[i+1] {
				fmt.Println(thisMigration.from, thisMigration.to) //Just for testing.
				//	toApply = append(toApply, thisMigration)

				renderedMigration = apply(renderedMigration, thisMigration, dryRunFlagPtr)
			}
		}
	}

	//write the rendered migration to a file
	if *dryRunFlagPtr {
		log.Println(renderedMigration)
	} else {
		err = os.WriteFile("valuesOutput.yaml", []byte(renderedMigration), 0644)
	}

	/*
		versionMap := make(map[*version.Version][]*version.Version)

		for _, migration := range migrations {
			if _, ok := versionMap[migration.from]; !ok {
				versionMap[migration.from] = make([]*version.Version, 1)
			}
			versionMap[migration.from] = append(versionMap[migration.from], migration.to)
		}           */

	/*

		sort.Slice(migrations, func(i, j int) bool {
			return migrations[i].from.LessThan(migrations[j].from)
		})

		// A (better) alternative to this is to build a directed migration graph, and then walk it to find the path from versionFrom to versionTo.
		// Verify that the migrations don't cross - ie, that migrationA.To <= migrationB.From
		for i := 0; i < len(migrations)-1; i++ {
			if migrations[i].to.Equal(migrations[i+1].from) {
				log.Fatalf("Migration %s to %s crosses migration %s to %s", migrations[i].from, migrations[i].to, migrations[i+1].from, migrations[i+1].to)
			}
		}

		//Check if we can actually go from/to the versions provided
		for _, file := range migrationFiles {
			migration, err := newMigration(file)
			if err != nil {
				log.Fatal(err)
			}

			if migration.from == versionFrom {
				if *versionToPtr == "" {
					*versionToPtr = migration.to.String()
				} else {
					if migration.to.String() == *versionToPtr {
						break
					}
				}
			}
		}
	*/

	/*

		valuesData, err := os.ReadFile("values.yaml")
		if err != nil {
			log.Fatal(err)
		}

		var values = map[string]interface{}{}
		err = yaml.Unmarshal(valuesData, &values)
		if err != nil {
			log.Fatal(err)
		}

		migration, err := os.ReadFile("migration.yaml")
		if err != nil {
			log.Fatal(err)
		}

		migrationTemplate, err := template.New("migration").Funcs(funcMap()).Parse(string(migration))
		if err != nil {
			log.Fatal(err)
		}

		var renderedMigration bytes.Buffer

		err = migrationTemplate.Execute(&renderedMigration, values)
		if err != nil {
			log.Fatal(err)
		}

		//write the rendered migration to a file
		if *dryRunFlagPtr {
			log.Println(renderedMigration.String())
		} else {
			err = os.WriteFile("migrationOutput.yaml", renderedMigration.Bytes(), 0644)
		}

	*/
}

// Note: this is taken from the Helm project, so we line up with that.  It needs review/refining.
func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	// Add some extra functionality
	extra := template.FuncMap{

		// This is a placeholder for the "include" function, which is
		// late-bound to a template. By declaring it here, we preserve the
		// integrity of the linter.
		"include":  func(string, interface{}) string { return "not implemented" },
		"tpl":      func(string, interface{}) interface{} { return "not implemented" },
		"required": func(string, interface{}) (interface{}, error) { return "not implemented", nil },
		// Provide a placeholder for the "lookup" function, which requires a kubernetes
		// connection.
		"lookup": func(string, string, string, string) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		},
	}

	for k, v := range extra {
		f[k] = v
	}

	return f
}
