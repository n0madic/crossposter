package crossposter

import (
	"time"
)

type (

	// Post data struct
	Post struct {
		Date        time.Time
		URL         string
		Author      string
		Title       string
		Text        string
		Attachments []string
		More        bool
	}

	// Entity type
	Entity struct {
		Type        string `json:"type" yaml:"type"`
		Key         string `json:"key" yaml:"key"`
		KeySecret   string `json:"key_secret" yaml:"key_secret"`
		Token       string `json:"token" yaml:"token"`
		TokenSecret string `json:"token_secret" yaml:"token_secret"`
	}

	// EntityInterface is interface
	EntityInterface interface {
		Get(name string) ([]Post, error)
		Post(name string, post *Post) (string, error)
	}

	// Initializer of entity
	Initializer func(entity Entity) (EntityInterface, error)
)

var (
	// Initializers of entities
	Initializers = make(map[string]Initializer)
)

// AddEntity add initializer
func AddEntity(name string, init Initializer) {
	_, exists := Initializers[name]
	if !exists {
		Initializers[name] = init
	}
}
