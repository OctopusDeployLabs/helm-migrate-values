package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"os"
	"strings"
)

func pullAndExtractChart(chart *string, client *action.Install) (*string, error) {
	//set up the registry client with appropriate information
	registryClient, err := newRegistryClient(client.CertFile, client.KeyFile, client.CaFile, client.InsecureSkipTLSverify, client.PlainHTTP)
	if err != nil {

		return nil, fmt.Errorf("missing registry client: %w", err)
	}

	client.SetRegistryClient(registryClient)

	chartPath, err := client.ChartPathOptions.LocateChart(*chart, settings)
	if err != nil {
		return nil, err
	}

	debug("Chart path: %s", chartPath)

	//if the located chart path is not a .tgz file, it must be a local directory
	if !strings.HasSuffix(chartPath, ".tgz") {
		return &chartPath, nil
	}

	//otherwise, unpack to a known temp directory
	//create a temp dir where this chart will be unpacked
	tempPath, err := os.MkdirTemp("", "chart-*")
	if err != nil {
		panic(err)
	}

	err = chartutil.ExpandFile(tempPath, chartPath)
	if err != nil {
		return nil, err
	}

	debug("Unpacked chart to %s", tempPath)

	return &tempPath, nil
}

func nameAndChart(args []string) (string, string, error) {
	if len(args) > 2 {
		return args[0], args[1], errors.Errorf("expected at most two arguments, unexpected arguments: %v", strings.Join(args[2:], ", "))
	}

	return args[0], args[1], nil
}
