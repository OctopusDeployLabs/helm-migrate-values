package main

import (
	"fmt"
	"helm-migrate-values/internal"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"log"
	"os"
)

// this loads settings from environment variables as well as the command line flags
var settings = cli.New()

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	var actionConfig = new(action.Configuration)
	logger := *internal.NewLogger(settings.Debug)
	logger.Debug("Debug mode enabled")

	cmd, err := NewRootCmd(actionConfig, settings, os.Stdout, logger)
	if err != nil {
		logger.Warning("%+v", err)
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
