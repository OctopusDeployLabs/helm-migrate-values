package pkg

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

type testCase struct {
	name      string
	values    string
	migration string
	error     error
}

func TestMigrate_HappyPath(t *testing.T) {

	originalConfig := `
agent:
  targetEnvironments: ["development", "test", "production"]
`
	var config map[string]interface{}
	err := yaml.Unmarshal([]byte(originalConfig), &config)
	assert.NoError(t, err)

	t.Run("TestMigrate", func(t *testing.T) {
		migrationTemplate := `
agent:
  target:
    environments: [{{ .agent.targetEnvironments | quoteEach | join ","}}]
`
		fs := MockFileSystem{
			fileNameData: map[string]string{
				"testdata/migrations/1.0.0-1.0.1.yaml": migrationTemplate,
			},
			dirNameEntries: map[string][]MockDirEntry{
				"testdata/migrations/": {MockDirEntry{name: "1.0.0-1.0.1.yaml", isDir: false}},
			},
		}

		output, err := Migrate(config, "1.0.0", nil, "testdata/migrations/", fs)

		expected := `
agent:
  target:
    environments: ["development", "test", "production"]
`
		assert.NoError(t, err)
		assert.Equal(t, expected, *output)
	})
}
