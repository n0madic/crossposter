package crossposter

import (
	"net/http"
	"time"
)

type (

	// Entity type
	Entity struct {
		Type         string            `json:"type" yaml:"type"`
		Description  string            `json:"description" yaml:"description"`
		Options      map[string]string `json:"options" yaml:"options"`
		Sources      []string          `json:"sources" yaml:"sources"`
		Destinations []string          `json:"destinations" yaml:"destinations"`
		Topics       []string          `json:"topics" yaml:"topics"`
		Wait         int               `json:"wait" yaml:"wait"`
	}

	// EntityInterface is interface
	EntityInterface interface {
		Get(name string, lastUpdate time.Time)
		Post(post Post)
		Handler(w http.ResponseWriter, r *http.Request)
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
