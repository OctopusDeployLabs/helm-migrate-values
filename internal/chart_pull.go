package internal

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"os"
	"strings"
)

// Locates the chart. If this is a remote (OCI/Repo URL) it downloads the chart and extract the tgz to a temporary file
func LocateChart(chart string, client *action.Install, settings *cli.EnvSettings, log Logger) (*string, bool, error) {
	err := setupRegistryClient(client, settings)
	if err != nil {
		return nil, false, err
	}

	log.Debug("Locating chart %s", chart)
	chartPath, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return nil, false, err
	}

	log.Debug("Chart path: %s", chartPath)

	//if the located chart path is not a .tgz file, it must be a local directory
	if !strings.HasSuffix(chartPath, ".tgz") {
		log.Debug("Chart is not a .tgz, using path %s", chartPath)
		return &chartPath, false, nil
	}

	//otherwise, unpack to a known temp directory
	//create a temp dir where this chart will be unpacked
	tempPath, err := os.MkdirTemp("", "chart-*")
	if err != nil {
		return nil, false, err
	}

	if err = chartutil.ExpandFile(tempPath, chartPath); err != nil {
		return nil, false, err
	}

	log.Debug("Unpacked chart to %s", tempPath)

	return &tempPath, true, nil
}

func setupRegistryClient(client *action.Install, settings *cli.EnvSettings) error {
	//set up the registry client with appropriate information
	registryClient, err := newRegistryClient(client.CertFile, client.KeyFile, client.CaFile, client.InsecureSkipTLSverify, client.PlainHTTP, settings.RegistryConfig, settings.Debug)

	if err == nil {
		client.SetRegistryClient(registryClient)
	}

	return err
}
