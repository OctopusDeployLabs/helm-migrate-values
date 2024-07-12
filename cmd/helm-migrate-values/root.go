package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm-migrate-values/pkg"
	"helm.sh/helm/v3/pkg/action"
	"io"
	"os"
)

const cmdDescription = `Migrate the user-supplied values of a Helm release to the current version of its chart's values schema.
The result is an output string that can be saved to a file and used as input to helm upgrade
e.g.
	
	helm migrate-values myRedis redis -o new_values.yaml
	helm upgrade -f new_values myRedis

The migration is defined as a series of schema transformations using Go Templating in the form of YAML files. These YAML files
should be stored in the chart itself as:

	{CHART_DIR}/value-migrations/{VERSION_FROM}-{VERSION_TO}.yaml

		CHART_DIR is directory in which the Helm chart is defined
		VERSION_FROM and VERSION_TO represent the versions of the values schemas. These should use the same versioning as the chart itself.

Arguments:
  RELEASE
    The name of the release you want to migrate.
  
  CHART
	The fully qualified name of the chart to use. This should match the chart used by the specified Helm release.
	The command will apply the migrations to the current values of the release and output the result as text.

The command will return an error if:
	- the specified release does use the specified chart
	- no migrations are defined in the chart
`

func newRootCmd(actionConfig *action.Configuration, out io.Writer, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "migrate-values [RELEASE] [CHART] [flags]",
		Short: "helm migrator for values schemas",
		Long:  cmdDescription,
		Args:  cobra.MinimumNArgs(2),
	}

	flags := cmd.PersistentFlags()
	settings.AddFlags(flags)
	var outputFile string
	flags.StringVarP(&outputFile, "output-file", "o", "",
		"The output file to which the result is saved. Standard output is used if this option is not set.")

	runner := newRunner(actionConfig, flags, outputFile)
	cmd.RunE = runner

	return cmd, nil
}

func newRunner(actionConfig *action.Configuration, flags *pflag.FlagSet, outputFile string) func(cmd *cobra.Command, args []string) error {
	// We use the install action for locating the chart
	var installAction = action.NewInstall(actionConfig)
	var listAction = action.NewList(actionConfig)

	addChartPathOptionsFlags(flags, &installAction.ChartPathOptions)

	return func(cmd *cobra.Command, args []string) error {
		helmDriver := os.Getenv("HELM_DRIVER")
		if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), helmDriver, debug); err != nil {
			return err
		}

		name, chart, err := nameAndChart(args)
		if err != nil {
			return err
		}

		chartDir, cleanupDirectory, err := locateChart(chart, installAction)
		if err != nil {
			return fmt.Errorf("failed to download chart: %w", err)
		}

		if cleanupDirectory {
			defer func() {
				if err = os.RemoveAll(*chartDir); err != nil {
					err = fmt.Errorf("failed to cleanup extracted chart: %w", err)
				}
			}()
		}

		if chartDir != nil {
			debug("Using chart at: %s", *chartDir)
		}

		release, err := getRelease(name, listAction)
		if err != nil {
			return err
		}

		if release != nil {
			debug("Release is using chart: %s", release.Chart.Metadata.Name)
			debug("Release is currently on chart version: %s", release.Chart.Metadata.Version)
			debug("Release has the values: %s", release.Config)
		}

		var fileSystem pkg.FileSystem = pkg.RealFileSystem{}
		migratedValues, err := pkg.Migrate(release.Config, release.Chart.Metadata.Version, nil, *chartDir+"/value/migrations/", fileSystem)

		//TODO: Apply the transformations (if needed) to the current values w.r.t the current chart version
		println(migratedValues)
		//TODO: Output the result or save to a file location

		return err
	}
}
