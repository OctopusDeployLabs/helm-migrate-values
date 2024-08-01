package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var filenameParserTests = []struct {
	name       string
	in         string
	outFromVer string
	outToVer   string
	hasErr     bool
}{
	{"Valid migration with different from and to versions", "1.0.0-1.0.1.yaml", "1.0.0", "1.0.1", false},
	{"From version with pre-release label", "1.0.0-alpha-1.0.1.yaml", "1.0.0-alpha", "1.0.1", false},
	{"To version with pre-release label", "1.0.0-1.0.1-beta.yaml", "1.0.0", "1.0.1-beta", false},
	{"Both versions with pre-release labels", "1.0.0-alpha-1.0.1-beta.yaml", "1.0.0-alpha", "1.0.1-beta", false},
	{"Pre-release labels with additional identifiers", "1.0.0-alpha.1-1.0.1-beta.1.yaml", "1.0.0-alpha.1", "1.0.1-beta.1", false},
	{"Versions with build metadata", "1.0.0+20130313144700-1.0.1+20130313144700.yaml", "1.0.0+20130313144700", "1.0.1+20130313144700", false},
	{"From version is equal to to version", "1.0.0-1.0.0.yaml", "", "", true},
	{"from version is greater than to version", "1.0.1-1.0.0.yaml", "", "", true},
	{"to version is not a valid semver", "1.0.0-1.0.yaml", "", "", true},
	{" from version is not a valid semver", "1.0-1.0.1.yaml", "", "", true},
	{"File does not have an extension", "1.0.0-1.0.1", "", "", true},
	{"File does not have .yaml extension", "1.0.0-1.0.1.txt", "", "", true},
	{"File does not have .yaml extension", "1.0.0-1.0.1.txt.yaml", "", "", true},
	{"Versions should not start with v", "v1.0.0-v1.0.1.yaml", "", "", true},
}

func TestParseFilenameIntoMigration(t *testing.T) {
	for _, tt := range filenameParserTests {
		t.Run(tt.name, func(t *testing.T) {
			migration, err := parseFilenameIntoMigration(tt.in)

			if tt.hasErr {
				assert.Error(t, err)
				assert.Nil(t, migration)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, migration)
				if migration != nil {
					assert.Equal(t, tt.outFromVer, migration.From.String())
					assert.Equal(t, tt.outToVer, migration.To.String())
				}
			}
		})
	}
}
