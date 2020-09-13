package crossposter

import (
	"sync"

	"github.com/asaskevich/EventBus"
)

var (
	// Events bus
	Events = EventBus.New()

	// WaitGroup global
	WaitGroup sync.WaitGroup
)
