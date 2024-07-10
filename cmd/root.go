package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"log"
	"os"
)

var actionConfig = new(action.Configuration)
var client = action.NewInstall(actionConfig)
var settings = cli.New()

var rootCmd = &cobra.Command{
	Use:   "migrate-values [RELEASE] [CHART] [flags]",
	Short: "helm migrator for values schemas",
	Long: `Migrate the user-supplied values of a Helm release to the current version of its chart's values schema.
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
`,
	Args: cobra.MinimumNArgs(2),
	RunE: RunCmd,
}

var outputFile string

func RunCmd(cmd *cobra.Command, args []string) error {
	_, chart, err := nameAndChart(args)
	if err != nil {
		return err
	}

	chartDir, err := pullAndExtractChart(&chart, client)
	if err != nil {
		return fmt.Errorf("failed to download chart: %w", err)
	}

	debug("Using chart at: %s", *chartDir)

	//TODO: Get the chart and version associated with the release
	//TODO: Get the current release values
	//TODO: Load the transformations from the migrations directory
	//TODO: Apply the transformations (if needed) to the current values w.r.t the current chart version
	//TODO: Output the result or save to a file location

	return nil
}

func init() {
	log.SetFlags(log.Lshortfile)

	rootCmd.Flags().StringVarP(&outputFile, "output-file", "o", "",
		"The output file to which the result is saved. Standard output is used if this option is not set.")

	settings.AddFlags(rootCmd.PersistentFlags())

	// add the default chart path options
	addChartPathOptionsFlags(rootCmd.Flags(), &client.ChartPathOptions)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
