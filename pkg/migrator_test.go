package pkg

import (
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"log"
	"testing"
)

var migratorTestCases = []struct {
	name           string
	currentConfig  string
	migration      string
	expectedOutput string
	expectError    bool
}{
	{
		name:          "Valid migration",
		currentConfig: `targetEnvironments: ["development", "test", "production"]`,
		migration: `
target:
  environments: [{{ .targetEnvironments | quoteEach | join ","}}]`,
		expectedOutput: `
target:
  environments: ["development", "test", "production"]`,
		expectError: false,
	},
	{
		name:          "No values to migrate",
		currentConfig: ``,
		migration: `
target:
  environments: [{{ .targetEnvironments | quoteEach | join ","}}]`,
		expectedOutput: ``,
		expectError:    true,
	},
	{
		name:          "Can handle conditionals",
		currentConfig: `targetEnvironments:`,
		migration: `
{{if .targetEnvironments}}
target:
  environments: [{{ .targetEnvironments | quoteEach | join ","}}]
{{end}}`,
		expectedOutput: `
`,
		expectError: false,
	},
	{
		name:           "Invalid migration",
		currentConfig:  `agent:`,
		migration:      `{{ invalid_function }}`,
		expectedOutput: ``,
		expectError:    true,
	},
}

func TestMigrator_ValuesTests(t *testing.T) {
	for _, tc := range migratorTestCases {
		t.Run(tc.name, func(t *testing.T) {

			var config map[string]interface{}
			err := yaml.Unmarshal([]byte(tc.currentConfig), &config)
			assert.NoError(t, err)

			ms := NewMockMigrations()
			m := newMigration("1.0.0", "1.0.1")
			ms.AddMigrationData(m, tc.migration)

			output, err := Migrate(config, m.From.String(), nil, ms)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if err == nil {
					assert.Equal(t, tc.expectedOutput, *output)
				}
			}
		})
	}
}

func stringToPtr(s string) *string {
	return &s
}

var versionsTestCases = []struct {
	name        string
	fromVersion string
	toVersion   *string
	migrations  []Migration
	expectError bool
}{
	{
		name:        "Upgrade single version",
		fromVersion: "1.0.0",
		toVersion:   nil,
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
		},
		expectError: false,
	},
	{
		name:        "Upgrade path exists across multiple versions",
		fromVersion: "1.0.0",
		toVersion:   nil,
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
			newMigration("1.0.1", "1.0.2"),
		},
		expectError: false,
	},
	{
		name:        "Upgrade path exists and from version is in middle of possible migration versions",
		fromVersion: "1.0.1",
		toVersion:   nil,
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
			newMigration("1.0.1", "1.0.2"),
		},
		expectError: false,
	},
	{
		name:        "Upgrade path doesn't exist and from version is in middle of missing migration versions",
		fromVersion: "1.0.1",
		toVersion:   nil,
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
			newMigration("1.0.2", "1.0.3"),
		},
		expectError: false,
	},
	{
		name:        "From version is after migration versions",
		fromVersion: "5.0.0",
		toVersion:   nil,
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
			newMigration("1.0.1", "1.0.2"),
		},
		expectError: false,
	},
	{
		name:        "From version is prior to start of migrations'",
		fromVersion: "0.0.1",
		toVersion:   nil,
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
		},
		expectError: false,
	},
	{
		name:        "Upgrade not required",
		fromVersion: "1.0.1",
		toVersion:   nil,
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
		},
		expectError: false,
	},
	{
		name:        "Migration path is broken",
		fromVersion: "1.0.0",
		toVersion:   nil,
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
			newMigration("1.0.2", "1.0.3"),
		},
		expectError: false,
	},
	{
		name:        "To version does not have a migration",
		fromVersion: "1.0.0",
		toVersion:   stringToPtr("1.0.3"),
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
			newMigration("1.0.1", "1.0.2"),
		},
		expectError: false,
	},
	{
		name:        "From version does not have a migration",
		fromVersion: "0.0.1",
		toVersion:   nil,
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
			newMigration("1.0.1", "1.0.2"),
		},
		expectError: false,
	},
	{
		name:        "Invalid from version",
		fromVersion: "1.invalid.0",
		toVersion:   nil,
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
			newMigration("1.0.2", "1.0.3"),
		},
		expectError: true,
	},
	{
		name:        "Invalid to version",
		fromVersion: "1.0.0",
		toVersion:   stringToPtr("1.invalid.0"),
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
			newMigration("1.0.2", "1.0.3"),
		},
		expectError: true,
	},
	{
		name:        "To version before from version",
		fromVersion: "2.0.0",
		toVersion:   stringToPtr("1.0.0"),
		migrations: []Migration{
			newMigration("1.0.0", "1.0.1"),
		},
		expectError: true,
	},
}

func TestMigrator_MigrationVersionsTests(t *testing.T) {
	for _, tc := range versionsTestCases {
		t.Run(tc.name, func(t *testing.T) {

			var config map[string]interface{}
			var _ = yaml.Unmarshal([]byte("agent:"), &config)

			someIrrelevantData := "target:"

			ms := NewMockMigrations()

			for _, migration := range tc.migrations {
				ms.AddMigrationData(migration, someIrrelevantData)
			}

			var _, err = Migrate(config, tc.fromVersion, tc.toVersion, ms)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newMigration(vFrom string, vTo string) Migration {
	from, err := version.NewVersion(vFrom)
	if err != nil {
		log.Fatal(err)
	}
	to, err := version.NewVersion(vTo)
	if err != nil {
		log.Fatal(err)
	}
	return Migration{
		From: *from,
		To:   *to,
	}
}
