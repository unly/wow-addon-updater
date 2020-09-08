package main

import (
	"fmt"
	"log"

	"github.com/unly/wow-addon-updater/config"
	"github.com/unly/wow-addon-updater/updater"
)

const configPath string = "config.yaml"

func main() {
	log.Println("starting the wow addon manager")

	conf, err := config.ReadConfig(configPath)
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
