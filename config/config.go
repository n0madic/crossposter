package config

import (
	"io/ioutil"
	"time"

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

// SubscribeConsumers on topics
func (c *Config) SubscribeConsumers(dontPost bool) error {
	for _, consumer := range c.Consumers {
		newConsumer, err := crossposter.Initializers[consumer.Type](consumer)
		if err != nil {
			return err
		}
		if !dontPost {
			for _, topic := range consumer.Topics {
				crossposter.Events.SubscribeAsync(topic, newConsumer.Post, true)
			}
		}
	}
	return nil
}

// RunProducers in goroutines
func (c *Config) RunProducers(lastUpdate time.Time) error {
	for _, producer := range c.Producers {
		if producer.Wait == 0 {
			producer.Wait = 5
		}
		newProducer, err := crossposter.Initializers[producer.Type](producer)
		if err != nil {
			return err
		}
		crossposter.WaitGroup.Add(1)
		go newProducer.Get(lastUpdate)
	}
	return nil
}
