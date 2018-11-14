package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/config"
	_ "github.com/n0madic/crossposter/entities"
)

const timeLayout = "2006-01-02T15:04:05"

var (
	bindHost        string
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
	flag.Int64Var(&defaultWaitTime, "waittime", 5, "Default wait time duration in minutes")
	flag.StringVar(&bindHost, "bind", ":8000", "Bind address")
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

	for entity, options := range cfg.Entities {
		log.Printf("Create %s entity: %s", options.Type, entity)
		ent, err := crossposter.Initializers[options.Type](entity, options)
		if err != nil {
			log.Fatalf("Can't create entity %s: %s", entity, err)
		}
		entities[entity] = ent
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.New("index").Parse(indexTpl)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			var tpl bytes.Buffer
			err = t.Execute(&tpl, cfg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			} else {
				w.Write(tpl.Bytes())
			}
		}
	})

	go func() {
		log.Fatal(http.ListenAndServe(bindHost, nil))
	}()

	for source := range cfg.Sources {
		if _, ok := cfg.Entities[cfg.Sources[source].Entity]; !ok {
			log.Fatalf("Not found entity '%s' for source '%s'", cfg.Sources[source].Entity, source)
		}
		for _, target := range cfg.Sources[source].Destinations {
			if _, ok := entities[target]; !ok {
				log.Fatalf("Not found target entity '%s' for source '%s'", target, source)
			}
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
				log.Printf("Check updates for [%s] %s", entityType, source)
				posts, err := entities[cfg.Sources[source].Entity].Get(source)
				if err != nil {
					log.Printf("Get post error for [%s] %s: %s", entityType, source, err)
				}

				sort.Slice(posts, func(i, j int) bool {
					return posts[i].Date.Before(posts[j].Date)
				})

				for _, post := range posts {
					if post.Date.After(LastUpdate) {
						for _, target := range cfg.Sources[source].Destinations {
							logMessage := fmt.Sprintf("Post from [%s] %s to [%s] %s", entityType, source, cfg.Entities[target].Type, target)
							if !dontPost {
								msg, err := entities[target].Post(target, &post)
								if err != nil {
									log.Printf("%s error: %s", logMessage, err)
								} else {
									LastUpdate = post.Date
									log.Printf("%s: %s", logMessage, msg)
								}
							} else {
								log.Printf("%s skipped!", logMessage)
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
