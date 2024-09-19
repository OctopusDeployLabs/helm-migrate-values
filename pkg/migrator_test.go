package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm-migrate-values/internal"
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
		name:                        "migrate across single version with earlier migrations present",
		currentVersion:              3,
		versionTo:                   ptr(4),
		includeMigrationsToVersions: []int{3, 4},
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
		includeMigrationsToVersions: []int{3, 4},
		expected:                    version4Config,
	},
	{
		name:                        "migrate with no end version specified",
		currentVersion:              1,
		versionTo:                   nil,
		includeMigrationsToVersions: []int{3, 4},
		expected:                    version4Config,
	},
	{
		name:                        "migrate with end version less than max in available migrations",
		currentVersion:              1,
		versionTo:                   ptr(3),
		includeMigrationsToVersions: []int{3, 4},
		expected:                    version3Config,
	},
	{
		name:                        "migrate with no migrations available",
		currentVersion:              1,
		versionTo:                   ptr(4),
		includeMigrationsToVersions: []int{},
		expected:                    nilConfig,
	},
}

func TestMigrator_MigrateAcrossVersions(t *testing.T) {
	for _, tc := range migrateAcrossVersionsTestCases {
		t.Run(tc.name, func(t *testing.T) {
			is := assert.New(t)
			req := require.New(t)

			currentConfig, ok := versionConfigs[tc.currentVersion]
			req.Truef(ok, "version %d not found", tc.currentVersion)
			req.NotNilf(currentConfig, "configuration for version %d not found, or has invalid YAML", tc.currentVersion)

			ms := loadMigrationsToVersions(tc.includeMigrationsToVersions)

			migrated, err := Migrate(currentConfig, tc.currentVersion, tc.versionTo, ms, *internal.NewLogger(false))
			req.NoError(err)

			is.EqualValues(tc.expected, migrated)
		})
	}
}

func ptr[K any](val K) *K {
	return &val
}

func loadMigrationsToVersions(versions []int) MigrationProvider {
	mp := &MemoryMigrationProvider{}
	for _, v := range versions {
		m, ok := migrationData[v]
		if ok {
			mp.AddMigrationData(v, m)
		}
	}
	return mp
}

var versionConfigs = map[int]map[string]interface{}{
	1: version1Config,
	2: version2Config,
	3: version3Config,
	4: version4Config,
}

var migrationData = map[int]map[string]interface{}{
	3: version3Migration,
	4: version4Migration,
}

var nilConfig = map[string]interface{}(nil)

var version1Config = map[string]interface{}{
	"agent": map[interface{}]interface{}{
		"targetEnvironment": "test",
	},
}

// Has no migration (realistically not going to happen in practice)
var version2Config = map[string]interface{}{
	"agent": map[interface{}]interface{}{
		"targetEnvironment": "test",
	},
}

var version3Config = map[string]interface{}{
	"agent": map[interface{}]interface{}{
		"targetEnvironments": []interface{}{"test"},
	},
}

var version4Config = map[string]interface{}{
	"agent": map[interface{}]interface{}{
		"target": map[interface{}]interface{}{
			"environments": []interface{}{"test"},
		},
	},
}

var version3Migration = map[string]interface{}{
	"agent": map[interface{}]interface{}{
		"targetEnvironments": []string{"{{ .agent.targetEnvironment }}"},
	},
}

var version4Migration = map[string]interface{}{
	"agent": map[interface{}]interface{}{
		"target": map[interface{}]interface{}{
			"environments": []string{"{{ .agent.targetEnvironments | join \",\" }}"},
		},
	},
}
