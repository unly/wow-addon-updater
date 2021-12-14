package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/unly/wow-addon-updater/config"
	"github.com/unly/wow-addon-updater/updater"
	"github.com/unly/wow-addon-updater/updater/sources/github"
	"github.com/unly/wow-addon-updater/updater/sources/tukui"
	"github.com/unly/wow-addon-updater/updater/sources/wowinterface"
	"github.com/unly/wow-addon-updater/util"
)

const (
	configPath string = "config.yaml"
)

var (
	addonSources = getSources()
	versionsPath = ".versions"
)

func main() {
	exitCode := 0
	err := runAndRecover()
	if err != nil {
		log.Printf("the WoW updater crashed... error: %v\n", err)
		exitCode = 1
	}

	log.Println("press Enter to quit")
	_, _ = fmt.Scanln()
	os.Exit(exitCode)
}

func runAndRecover() (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("panic: %v", v)
		}
	}()

	err = run()
	return
}

func run() error {
	defer closeSources(addonSources)

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)
	path := flag.String("c", configPath, "path to the config file")
	flag.Parse()
	if path == nil {
		return fmt.Errorf("no configuration file to read in")
	}

	log.Println("starting the wow addon manager")

	if !util.FileExists(*path) {
		return generateDefaultConfig(*path)
	}

	conf, err := config.ReadConfig(*path)
	if err != nil {
		return fmt.Errorf("failed to read in the config file: %v", err)
	}

	updater, err := updater.NewUpdater(conf, addonSources, versionsPath)
	if err != nil {
		return fmt.Errorf("failed to initialize the updater: %v", err)
	}

	err = updater.UpdateAddons()
	if err != nil {
		return fmt.Errorf("failed to update addon versions: %v", err)
	}

	log.Println("enjoy the updates!")

	return nil
}

func generateDefaultConfig(path string) error {
	err := config.CreateDefaultConfig(path)
	if err != nil {
		return fmt.Errorf("failed to create a default config file: %v", err)
	}

	log.Printf("no config file found. created empty config at: %s\n", path)
	log.Println("see https://github.com/unly/wow-addon-updater for information")

	return nil
}

func getSources() []updater.UpdateSource {
	sources := make([]updater.UpdateSource, 0)
	tukuiSource, err := tukui.New(new(http.Client))
	if err != nil {
		panic(err)
	}
	sources = append(sources, tukuiSource)
	wowinterfaceSource, err := wowinterface.New(new(http.Client))
	if err != nil {
		panic(err)
	}
	sources = append(sources, wowinterfaceSource)
	githubSource, err := github.New(new(http.Client))
	if err != nil {
		panic(err)
	}
	sources = append(sources, githubSource)
	return sources
}

func closeSources(sources []updater.UpdateSource) {
	for _, s := range sources {
		s.Close()
	}
}
