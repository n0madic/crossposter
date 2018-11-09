package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/n0madic/crossposter"
)

// Config struct
type Config struct {
	Entities map[string]crossposter.Entity
	Sources  map[string]struct {
		Description  string   `yaml:"description"`
		Entity       string   `yaml:"entity"`
		Waiting      int64    `yaml:"waiting"`
		Destinations []string `yaml:"destinations"`
	} `yaml:"sources"`
	filename string
}

// New config create
func New(filename string) (Config, error) {
	config := Config{}
	config.filename = filename
	err := config.LoadConfig()
	return config, err
}

// LoadConfig load config
func (c *Config) LoadConfig() error {
	configfile, err := ioutil.ReadFile(c.filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(configfile, &c)
	if err != nil {
		return err
	}
	return nil
}
