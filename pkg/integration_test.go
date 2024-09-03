package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"testing"
)

func TestMigrator_IntegrationTests(t *testing.T) {
	is := assert.New(t)
	req := require.New(t)

	customValues := map[string]interface{}{
		"myKey": "myValue",
		"agent": map[string]interface{}{
			"targetEnvironments": []interface{}{"Development", "Test", "Prod"},
		},
	}

	chartV1Path := "test-charts/v1/" // Note that these aren't necessarily valid charts, they're just what we need for testing.
	chartV2Path := "test-charts/v2/"

	migration := map[string]interface{}{
		"agent": map[string]interface{}{
			"deploymentTarget": map[string]interface{}{
				"initial": map[string]interface{}{
					"environments": []string{"Development", "Test", "Prod"},
				},
			},
		},
		"myKey": nil,
	}

	expected := map[string]interface{}{
		"agent": map[interface{}]interface{}{
			"deploymentTarget": map[interface{}]interface{}{
				"initial": map[interface{}]interface{}{
					"environments": []interface{}{"Development", "Test", "Prod"},
				},
			},
		},
	}

	// Load v1 chart
	chV1, err := loader.Load(chartV1Path)
	req.NoError(err, "Error loading chart v1")

	config := actionConfigFixture(t)
	install := action.NewInstall(config)
	install.Namespace = "default"
	install.ReleaseName = "release-1"

	// Install the v1 chart with the custom values
	rel1, err := install.Run(chV1, customValues)
	req.NoError(err, "Error installing chart v1")

	// Make a migration
	ms := &MemoryMigrationSource{}
	ms.AddMigrationData(2, migration)

	// Migrate the release user values (config)
	migratedValues, err := Migrate(rel1.Config, nil, ms)
	req.NoError(err, "Error migrating values")

	// Load the v2 chart
	chV2, err := loader.Load(chartV2Path)
	req.NoError(err, "Error loading chart v2")

	upAction := action.NewUpgrade(config)
	upAction.ResetThenReuseValues = true

	// Apply migrated values to the new chart
	rel2, err := upAction.Run(rel1.Name, chV2, migratedValues)
	req.NoError(err, "Error upgrading chart")
	req.NotNil(rel2, "Release was not created")

	// Now make sure it is actually upgraded
	updatedRel, err := config.Releases.Get(rel2.Name, 2)
	req.NoError(err, "Error getting updated release")
	req.NotNil(updatedRel, "Updated Release is nil")

	is.Equal(release.StatusDeployed, updatedRel.Info.Status)
	is.Equal(expected, updatedRel.Config)
}
