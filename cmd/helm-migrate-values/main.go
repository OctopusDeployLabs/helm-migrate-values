package main

import (
	"fmt"
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

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
