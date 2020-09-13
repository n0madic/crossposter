package test

import (
	"net/http"
	"time"

	"github.com/n0madic/crossposter"
	log "github.com/sirupsen/logrus"
)

// Test entity
type Test struct {
	entity *crossposter.Entity
	post   crossposter.Post
}

func init() {
	crossposter.AddEntity("test", New)
}

// New return test entity
func New(entity crossposter.Entity) (crossposter.EntityInterface, error) {
	test := &Test{
		entity: &entity,
		post: crossposter.Post{
			Date:   time.Now(),
			URL:    entity.Options["url"],
			Title:  entity.Options["title"],
			Author: entity.Options["author"],
			Text:   entity.Options["text"],
		},
	}
	if entity.Options["attachment"] != "" {
		test.post.Attachments = []string{entity.Options["attachment"]}
	}
	return test, nil
}

// Get test message
func (test *Test) Get(name string, lastUpdate time.Time) {
	defer crossposter.WaitGroup.Done()

	testLogger := log.WithFields(log.Fields{"name": name, "type": test.entity.Type})

	for {
		testLogger.Info("Check test message")
		for _, topic := range test.entity.Topics {
			crossposter.Events.Publish(topic, test.post)
		}
		time.Sleep(time.Duration(test.entity.Wait) * time.Minute)
	}
}

// Post test message
func (test *Test) Post(post crossposter.Post) {
	log.WithFields(log.Fields{
		"title":       post.Title,
		"author":      post.Author,
		"date":        post.Date,
		"url":         post.URL,
		"text":        post.Text,
		"more":        post.More,
		"attachments": post.Attachments,
	}).Info("Test message")
}

// Handler test message
func (test *Test) Handler(w http.ResponseWriter, r *http.Request) {}
