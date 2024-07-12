package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var newMigrationTests = []struct {
	in         string
	outFromVer string
	outToVer   string
	hasErr     bool
}{
	// Existing test cases...
	{"migration-v1.0.0-v1.0.1.yaml", "1.0.0", "1.0.1", false},
	{"migration-v1.0.0-alpha-v1.0.1.yaml", "1.0.0-alpha", "1.0.1", false},                                                 // 'from' version with pre-release label
	{"migration-v1.0.0-v1.0.1-beta.yaml", "1.0.0", "1.0.1-beta", false},                                                   // 'to' version with pre-release label
	{"migration-v1.0.0-alpha-v1.0.1-beta.yaml", "1.0.0-alpha", "1.0.1-beta", false},                                       // both versions with pre-release labels
	{"migration-v1.0.0-alpha.1-v1.0.1-beta.1.yaml", "1.0.0-alpha.1", "1.0.1-beta.1", false},                               // pre-release labels with additional identifiers
	{"migration-v1.0.0+20130313144700-v1.0.1+20130313144700.yaml", "1.0.0+20130313144700", "1.0.1+20130313144700", false}, // versions with build metadata
	{"migration-v1.0.0-v1.0.1.yaml", "1.0.0", "1.0.1", false},                                                             // valid migration with different 'from' and 'to' versions
	{"migration-v1.0.0-v1.0.0.yaml", "", "", true},                                                                        // 'from' version is equal to 'to' version
	{"migration-v1.0.1-v1.0.0.yaml", "", "", true},                                                                        // 'from' version is greater than 'to' version
	{"migration-v1.0.0-v1.0.yaml", "", "", true},                                                                          // 'to' version is not a valid semver
	{"migration-v1.0-v1.0.1.yaml", "", "", true},                                                                          // 'from' version is not a valid semver
	{"migration-v1.0.0-v1.0.1", "", "", true},                                                                             // file does not have an extension
	{"migration-v1.0.0-v1.0.1.txt", "", "", true},                                                                         // file does not have .yaml extension
	{"migration-v1.0.0-v1.0.1.txt.yaml", "", "", true},                                                                    // file does not have .yaml extension
	{"v1.0.0-v1.0.1.yaml", "", "", true},                                                                                  // file does not start with 'migration-'
	{"migration-1.0.0-1.0.1.yaml", "", "", true},                                                                          // versions do not start with 'v''
}

func TestNewMigration(t *testing.T) {
	for _, tt := range newMigrationTests {
		t.Run(tt.in, func(t *testing.T) {
			migration, err := NewMigration(tt.in)

			if tt.hasErr {
				assert.Error(t, err)
				assert.Nil(t, migration)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, migration)
				if migration != nil {
					assert.Equal(t, tt.outFromVer, migration.from.String())
					assert.Equal(t, tt.outToVer, migration.to.String())
				}
			}
		})
	}
}
