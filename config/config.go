package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Classic WowConfig `yaml:"classic"`
	Retail  WowConfig `yaml:"retail"`
}

type WowConfig struct {
	Path   string   `yaml:"path"`
	AddOns []string `yaml:"addons"`
}

func ReadConfig(path string) (Config, error) {
	var c Config
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(yamlFile, &c)

	return c, err
}
