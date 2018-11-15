package main

import (
	"bytes"
	"flag"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/config"
	_ "github.com/n0madic/crossposter/entities"
	log "github.com/sirupsen/logrus"
)

const timeLayout = "2006-01-02T15:04:05"

var (
	bindHost      string
	configYAML    string
	dontPost      bool
	lastUpdate    time.Time
	lastUpdateStr string
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	})
	flag.StringVar(&configYAML, "config", "config.yaml", "Config file")
	flag.StringVar(&lastUpdateStr, "last", time.Now().Format(timeLayout), "Initial date for update")
	flag.BoolVar(&dontPost, "dontpost", false, "Do not produce posts")
	flag.StringVar(&bindHost, "bind", ":8000", "Bind address")
}

func main() {
	flag.Parse()
	cfg, err := config.New(configYAML)
	if err != nil {
		log.Fatalln(err)
	}

	lastUpdate, err = time.Parse(timeLayout, lastUpdateStr)
	if err != nil {
		log.Fatalf("Can't parse last update time: %s\n", err)
	}

	for _, entity := range cfg.Entities {
		if dontPost {
			entity.Topics = []string{}
		}
		newEntity, err := crossposter.Initializers[entity.Type](entity)
		if err != nil {
			log.Fatalln(err)
		}
		switch entity.Role {
		case "producer":
			for _, source := range entity.Sources {
				crossposter.WaitGroup.Add(1)
				go newEntity.Get(source, lastUpdate)
			}
		case "consumer":
			for _, topic := range entity.Topics {
				crossposter.Events.SubscribeAsync(topic, newEntity.Post, true)
			}
		default:
			log.Fatalf("%s role is not supported", entity.Role)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.New("index").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(indexTpl)
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

	crossposter.WaitGroup.Wait()
}
