package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config contains the separated configurations for WoW retail and classic.
type Config struct {
	Classic WowConfig `yaml:"classic"`
	Retail  WowConfig `yaml:"retail"`
}

// WowConfig contains the path of the interface directory where to write files to.
// The list of addons should be a list of supported URLs.
type WowConfig struct {
	// path to the respective interface directory of the installation
	Path string `yaml:"path"`
	// list of addon URLs to update
	AddOns []string `yaml:"addons"`
}

// ReadConfig reads in the configuration from the given path.
// The content is expected to be YAML.
// Returns an error if not existing.
func ReadConfig(path string) (Config, error) {
	var c Config
	in, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(in, &c)

	return c, err
}

// CreateDefaultConfig writes an empty config in YAML to the given path.
func CreateDefaultConfig(path string) error {
	var c Config
	out, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, os.FileMode(0666))
}
