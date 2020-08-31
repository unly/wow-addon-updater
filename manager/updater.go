package manager

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/unly/wow-addon-updater/config"
	"github.com/unly/wow-addon-updater/manager/sources"
	"gopkg.in/yaml.v3"
)

const versionFile string = ".versions"

type updater struct {
	config   config.WowConfig
	versions map[string]string
	sources  []UpdateSource
}

type UpdateSource interface {
	GetURLRegex() *regexp.Regexp
	GetLatestVersion(addon string) (string, error)
	DownloadAddon(addon string) (string, error)
}

type Addon struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type AddonVersions struct {
	Addons []Addon `yaml:"addons"`
}

func NewUpdater(config config.WowConfig) (*updater, error) {
	sources := []UpdateSource{
		sources.NewGitHubSource(),
	}

	readVersions, err := readVersionsFile(versionFile)
	if err != nil {
		return nil, err
	}

	versions := make(map[string]string, len(readVersions.Addons))
	for _, addon := range readVersions.Addons {
		versions[addon.Name] = addon.Version
	}

	return &updater{
		config:   config,
		versions: versions,
		sources:  sources,
	}, nil
}

func (u *updater) getCurrentVersion(addon string) (string, error) {
	version, _ := u.versions[addon]
	return version, nil
}

func (u *updater) setCurrentVersion(addon, version string) {
	u.versions[addon] = version
}

func (u *updater) saveVersionsFile() error {
	addons := make([]Addon, len(u.versions))
	i := 0

	for addon, version := range u.versions {
		addons[i] = Addon{
			Name:    addon,
			Version: version,
		}
		i++
	}

	addonVersions := AddonVersions{
		Addons: addons,
	}

	out, err := yaml.Marshal(addonVersions)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(versionFile, out, os.FileMode(0666))

	return err
}

func (u *updater) UpdateAddons() error {
	defer u.saveVersionsFile()

	for _, addon := range u.config.AddOns {
		source, err := u.getSource(addon)
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

func (u *updater) getSource(addon string) (UpdateSource, error) {
	for _, source := range u.sources {
		if source.GetURLRegex().Match([]byte(addon)) {
			return source, nil
		}
	}

	return nil, fmt.Errorf("addon url: %s is not supported", addon)
}

func (u *updater) updateAddon(addon string, source UpdateSource) error {
	log.Printf("updating addon: %s\n", addon)

	currentVersion, err := u.getCurrentVersion(addon)
	if err != nil {
		return err
	}

	latestVersion, err := source.GetLatestVersion(addon)
	if err != nil {
		return err
	}

	if currentVersion == latestVersion {
		log.Println("no need for an update")
		return nil
	}

	path, err := source.DownloadAddon(addon)
	if err != nil {
		return err
	}

	_, err = unzip(path, u.config.Path)
	if err != nil {
		return err
	}

	u.setCurrentVersion(addon, latestVersion)
	log.Printf("updated to version: %s\n", latestVersion)
	return nil
}

// from https://golangcode.com/unzip-files-in-go/
func unzip(src string, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func readVersionsFile(path string) (AddonVersions, error) {
	versions := AddonVersions{
		Addons: make([]Addon, 0),
	}

	if !fileExists(path) {
		return versions, nil
	}

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return versions, err
	}

	err = yaml.Unmarshal(yamlFile, &versions)

	return versions, err
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}
