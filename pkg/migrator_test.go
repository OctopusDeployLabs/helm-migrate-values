package pkg

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

var testCases = []struct {
	name           string
	currentConfig  string
	migration      string
	expectedOutput string
	expectError    bool
}{
	{
		name:          "Happy path",
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
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var config map[string]interface{}
			err := yaml.Unmarshal([]byte(tc.currentConfig), &config)
			assert.NoError(t, err)

			fs := MockFileSystem{
				fileNameData: map[string]string{
					"testdata/migrations/1.0.0-1.0.1.yaml": tc.migration,
				},
				dirNameEntries: map[string][]MockDirEntry{
					"testdata/migrations/": {MockDirEntry{name: "1.0.0-1.0.1.yaml", isDir: false}},
				},
			}

			output, err := Migrate(config, "1.0.0", nil, "testdata/migrations/", fs)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, *output)
			}
		})
	}
}

func stringToPtr(s string) *string {
	return &s
}

var versionsTestCases = []struct {
	name              string
	fromVersion       string
	toVersion         *string
	migrationVersions []string
	expectError       bool
}{
	{
		name:              "Upgrade single version",
		fromVersion:       "1.0.0",
		toVersion:         nil,
		migrationVersions: []string{"1.0.0-1.0.1.yaml"},
		expectError:       false,
	},
	{
		name:              "Upgrade path exists across multiple versions",
		fromVersion:       "1.0.0",
		toVersion:         nil,
		migrationVersions: []string{"1.0.0-1.0.1.yaml", "1.0.1-1.0.2.yaml"},
		expectError:       false,
	},
	{
		name:              "Upgrade path exists and from version is in middle of possible migration versions",
		fromVersion:       "1.0.1",
		toVersion:         nil,
		migrationVersions: []string{"1.0.0-1.0.1.yaml", "1.0.1-1.0.2.yaml"},
		expectError:       false,
	},
	{
		name:              "Upgrade path doesn't exist and from version is in middle of missing migration versions",
		fromVersion:       "1.0.1",
		toVersion:         nil,
		migrationVersions: []string{"1.0.0-1.0.1.yaml", "1.0.2-1.0.3.yaml"},
		expectError:       false,
	},
	{
		name:              "From version is after migration versions",
		fromVersion:       "5.0.0",
		toVersion:         nil,
		migrationVersions: []string{"1.0.0-1.0.1.yaml", "1.0.1-1.0.2.yaml"},
		expectError:       false,
	},
	{
		name:              "From version is prior to start of migrations'",
		fromVersion:       "0.0.1",
		toVersion:         nil,
		migrationVersions: []string{"1.0.0-1.0.1.yaml"},
		expectError:       false,
	},
	{
		name:              "Upgrade not required",
		fromVersion:       "1.0.1",
		toVersion:         nil,
		migrationVersions: []string{"1.0.0-1.0.1.yaml"},
		expectError:       false,
	},
	{
		name:              "Migration path is broken",
		fromVersion:       "1.0.0",
		toVersion:         nil,
		migrationVersions: []string{"1.0.0-1.0.1.yaml", "1.0.2-1.0.3.yaml"},
		expectError:       false,
	},
	{
		name:              "To version does not have a migration",
		fromVersion:       "1.0.0",
		toVersion:         stringToPtr("1.0.3"),
		migrationVersions: []string{"1.0.0-1.0.1.yaml", "1.0.1-1.0.2.yaml"},
		expectError:       false,
	},
	{
		name:              "From version does not have a migration",
		fromVersion:       "0.0.1",
		toVersion:         nil,
		migrationVersions: []string{"1.0.0-1.0.1.yaml", "1.0.1-1.0.2.yaml"},
		expectError:       false,
	},
	{
		name:              "Invalid from version",
		fromVersion:       "1.invalid.0",
		toVersion:         nil,
		migrationVersions: []string{"1.0.0-1.0.1.yaml", "1.0.2-1.0.3.yaml"},
		expectError:       true,
	},
	{
		name:              "Invalid to version",
		fromVersion:       "1.0.0",
		toVersion:         stringToPtr("1.invalid.0"),
		migrationVersions: []string{"1.0.0-1.0.1.yaml", "1.0.2-1.0.3.yaml"},
		expectError:       true,
	},
	{
		name:              "To version before from version",
		fromVersion:       "2.0.0",
		toVersion:         stringToPtr("1.0.0"),
		migrationVersions: []string{"1.0.0-1.0.1.yaml"},
		expectError:       true,
	},
}

func TestMigrator_MigrationVersionsTests(t *testing.T) {
	for _, tc := range versionsTestCases {
		t.Run(tc.name, func(t *testing.T) {

			var config map[string]interface{}
			var _ = yaml.Unmarshal([]byte("agent:"), &config)

			var fileNameData = make(map[string]string)
			for _, migrationVersion := range tc.migrationVersions {
				fileNameData["testdata/migrations/"+migrationVersion] = "target:"
			}

			var mockDirEntries []MockDirEntry
			for _, migrationVersion := range tc.migrationVersions {
				mockDirEntries = append(mockDirEntries, MockDirEntry{name: migrationVersion, isDir: false})
			}
			var dirNameEntries = make(map[string][]MockDirEntry)
			dirNameEntries["testdata/migrations/"] = mockDirEntries

			fs := MockFileSystem{
				fileNameData:   fileNameData,
				dirNameEntries: dirNameEntries,
			}

			var _, err = Migrate(config, tc.fromVersion, tc.toVersion, "testdata/migrations/", fs)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
