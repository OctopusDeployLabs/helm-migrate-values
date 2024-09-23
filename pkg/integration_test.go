package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"testing"
)

var testCases = []struct {
	name       string
	chart1Path string
	chart2Path string
}{
	{
		name:       "migrate charts in folders",
		chart1Path: "test-charts/v1/", // Note that these aren't necessarily valid charts, they're just what we need for testing.
		chart2Path: "test-charts/v2/",
	},
	{
		name:       "migrate charts in tgz",
		chart1Path: "test-charts/my-chart-1.0.0.tgz",
		chart2Path: "test-charts/my-chart-2.0.0.tgz",
	},
}

func TestMigrator_IntegrationTests(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			migrateCharts(t, tc.chart1Path, tc.chart2Path)
		})
	}
}

func migrateCharts(t *testing.T, chartV1Path, chartV2Path string) {

	is := assert.New(t)
	req := require.New(t)

	customValues := map[string]interface{}{
		"myKey": "myValue",
		"project": map[string]interface{}{
			"targetEnvironments": []interface{}{"Development", "Test", "Prod"},
		},
	}

	expected := map[string]interface{}{
		"project": map[interface{}]interface{}{
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

	// Migrate the release user values (config)
	migratedValues, err := MigrateFromPath(rel1.Config, 1, nil, "test-charts/v2/value-migrations/", *NewLogger(false))
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
