package pkg

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"helm.sh/helm/v3/pkg/time"
	"log"
	"testing"
)

func TestMigrator_IntegrationTests(t *testing.T) {

	defaultValuesV1 := `agent:
  targetEnvironments: []`

	customValuesV1 := `agent:
  targetEnvironments:
  - Development
  - Test
  - Prod`

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
    environments: [{{ .agent.targetEnvironments | quoteEach | join ","}}]`

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
	if err != nil {
		t.Errorf("Error unmarshalling custom values: %v", err)
		return
	}

	rel1, err := install.Run(chV1, cfgMap)
	if err != nil {
		t.Errorf("Error installing chart: %v", err)
		return
	}
	//rel := getMockRelease(chV1, customValuesV1, t)
	//rel.Name = "release-2"

	//Make a migration
	ms := &MockMigrationSource{}
	ms.AddMigrationData(2, migration)

	// Migrate the release user values (config)
	migratedValues, err := Migrate(rel1.Config, nil, ms)
	if err != nil {
		t.Errorf("Error migrating values: %v", err)
		return
	}

	migratedValuesMap, err := yamlUnmarshal(t, *migratedValues)
	if err != nil {
		t.Errorf("Error unmarshalling migrated values: %v", err)
		return
	}

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

	//rel2 := getMockRelease(chV1, *migratedValues, t)
	//rel2.Name = "release-2"

	upAction := action.NewUpgrade(config)

	//upAction := upgradeAction(config)
	/*
		err = config.Releases.Create(rel)
		if err != nil {
			t.Errorf("Error creating release: %v", err)
		}                   */
	//is.NoError(err)

	upAction.ResetThenReuseValues = true
	// setting newValues and upgrading
	res, err := upAction.Run(rel1.Name, chV2, migratedValuesMap)

	if err != nil {
		t.Errorf("Error upgrading chart: %v", err)
		return
	}
	if res == nil {
		assert.Fail(t, "Release is nil")
		return
	}

	// Now make sure it is actually upgraded
	releases, err := config.Releases.ListReleases()
	log.Println(releases[0].Name)
	log.Println(releases[0].Version)
	updatedRes, err := config.Releases.Get(res.Name, 2)
	//is.NoError(err)

	if updatedRes == nil {
		assert.Fail(t, "Updated Release is nil")
		return
	}

	assert.Equal(t, release.StatusDeployed, updatedRes.Info.Status)

	updatedVals, _ := yaml.Marshal(updatedRes.Config)
	t.Log(string(updatedVals))
	//	assert.Equal(t, expectedValues, updatedRes.Config)

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

func upgradeAction(config *action.Configuration) *action.Upgrade {
	upAction := action.NewUpgrade(config)
	upAction.Namespace = "spaced"

	return upAction
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

type chartOptions struct {
	*chart.Chart
}

type chartOption func(*chartOptions)

func buildChart(opts ...chartOption) *chart.Chart {
	c := &chartOptions{
		Chart: &chart.Chart{
			// TODO: This should be more complete.
			Metadata: &chart.Metadata{
				APIVersion: "v1",
				Name:       "hello",
				Version:    "0.1.0",
			},
			// This adds a basic template and hooks.
			Templates: []*chart.File{
				{Name: "templates/hello", Data: []byte("hello: world")},
				{Name: "templates/hooks", Data: []byte(manifestWithHook)}, //we don't need this
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c.Chart
}

var manifestWithHook = `kind: ConfigMap
metadata:
  name: test-cm
  annotations:
    "helm.sh/hook": post-install,pre-delete,post-upgrade
data:
  name: value`

type LogWriter struct {
}

func (*LogWriter) Write(p []byte) (n int, err error) {
	log.Print(string(p))
	return len(p), nil
}

func getMockRelease(ch *chart.Chart, cfg string, t *testing.T) *release.Release {

	date := time.Unix(374072400, 0).UTC()
	info := &release.Info{
		FirstDeployed: date,
		LastDeployed:  date,
		Status:        release.StatusDeployed,
	}

	var cfgMap map[string]interface{}
	err := yaml.Unmarshal([]byte(cfg), cfgMap)
	if err != nil {
		t.Errorf("Error unmarshalling default values: %v", err)
	}

	// TODO: narrow down what we actually need here
	return &release.Release{
		Name:      "mock-release",
		Info:      info,
		Chart:     ch,
		Config:    cfgMap, //map[string]interface{}{"name": "value"},
		Version:   1,
		Namespace: "namespace",
		Manifest:  MockManifest,
	}
}

var MockManifest = `apiVersion: v1
kind: Secret
metadata:
  name: fixture
`

/*

	/*
	existingValues := map[string]interface{}{
		"name":        "value",
		"maxHeapSize": "128m",
		"replicas":    2,
	}
	newValues := map[string]interface{}{
		"name":        "newValue",
		"maxHeapSize": "512m",
		"cpu":         "12m",
	}
	expectedValues := map[string]interface{}{
		"name":        "newValue",
		"maxHeapSize": "512m",
		"cpu":         "12m",
		"replicas":    2,
	}       */
