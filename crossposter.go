package crossposter

import (
	"net/http"
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
		Type        string            `json:"type" yaml:"type"`
		Description string            `json:"description" yaml:"description"`
		Options     map[string]string `json:"options" yaml:"options"`
	}

	// EntityInterface is interface
	EntityInterface interface {
		Get(name string) ([]Post, error)
		Post(name string, post *Post) (string, error)
		Handler(w http.ResponseWriter, r *http.Request)
	}

	// Initializer of entity
	Initializer func(name string, entity Entity) (EntityInterface, error)
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
