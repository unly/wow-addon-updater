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

const (
	versionFile string      = ".versions"
	fileMod     os.FileMode = os.FileMode(0666)
)

var addonSources []UpdateSource = []UpdateSource{
	sources.NewGitHubSource(),
}

type Updater struct {
	classic gameUpdater
	retail  gameUpdater
}

type gameUpdater struct {
	config   config.WowConfig
	versions map[string]string
}

type UpdateSource interface {
	GetURLRegex() *regexp.Regexp
	GetLatestVersion(addon string) (string, error)
	DownloadAddon(addon, dir string) error
}

type Addon struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Versions struct {
	Classic []Addon `yaml:"classic"`
	Retail  []Addon `yaml:"retail"`
}

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

func (u *Updater) UpdateAddons() error {
	defer func() {
		if err := saveVersionsFile(u, versionFile, fileMod); err != nil {
			log.Panicf("failed to write versions file: %v", err)
		}
	}()

	err := u.retail.updateAddons()
	if err != nil {
		return err
	}
	err = u.classic.updateAddons()

	return err
}

func (u *gameUpdater) updateAddons() error {
	for _, addon := range u.config.AddOns {
		source, err := getSource(addon)
		if err != nil {
			return err
		}

		err = u.updateAddon(addon, source)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *gameUpdater) getCurrentVersion(addon string) string {
	version, _ := u.versions[addon]
	return version
}

func (u *gameUpdater) setCurrentVersion(addon, version string) {
	u.versions[addon] = version
}

func (u *gameUpdater) updateAddon(addon string, source UpdateSource) error {
	log.Printf("updating addon: %s\n", addon)

	currentVersion := u.getCurrentVersion(addon)

	latestVersion, err := source.GetLatestVersion(addon)
	if err != nil {
		return err
	}

	if currentVersion == latestVersion {
		log.Println("no need for an update")
		return nil
	}

	err = source.DownloadAddon(addon, u.config.Path)
	if err != nil {
		return err
	}

	u.setCurrentVersion(addon, latestVersion)
	log.Printf("updated to version: %s\n", latestVersion)
	return nil
}

func getSource(addon string) (UpdateSource, error) {
	for _, source := range addonSources {
		if source.GetURLRegex().Match([]byte(addon)) {
			return source, nil
		}
	}

	return nil, fmt.Errorf("addon url: %s is not supported", addon)
}

func readVersionsFile(path string) (Versions, error) {
	var versions Versions

	if !util.FileExists(path) {
		return versions, nil
	}

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return versions, err
	}

	err = yaml.Unmarshal(yamlFile, &versions)

	return versions, err
}

func saveVersionsFile(u *Updater, path string, mode os.FileMode) error {
	versions := Versions{
		Classic: getAddons(&u.classic),
		Retail:  getAddons(&u.retail),
	}

	out, err := yaml.Marshal(versions)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, out, mode)

	return err
}

func mapAddonVersions(addons []Addon) map[string]string {
	addonVersions := make(map[string]string, len(addons))
	for _, addon := range addons {
		addonVersions[addon.Name] = addon.Version
	}

	return addonVersions
}

func getAddons(g *gameUpdater) []Addon {
	addons := make([]Addon, len(g.versions))
	i := 0

	for addon, version := range g.versions {
		addons[i] = Addon{
			Name:    addon,
			Version: version,
		}
		i++
	}

	return addons
}
