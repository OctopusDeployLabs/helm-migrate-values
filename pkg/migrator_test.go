package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var migrateAcrossVersionsTestCases = []struct {
	name                        string
	currentVersion              int
	versionTo                   *int
	includeMigrationsToVersions []int
	expected                    map[string]interface{}
}{
	{
		name:                        "migrate across single version",
		currentVersion:              3,
		versionTo:                   ptr(4),
		includeMigrationsToVersions: []int{4},
		expected:                    version4Config,
	},
	{
		name:                        "migrate across multiple versions",
		currentVersion:              2,
		versionTo:                   ptr(4),
		includeMigrationsToVersions: []int{3, 4},
		expected:                    version4Config,
	},
	{
		name:                        "migrate with a missing migration",
		currentVersion:              1,
		versionTo:                   ptr(4),
		includeMigrationsToVersions: []int{2, 3, 4},
		expected:                    version4Config,
	},
	{
		name:                        "migrate with no end version specified",
		currentVersion:              1,
		versionTo:                   nil,
		includeMigrationsToVersions: []int{2, 3, 4},
		expected:                    version4Config,
	},
	{
		name:                        "migrate with end version less than max in available migrations",
		currentVersion:              1,
		versionTo:                   ptr(3),
		includeMigrationsToVersions: []int{2, 3, 4},
		expected:                    version3Config,
	},
	{
		name:                        "migrate with no migrations available",
		currentVersion:              1,
		versionTo:                   ptr(4),
		includeMigrationsToVersions: []int{},
		expected:                    version1Config,
	},
}

func TestMigrator_MigrateAcrossVersions(t *testing.T) {
	for _, tc := range migrateAcrossVersionsTestCases {
		t.Run(tc.name, func(t *testing.T) {
			is := assert.New(t)
			req := require.New(t)

			currentVer, ok := versionConfigs[tc.currentVersion]
			req.Truef(ok, "version %d not found", tc.currentVersion)
			req.NotNilf(currentVer, "version %d not found or has invalid YAML", tc.currentVersion)

			ms := loadMigrationsToVersions(tc.includeMigrationsToVersions)

			migrated, err := Migrate(currentVer, tc.versionTo, ms)
			req.NoError(err)

			is.EqualValues(tc.expected, migrated)
		})
	}
}

func ptr[K any](val K) *K {
	return &val
}

func loadMigrationsToVersions(versions []int) MigrationSource {
	ms := &MemoryMigrationSource{}
	for _, v := range versions {
		m, ok := migrationData[v]
		if ok {
			ms.AddMigrationData(v, m)
		}
	}
	return ms
}

var versionConfigs = map[int]map[string]interface{}{
	1: version1Config,
	2: version2Config,
	3: version3Config,
	4: version4Config,
}

var migrationData = map[int]string{
	3: version3Migration,
	4: version4Migration,
}

var version1Config, _ = yamlUnmarshal(`agent:
  targetEnvironment: "test"`)

// Has no migration (realistically not going to happen in practice)
var version2Config, _ = yamlUnmarshal(`agent:
  targetEnvironment: "test"`)

var version3Config, _ = yamlUnmarshal(`agent:
  targetEnvironments: ["test"]`)

var version4Config, _ = yamlUnmarshal(`agent:
  target:
    environments: ["test"]`)

var version3Migration = `agent:
  targetEnvironments: [{{ .agent.targetEnvironment | quote }}]`

var version4Migration = `agent:
  target:
    environments: [{{ .agent.targetEnvironments | quoteEach | join "," }}]`
