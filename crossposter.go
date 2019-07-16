package crossposter

import (
	"net/http"
	"sync"
	"time"

	"github.com/asaskevich/EventBus"
)

// WaitTime default wait time
const WaitTime = 5

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
		Type         string            `json:"type" yaml:"type"`
		Description  string            `json:"description" yaml:"description"`
		Options      map[string]string `json:"options" yaml:"options"`
		Sources      []string          `json:"sources" yaml:"sources"`
		Destinations []string          `json:"destinations" yaml:"destinations"`
		Topics       []string          `json:"topics" yaml:"topics"`
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

	// Events bus
	Events = EventBus.New()

	// WaitGroup global
	WaitGroup sync.WaitGroup
)

// AddEntity add initializer
func AddEntity(name string, init Initializer) {
	_, exists := Initializers[name]
	if !exists {
		Initializers[name] = init
	}
}
