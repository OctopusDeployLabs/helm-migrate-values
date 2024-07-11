package main

import (
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"os"
	"strings"
)

// Locates the chart. If this is a remote (OCI/Repo URL) it downloads the chart and extract the tgz to a temporary file
func locateChart(chart string, client *action.Install) (*string, bool, error) {
	err := setupRegistryClient(client)
	if err != nil {
		return nil, false, err
	}

	chartPath, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return nil, false, err
	}

	debug("Chart path: %s", chartPath)

	//if the located chart path is not a .tgz file, it must be a local directory
	if !strings.HasSuffix(chartPath, ".tgz") {
		debug("Chart is not", chartPath)
		return &chartPath, false, nil
	}

	//otherwise, unpack to a known temp directory
	//create a temp dir where this chart will be unpacked
	tempPath, err := os.MkdirTemp("", "chart-*")
	if err != nil {
		return nil, false, err
	}

	err = chartutil.ExpandFile(tempPath, chartPath)
	if err != nil {
		return nil, false, err
	}

	debug("Unpacked chart to %s", tempPath)

	return &tempPath, true, nil
}

func setupRegistryClient(client *action.Install) error {
	//set up the registry client with appropriate information
	registryClient, err := newRegistryClient(client.CertFile, client.KeyFile, client.CaFile, client.InsecureSkipTLSverify, client.PlainHTTP)

	if err == nil {
		client.SetRegistryClient(registryClient)
	}

	return err
}

func nameAndChart(args []string) (string, string, error) {
	if len(args) > 2 {
		return args[0], args[1], errors.Errorf("expected at most two arguments, unexpected arguments: %v", strings.Join(args[2:], ", "))
	}

	return args[0], args[1], nil
}
