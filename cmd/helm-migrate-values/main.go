package main

import (
	"fmt"
	"github.com/spf13/cobra"
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

	cmd, err := newRootCmd(actionConfig, os.Stdout, os.Args[1:])
	if err != nil {
		warning("%+v", err)
		os.Exit(1)
	}

	cobra.OnInitialize(func() {
		helmDriver := os.Getenv("HELM_DRIVER")
		if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), helmDriver, debug); err != nil {
			log.Fatal(err)
		}
	})

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
