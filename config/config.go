package config

import (
	"io/ioutil"

	"github.com/n0madic/crossposter"
	yaml "gopkg.in/yaml.v2"
)

// Config struct
type Config struct {
	Consumers []crossposter.Entity
	Producers []crossposter.Entity
	filename  string
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
