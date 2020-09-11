package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/unly/wow-addon-updater/config"
	"github.com/unly/wow-addon-updater/updater"
)

const configPath string = "config.yaml"

func main() {
	path := flag.String("c", configPath, "path to the config file")
	flag.Parse()
	if path == nil {
		log.Panicf("no configuration file to read in")
	}

	log.Println("starting the wow addon manager")

	conf, err := config.ReadConfig(*path)
	if err != nil {
		log.Panicf("failed to read in the config file: %v\n", err)
	}

	updater, err := updater.NewUpdater(conf)
	if err != nil {
		log.Panicf("failed to initialize the updater: %v\n", err)
	}

	err = updater.UpdateAddons()
	if err != nil {
		log.Panicf("failed to update addon versions: %v\n", err)
	}

	log.Println("enjoy the updates!")
	log.Println("press Enter to quit")
	fmt.Scanln()
}
