package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"helm-migrate-values/internal"
	"helm-migrate-values/pkg"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const cmdDescription = `Migrate the user-supplied values of a Helm release to the current version of its chart's values schema.
The result is an output string that can be saved to a file and used as input to helm upgrade
e.g.
	
	helm migrate-values myRedis redis -o new_values.yaml
	helm upgrade -f new_values myRedis

The migration is defined as a series of schema transformations using Go Templating in the form of YAML files. These YAML files
should be stored in the chart itself as:

	{CHART_DIR}/value-migrations/to-v{VERSION_TO}.yaml

		CHART_DIR is directory in which the Helm chart is defined
		VERSION_TO represents the major version of the values schema. These should use the same versioning as the chart itself.

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

func NewRootCmd(actionConfig *action.Configuration, settings *cli.EnvSettings, out io.Writer, log internal.Logger) (*cobra.Command, error) {
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

	runner := newRunner(actionConfig, flags, settings, out, &outputFile, log)
	cmd.RunE = runner

	return cmd, nil
}

func newRunner(actionConfig *action.Configuration, flags *pflag.FlagSet, settings *cli.EnvSettings, out io.Writer, outputFile *string, log internal.Logger) func(cmd *cobra.Command, args []string) error {
	// We use the install action for locating the chart
	var installAction = action.NewInstall(actionConfig)
	var listAction = action.NewList(actionConfig)

	internal.AddChartPathOptionsFlags(flags, &installAction.ChartPathOptions)

	return func(cmd *cobra.Command, args []string) error {
		helmDriver := os.Getenv("HELM_DRIVER")
		if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), helmDriver, log.Debug); err != nil {
			return err
		}

		name, chart, err := nameAndChart(args)
		if err != nil {
			return err
		}

		chartDir, cleanupDirectory, err := internal.LocateChart(chart, installAction, settings, log)
		if err != nil {
			return fmt.Errorf("failed to download chart: %w", err)
		}

		if cleanupDirectory {
			defer func() {
				log.Debug("Cleaning up extracted chart at %s", *chartDir)
				if err = os.RemoveAll(*chartDir); err != nil {
					err = fmt.Errorf("failed to cleanup extracted chart: %w", err)
				}
			}()
		}

		if chartDir != nil {
			log.Debug("Using chart at: %s", *chartDir)
		}

		release, err := internal.GetRelease(name, listAction)
		if err != nil {
			return err
		}

		if release != nil {
			log.Debug("Release is using chart: %s", release.Chart.Metadata.Name)
			log.Debug("Release is currently on chart version: %s", release.Chart.Metadata.Version)

			if log.IsDebug {
				value, err := yaml.Marshal(release.Config)
				if err != nil {
					log.Debug("Release has the following user-supplied values:\n%s", value)
				}
			}

		}

		migrationRequired := release.Config != nil && len(release.Config) > 0

		if migrationRequired {

			majorVerRegEx := regexp.MustCompile(`^(\d+)\..*`)
			matches := majorVerRegEx.FindStringSubmatch(release.Chart.Metadata.Version)
			if len(matches) == 0 {
				return fmt.Errorf("failed to extract major version from chart version: %s", release.Chart.Metadata.Version)
			}
			relMajorVer, _ := strconv.Atoi(matches[1])
			migrationsPath := filepath.Join(*chartDir, release.Chart.Name(), "value-migrations")

			// vTo is always nil, because it really only makes sense include migrations up to the current chart version.
			// The migrations library does support migrating to a specific version though.
			migratedConfig, err := pkg.MigrateFromPath(release.Config, relMajorVer, nil, migrationsPath, log)
			if err != nil {
				return err
			}

			if len(migratedConfig) > 0 {

				migratedValues, err := yaml.Marshal(migratedConfig)
				if err != nil {
					return fmt.Errorf("migrated values are in an invalid format: %w", err)
				}

				if *outputFile == "" {
					message := fmt.Sprintf("Migrated user-supplied values for release %s:\n%s", name, string(migratedValues))
					if _, err = fmt.Fprint(out, message); err != nil {
						return fmt.Errorf("error writing migrated values to standard output: %w", err)
					}
				} else {
					if err = writeOutputValues(err, *outputFile, migratedValues); err != nil {
						return err
					}
				}
			} else {
				migrationRequired = false
			}

		}

		if !migrationRequired {
			fmt.Printf("No migration required for release %s", name)
		}

		return nil
	}
}

func writeOutputValues(err error, outputFile string, migratedValues []byte) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output values file: %w", err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(fmt.Errorf("error closing output values file: %w", err))
		}
	}(f)

	_, err = f.Write(migratedValues)
	if err != nil {
		return fmt.Errorf("error writing migrated values to output values file: %w", err)
	}

	if err = f.Sync(); err != nil {
		return fmt.Errorf("error writing migrated values to output values file: %w", err)
	}

	return nil
}

func nameAndChart(args []string) (string, string, error) {
	if len(args) > 2 {
		return args[0], args[1], errors.Errorf("expected at most two arguments, unexpected arguments: %v", strings.Join(args[2:], ", "))
	}

	return args[0], args[1], nil
}
