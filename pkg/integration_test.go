package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
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

	defaultValuesV1 := `agent:
  targetEnvironments: []`

	customValuesV1 := `myKey: "myValue"
agent:
  targetEnvironments: ["Development", "Test", "Prod"]`

	templateV1 := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: "this-chart"
  namespace: {{ .Release.Namespace | quote }}
spec:
  template:
    spec:
      containers:
        - name: {{ .Chart.Name }}
          roles:
          - name: "TargetRole"
          value: {{ join "," .Values.agent.targetRoles | quote }}
          env:
          - name: "TargetEnvironment"
          value: {{ join "," .Values.agent.targetEnvironments | quote }}`

	defaultValuesV2 := `agent:
  target:
    environments: []`

	templateV2 := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: "this-chart"
  namespace: {{ .Release.Namespace | quote }}
spec:
  template:
    spec:
      containers:
        - name: {{ .Chart.Name }}
          env:
          target:
            environments: [{{ .Values.agent.target.environments}}]`

	migration := `agent:
  target:
    environments: [{{ .agent.targetEnvironments | quoteEach | join ","}}]
myKey: null
`

	// Make a chart, using  template v1  and default values v1
	chV1 := &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion: chart.APIVersionV1,
			Name:       "testUpgradeChart",
			Version:    "1.0.0",
		},
		Templates: []*chart.File{
			{Name: "templates/deployment.yaml", Data: []byte(templateV1)},
			{Name: "values.yaml", Data: []byte(defaultValuesV1)},
		},
	}

	config := actionConfigFixture(t)
	install := action.NewInstall(config)
	install.ReleaseName = "release-1"

	cfgMap, err := yamlUnmarshal(t, customValuesV1)
	req.NoError(err, "Error unmarshalling custom values")

	rel1, err := install.Run(chV1, cfgMap)
	req.NoError(err, "Error installing chart v1")

	//Make a migration
	ms := &MockMigrationSource{}
	ms.AddMigrationData(2, migration)

	// Migrate the release user values (config)
	migratedValues, err := Migrate(rel1.Config, nil, ms)
	req.NoError(err, "Error migrating values")

	migratedValuesMap, err := yamlUnmarshal(t, *migratedValues)
	req.NoError(err, "Error unmarshalling migrated values")

	// Create a new chart using the default values v2 and template v2
	// apply migrated values to the new chart
	chV2 := &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion: chart.APIVersionV1,
			Name:       "testUpgradeChart",
			Version:    "2.0.0",
		},
		Templates: []*chart.File{
			{Name: "templates/deployment.yaml", Data: []byte(templateV2)},
			{Name: "values.yaml", Data: []byte(defaultValuesV2)},
		},
	}

	upAction := action.NewUpgrade(config)
	upAction.ResetThenReuseValues = true

	rel2, err := upAction.Run(rel1.Name, chV2, migratedValuesMap)
	req.NoError(err, "Error upgrading chart")
	req.NotNil(rel2, "Release was not created")

	// Now make sure it is actually upgraded
	releases, err := config.Releases.ListReleases()
	log.Println(releases[0].Name)
	log.Println(releases[0].Version)
	updatedRel, err := config.Releases.Get(rel2.Name, 2)
	req.NoError(err, "Error getting updated release")

	if updatedRel == nil {
		assert.Fail(t, "Updated Release is nil")
		return
	}

	assert.Equal(t, release.StatusDeployed, updatedRel.Info.Status)

	updatedVals, _ := yaml.Marshal(updatedRel.Config)
	t.Log(string(updatedVals))
	//	assert.Equal(t, expectedValues, updatedRel.Config)

}

func yamlUnmarshal(t *testing.T, customValuesV1 string) (map[string]interface{}, error) {
	var cfgMap map[string]interface{}
	err := yaml.Unmarshal([]byte(customValuesV1), &cfgMap)
	if err != nil {
		t.Errorf("Error unmarshalling default values: %v", err)
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

	releases := storage.Init(driver.NewMemory())

	return &action.Configuration{
		Releases:       releases,
		KubeClient:     &kubefake.PrintingKubeClient{Out: &LogWriter{}}, //&kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: io.Discard}},
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
