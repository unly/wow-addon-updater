package updater

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/unly/wow-addon-updater/config"
	"github.com/unly/wow-addon-updater/updater/sources"
	"github.com/unly/wow-addon-updater/util"
	"gopkg.in/yaml.v3"
)

const versionFile string = ".versions"

var addonSources []UpdateSource = []UpdateSource{
	sources.NewGitHubSource(),
	sources.NewTukUISource(),
}

// Updater is the main struct to update all addons for both
// retail and classic installations.
type Updater struct {
	classic gameUpdater
	retail  gameUpdater
}

type gameUpdater struct {
	config   config.WowConfig
	versions map[string]addon
}

// UpdateSource can be a possible source to get WoW addons from
type UpdateSource interface {
	// GetURLRegex returns a regular expression that matches a URL the source can handle
	GetURLRegex() *regexp.Regexp
	// GetLatestVersion returns the latest version for the given addon URL
	// Returns an empty string if there is no version
	GetLatestVersion(addonURL string) (string, error)
	// DownloadAddon downloads and extracts the addon to the given directory
	DownloadAddon(addonURL, dir string) error
}

type addon struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type versions struct {
	Classic []addon `yaml:"classic"`
	Retail  []addon `yaml:"retail"`
}

// NewUpdater returns a pointer to a newly created Updater or an error if it fails to read in
// the version tracking file.
// Uses the config.Config to identify the addons
func NewUpdater(config config.Config) (*Updater, error) {
	readVersions, err := readVersionsFile(versionFile)
	if err != nil {
		return nil, err
	}

	return &Updater{
		classic: gameUpdater{
			config:   config.Classic,
			versions: mapAddonVersions(readVersions.Classic),
		},
		retail: gameUpdater{
			config:   config.Retail,
			versions: mapAddonVersions(readVersions.Retail),
		},
	}, nil
}

// UpdateAddons updates all the addons given in the configuration
func (u *Updater) UpdateAddons() error {
	defer func() {
		if err := saveVersionsFile(u, versionFile); err != nil {
			log.Panicf("failed to write versions file: %v", err)
		}
	}()

	err := u.retail.updateAddons()
	if err != nil {
		return err
	}
	err = u.classic.updateAddons()
	if err != nil {
		return err
	}

	return nil
}

func (g *gameUpdater) updateAddons() error {
	for _, addon := range g.config.AddOns {
		source, err := getSource(addon)
		if err != nil {
			return err
		}

		err = g.updateAddon(addon, source)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *gameUpdater) getCurrentVersion(addonURL string) string {
	add, ok := g.versions[addonURL]
	if !ok {
		return ""
	}

	return add.Version
}

func (g *gameUpdater) setCurrentVersion(addonURL, version string) {
	add, ok := g.versions[addonURL]
	if !ok {
		add = addon{
			Name: addonURL,
		}
	}

	add.Version = version
	g.versions[addonURL] = add
}

func (g *gameUpdater) updateAddon(addonURL string, source UpdateSource) error {
	log.Printf("updating addon: %s\n", addonURL)

	currentVersion := g.getCurrentVersion(addonURL)

	latestVersion, err := source.GetLatestVersion(addonURL)
	if err != nil {
		return err
	}

	if currentVersion == latestVersion {
		log.Println("no need for an update")
		return nil
	}

	err = source.DownloadAddon(addonURL, g.config.Path)
	if err != nil {
		return err
	}

	g.setCurrentVersion(addonURL, latestVersion)
	log.Printf("updated to version: %s\n", latestVersion)
	return nil
}

func getSource(addonURL string) (UpdateSource, error) {
	for _, source := range addonSources {
		if source.GetURLRegex().MatchString(addonURL) {
			return source, nil
		}
	}

	return nil, fmt.Errorf("addon url: %s is not supported", addonURL)
}

func readVersionsFile(path string) (versions, error) {
	var vers versions

	if !util.FileExists(path) {
		return vers, nil
	}

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return vers, err
	}

	err = yaml.Unmarshal(yamlFile, &vers)

	return vers, err
}

func saveVersionsFile(u *Updater, path string) error {
	vers := versions{
		Classic: getAddons(&u.classic),
		Retail:  getAddons(&u.retail),
	}

	out, err := yaml.Marshal(vers)
	if err != nil {
		return err
	}

	return util.WriteToHiddenFile(path, out, os.FileMode(0666))
}

func mapAddonVersions(addons []addon) map[string]addon {
	addonVersions := make(map[string]addon, len(addons))
	for _, addon := range addons {
		addonVersions[addon.Name] = addon
	}

	return addonVersions
}

func getAddons(g *gameUpdater) []addon {
	addons := make([]addon, len(g.versions))
	i := 0

	for _, addon := range g.versions {
		addons[i] = addon
		i++
	}

	return addons
}
