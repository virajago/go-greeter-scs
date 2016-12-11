package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
}

func main() {
	app := cli.NewApp()
	app.Name = "eureka-cli"
	app.Usage = "Command-line client for Netflix Eureka"

	app.Commands = []cli.Command{
		registerCmd,
		deregisterCmd,
		heartbeatCmd,
		instancesCmd,
		overrideCmd,
		removeOverrideCmd,
	}

	app.Run(os.Args)
}
