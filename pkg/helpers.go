package pkg

import (
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"log"
	"testing"
)

func yamlUnmarshal(customValuesV1 string) (map[string]interface{}, error) {
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
			t.Logf(format, v...)
		},
	}
}

type LogWriter struct {
}

func (*LogWriter) Write(p []byte) (n int, err error) {
	log.Print(string(p))
	return len(p), nil
}
