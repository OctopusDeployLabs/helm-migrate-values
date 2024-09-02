package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"log"
	"testing"
)

func TestMigrator_IntegrationTests(t *testing.T) {
	is := assert.New(t)
	req := require.New(t)

	customValuesV1 := `myKey: "myValue"
agent:
  targetEnvironments: ["Development", "Test", "Prod"]`

	chartV1Path := "test-charts/v1/"
	chartV2Path := "test-charts/v2/"

	migration := `agent:
  deploymentTarget:
    initial:
      environments: [{{ .agent.targetEnvironments | quoteEach | join ","}}]
myKey: null
`
	expected := `agent:
  deploymentTarget:
    initial:
      environments:
      - Development
      - Test
      - Prod
`

	chV1, err := loader.Load(chartV1Path)
	req.NoError(err, "Error loading chart v1")

	config := actionConfigFixture(t)
	install := action.NewInstall(config)
	install.Namespace = "default"
	install.ReleaseName = "release-1"

	cfgMap, err := YamlUnmarshal(customValuesV1)
	req.NoError(err, "Error unmarshalling custom values")

	rel1, err := install.Run(chV1, cfgMap)
	req.NoError(err, "Error installing chart v1")

	//Make a migration
	ms := &MemoryMigrationSource{}
	ms.AddMigrationData(2, migration)

	// Migrate the release user values (config)
	migratedValues, err := Migrate(rel1.Config, nil, ms)
	req.NoError(err, "Error migrating values")

	// Create a new chart using the default values v2 and template v2
	// apply migrated values to the new chart
	chV2, err := loader.Load(chartV2Path)
	req.NoError(err, "Error loading chart v2")

	upAction := action.NewUpgrade(config)
	upAction.ResetThenReuseValues = true

	rel2, err := upAction.Run(rel1.Name, chV2, migratedValues)
	req.NoError(err, "Error upgrading chart")
	req.NotNil(rel2, "Release was not created")

	// Now make sure it is actually upgraded
	releases, err := config.Releases.ListReleases()
	log.Println(releases[0].Name)
	log.Println(releases[0].Version)
	updatedRel, err := config.Releases.Get(rel2.Name, 2)
	req.NoError(err, "Error getting updated release")

	req.NotNil(updatedRel, "Updated Release is nil")

	is.Equal(release.StatusDeployed, updatedRel.Info.Status)

	updatedVals, _ := yaml.Marshal(updatedRel.Config)

	is.Equal(expected, string(updatedVals))
}

func YamlUnmarshal(customValuesV1 string) (map[string]interface{}, error) {
	var cfgMap map[string]interface{}
	err := yaml.Unmarshal([]byte(customValuesV1), &cfgMap)
	if err != nil {
		return nil, err
	}
	return cfgMap, nil
}

// from https://github.com/helm/helm/blob/main/pkg/action/action_test.go#L39
func actionConfigFixture(t *testing.T) *action.Configuration {
	t.Helper()

	registryClient, err := registry.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	return &action.Configuration{
		Releases:       storage.Init(driver.NewMemory()),
		KubeClient:     &kubefake.PrintingKubeClient{Out: &LogWriter{}},
		Capabilities:   chartutil.DefaultCapabilities,
		RegistryClient: registryClient,
		Log: func(format string, v ...interface{}) {
			t.Helper()
			//if *verbose {
			t.Logf(format, v...)
			//}
		},
	}
}

type LogWriter struct {
}

func (*LogWriter) Write(p []byte) (n int, err error) {
	log.Print(string(p))
	return len(p), nil
}
