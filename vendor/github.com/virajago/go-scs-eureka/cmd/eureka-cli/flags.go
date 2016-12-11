package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"

	"github.com/codegangsta/cli"

	"github.com/st3v/go-eureka"
)

var appNameFlag = cli.StringFlag{
	Name:  "app, a",
	Value: "",
	Usage: "Application Name, required if instance ID has been specified",
}

var instanceIDFlag = cli.StringFlag{
	Name:  "instance, i",
	Value: "",
	Usage: "Instance ID",
}

var instanceFlag = cli.StringFlag{
	Name:  "instance, i",
	Value: "",
	Usage: "Path to an XML file that defines the Eureka instance",
}

var endpointsFlag = cli.StringSliceFlag{
	Name:  "endpoint, e",
	Value: &cli.StringSlice{},
	Usage: "Eureka service endpoint, e.g. http://127.0.0.1/eureka/v2",
}

func getEndpoints(c *cli.Context, cmd string) []string {
	endpoints := c.StringSlice("endpoint")
	if len(endpoints) == 0 {
		cli.ShowCommandHelp(c, cmd)
		log.Fatalln("--endpoint flag is required")
	}
	return endpoints
}

func getInstance(c *cli.Context, cmd string) *eureka.Instance {
	xmlPath := c.String("instance")
	if xmlPath == "" {
		cli.ShowCommandHelp(c, cmd)
		log.Fatalln("--instance flag is required")
	}

	data, err := ioutil.ReadFile(xmlPath)
	if err != nil {
		log.Fatalf("Error reading instance file: %s", err)
	}

	instance := new(eureka.Instance)
	if err := xml.Unmarshal(data, instance); err != nil {
		log.Fatalf("Error parsing instance file: %s", err)
	}

	return instance
}
