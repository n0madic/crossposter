package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/config"
	_ "github.com/n0madic/crossposter/entities/twitter"
	_ "github.com/n0madic/crossposter/entities/vk"
)

const timeLayout = "2006-01-02T15:04:05"

var (
	configJSON      string
	defaultWaitTime int64
	dontPost        bool
	lastUpdate      time.Time
	lastUpdateStr   string
	wg              sync.WaitGroup
)

func init() {
	flag.StringVar(&configJSON, "config", "config.yaml", "Config file")
	flag.StringVar(&lastUpdateStr, "last", time.Now().Format(timeLayout), "Initial date for update")
	flag.BoolVar(&dontPost, "dontpost", false, "Do not post on targets")
	flag.Int64Var(&defaultWaitTime, "waittime", 10, "Default wait time duration")
}

func main() {
	flag.Parse()
	cfg, err := config.New(configJSON)
	if err != nil {
		log.Fatalln(err)
	}

	lastUpdate, err = time.Parse(timeLayout, lastUpdateStr)
	if err != nil {
		log.Fatalf("Can't parse last update time: %s\n", err)
	}

	entities := make(map[string]crossposter.EntityInterface)

	for entity := range cfg.Entities {
		log.Printf("Create %s entity: %s", cfg.Entities[entity].Type, entity)
		acc, err := crossposter.Initializers[cfg.Entities[entity].Type](cfg.Entities[entity])
		if err != nil {
			log.Fatalf("Can't create entity %s: %s", entity, err)
		}
		entities[entity] = acc
	}

	for source := range cfg.Sources {
		if _, ok := cfg.Entities[cfg.Sources[source].Entity]; !ok {
			log.Fatalf("Not found entity '%s' for source '%s'", cfg.Sources[source].Entity, source)
		}
		wg.Add(1)
		go func(source string) {
			defer wg.Done()
			entityType := cfg.Entities[cfg.Sources[source].Entity].Type
			LastUpdate := lastUpdate

			waitTime := defaultWaitTime
			if cfg.Sources[source].Waiting != 0 {
				waitTime = cfg.Sources[source].Waiting
			}

			for {
				log.Printf("Get post for [%s] %s", entityType, source)
				posts, err := entities[cfg.Sources[source].Entity].Get(source)
				if err != nil {
					log.Printf("Get post error for [%s] %s", entityType, source)
				}

				sort.Slice(posts, func(i, j int) bool {
					return posts[i].Date.Before(posts[j].Date)
				})

				for _, post := range posts {
					if post.Date.After(LastUpdate) {
						for _, target := range cfg.Sources[source].Destinations {
							if targetEntity, ok := entities[target]; ok {
								logMessage := fmt.Sprintf("Post from [%s] %s to [%s] %s", entityType, source, cfg.Entities[target].Type, target)
								if !dontPost {
									msg, err := targetEntity.Post(target, &post)
									if err != nil {
										log.Printf("%s error: %s", logMessage, err)
									} else {
										LastUpdate = post.Date
										log.Printf("%s: %s", logMessage, msg)
									}
								} else {
									log.Printf("%s skipped!", logMessage)
								}
							} else {
								log.Fatalf("Not found target entity '%s' for source '%s'", target, source)
							}
						}
					}
				}

				time.Sleep(time.Duration(waitTime) * time.Minute)
			}
		}(source)
	}
	wg.Wait()
}
